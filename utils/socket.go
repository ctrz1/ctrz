package utils

import (
	"context"
	"ctrz/spec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

const (
	maxMessageSize  = 1 << 20 // 1 MiB
	maxDialAttempts = 100
	initialBackoff  = 10 * time.Millisecond
	maxDialBackoff  = 500 * time.Millisecond
)

type SocketListener struct {
	l *net.UnixListener
}

func NewSocketListener(path string) (*SocketListener, error) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("Error removing stale socket: %w", err)
	}
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, fmt.Errorf("Error listening on %s: %w", path, err)
	}
	return &SocketListener{l: l}, nil
}

func (s *SocketListener) Close() error { 
	return s.l.Close() 
}

func (s *SocketListener) ListenToSocket(ctx context.Context) (cont spec.ContainerSocket, err error) {
	_ = s.l.SetDeadline(time.Time{})

	stop := context.AfterFunc(ctx, func() {
		_ = s.l.SetDeadline(time.Now())
	})
	c, err := s.l.Accept()
	stop()
	if err != nil {
		return cont, wrap(ctx, "accept", err)
	}
	defer func() {
		err = errors.Join(err, c.Close())
	}()

	stopConn := context.AfterFunc(ctx, func() { 
		_ = c.SetDeadline(time.Now()) 
	})
	defer stopConn()

	dec := json.NewDecoder(io.LimitReader(c, maxMessageSize))
	if err := dec.Decode(&cont); err != nil {
		return spec.ContainerSocket{}, wrap(ctx, "decode", err)
	}
	return cont, nil
}

func WriteToSocket(ctx context.Context, path string, cont spec.ContainerSocket) (err error) {
	c, err := dialWithRetry(ctx, path)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, c.Close())
	}()

	if dl, ok := ctx.Deadline(); ok {
		_ = c.SetDeadline(dl)
	}
	if err := json.NewEncoder(c).Encode(cont); err != nil {
		return fmt.Errorf("Error sending: %w", err)
	}
	return nil
}

func dialWithRetry(ctx context.Context, path string) (net.Conn, error) {
	var d net.Dialer
	backoff := initialBackoff

	for attempt := 1; attempt <= maxDialAttempts; attempt++ {
		c, err := d.DialContext(ctx, "unix", path)
		if err == nil {
			return c, nil
		}
		if !errors.Is(err, syscall.ENOENT) && !errors.Is(err, syscall.ECONNREFUSED) {
			return nil, fmt.Errorf("Error dialing %s: %w", path, err)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Error dialing %s: gave up waiting: %w", path, errors.Join(ctx.Err(), err))
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxDialBackoff)
	}
	return nil, fmt.Errorf("Error connecting to %s after %d tries\n", path, maxDialAttempts)
}

func wrap(ctx context.Context, op string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf("%s: %w", op, ctxErr)
	}
	return fmt.Errorf("%s: %w", op, err)
}
