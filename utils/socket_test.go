package utils

import (
	"context"
	"ctrz/spec"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
)

func testSockPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "ctrz")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "c.sock")
}

func newListener(t *testing.T) (*SocketListener, string) {
	t.Helper()
	path := testSockPath(t)
	l, err := NewSocketListener(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l, path
}

func timeoutCtx(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

func sendAsync(ctx context.Context, path string, msg spec.ContainerSocket) <-chan error {
	errc := make(chan error, 1)
	go func() { errc <- WriteToSocket(ctx, path, msg) }()
	return errc
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		size int // length of the ID string
	}{
		{"empty", 0},
		{"small", 16},
		{"larger than the old 512 byte buffer", 4 << 10},
		{"large", 600 << 10},
		{"close to the limit", maxMessageSize - 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, path := newListener(t)
			ctx := timeoutCtx(t, 5*time.Second)

			want := spec.ContainerSocket{ID: strings.Repeat("x", tt.size), ExposePorts: true}
			errc := sendAsync(ctx, path, want)

			got, err := l.ListenToSocket(ctx)
			if err != nil {
				t.Fatalf("ListenToSocket: %v", err)
			}
			if err := <-errc; err != nil {
				t.Fatalf("WriteToSocket: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("message corrupted: got ID of len %d (ExposePorts=%v), want len %d (ExposePorts=%v)",
					len(got.ID), got.ExposePorts, len(want.ID), want.ExposePorts)
			}
		})
	}
}

func TestSenderStartsBeforeListener(t *testing.T) {
	path := testSockPath(t)
	ctx := timeoutCtx(t, 5*time.Second)

	errc := sendAsync(ctx, path, spec.ContainerSocket{ID: "late"})

	// The sender is already retrying against a socket that doesn't exist yet.
	time.Sleep(300 * time.Millisecond)
	l, err := NewSocketListener(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })

	got, err := l.ListenToSocket(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "late" {
		t.Fatalf("got %+v", got)
	}
	if err := <-errc; err != nil {
		t.Fatalf("WriteToSocket: %v", err)
	}
}

func TestReceiverTimeout(t *testing.T) {
	l, _ := newListener(t)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := l.ListenToSocket(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want DeadlineExceeded, got %v", err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("took %v to give up", d)
	}
}

func TestReceiverCancel(t *testing.T) {
	l, _ := newListener(t)

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)

	_, err := l.ListenToSocket(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want Canceled, got %v", err)
	}
}

// A receive that timed out must not poison the listener for the next call.
func TestListenerUsableAfterTimeout(t *testing.T) {
	l, path := newListener(t)

	short, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := l.ListenToSocket(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("first call: want DeadlineExceeded, got %v", err)
	}

	ctx := timeoutCtx(t, 5*time.Second)
	errc := sendAsync(ctx, path, spec.ContainerSocket{ID: "second"})

	got, err := l.ListenToSocket(ctx)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if got.ID != "second" {
		t.Fatalf("got %+v", got)
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
}

func TestSenderTimeoutKeepsDialError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := WriteToSocket(ctx, testSockPath(t), spec.ContainerSocket{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want DeadlineExceeded, got %v", err)
	}
	if !errors.Is(err, syscall.ENOENT) {
		t.Fatalf("want the underlying ENOENT preserved, got %v", err)
	}
}

func TestSenderCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)

	err := WriteToSocket(ctx, testSockPath(t), spec.ContainerSocket{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want Canceled, got %v", err)
	}
}

func TestNonRetryableDialErrorFailsFast(t *testing.T) {
	ctx := timeoutCtx(t, 5*time.Second)

	// Longer than sun_path allows: a permanent error, not "listener not up yet".
	path := "/tmp/" + strings.Repeat("a", 200)

	start := time.Now()
	err := WriteToSocket(ctx, path, spec.ContainerSocket{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("kept retrying until the deadline: %v", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("took %v to fail", d)
	}
}

func TestOversizedMessageRejected(t *testing.T) {
	l, path := newListener(t)
	ctx := timeoutCtx(t, 5*time.Second)

	big := spec.ContainerSocket{ID: strings.Repeat("x", 3*maxMessageSize)}
	errc := sendAsync(ctx, path, big)

	_, err := l.ListenToSocket(ctx)
	if err == nil {
		t.Fatal("expected an error for a message over the limit")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("receiver hung instead of rejecting: %v", err)
	}
	// The receiver hung up mid-stream, so the sender must fail rather than hang.
	if err := <-errc; err == nil {
		t.Fatal("sender should have failed once the receiver closed the connection")
	}
}

func TestMalformedMessage(t *testing.T) {
	l, path := newListener(t)
	ctx := timeoutCtx(t, 5*time.Second)

	c, err := net.Dial("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Write([]byte("this is not json{")); err != nil {
		t.Fatal(err)
	}
	_ = c.Close()

	if _, err := l.ListenToSocket(ctx); err == nil {
		t.Fatal("expected a decode error")
	}
}

func TestCloseRemovesSocketFile(t *testing.T) {
	l, path := newListener(t)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("socket should exist while listening: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("socket file should be gone after Close, got %v", err)
	}
}

func TestStaleSocketIsReplaced(t *testing.T) {
	path := testSockPath(t)

	l1, err := NewSocketListener(path)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a crash: the socket file survives the listener.
	l1.l.SetUnlinkOnClose(false)
	_ = l1.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("test setup: stale socket should exist: %v", err)
	}

	l2, err := NewSocketListener(path)
	if err != nil {
		t.Fatalf("stale socket should be replaced: %v", err)
	}
	_ = l2.Close()
}

func TestNewSocketListenerMissingDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist", "c.sock")
	if _, err := NewSocketListener(path); err == nil {
		t.Fatal("expected an error when the directory does not exist")
	}
}