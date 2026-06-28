//go:build linux
// +build linux

package fs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/rootfs.tar.gz
var rootfsTar []byte

func MountRootFs(containerID string) (string, error) {
	rootfsDir := filepath.Join(
		"/var/lib/ctrz/containers",
		containerID,
	)

	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return "", err
	}

	if err := extractTarGz(rootfsTar, rootfsDir); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/rootfs", rootfsDir), nil
}

func extractTarGz(data []byte, dest string) error {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Error creating gzip reader: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Error reading tar entry: %v", err)
		}

		target := filepath.Join(dest, hdr.Name)

		cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
		cleanTarget := filepath.Clean(target)

		if !strings.HasPrefix(cleanTarget, cleanDest) {
			return fmt.Errorf("invalid path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanTarget, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("create dir %s: %w", cleanTarget, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
				return fmt.Errorf("Error creating parent dir: %v", err)
			}
			f, err := os.OpenFile(
				cleanTarget,
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
				os.FileMode(hdr.Mode),
			)
			if err != nil {
				return fmt.Errorf("Error creating file %s: %v", cleanTarget, err)
			}
			_, copyErr := io.Copy(f, tr)
			closeErr := f.Close()

			if copyErr != nil {
				return fmt.Errorf("Error writing file %s: %v", cleanTarget, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("Error closing file %s: %v", cleanTarget, closeErr)
			}

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
				return err
			}

			if err := os.Symlink(/*filepath.Join(dest, hdr.Linkname)*/hdr.Linkname, cleanTarget); err != nil {
				return fmt.Errorf(
					"Error creating symlink %s -> %s: %v",
					cleanTarget,
					hdr.Linkname,
					err,
				)
			}

		default:
			// Ignore things like devices, fifos, hardlinks for now.
		}
	}

	return nil
}
