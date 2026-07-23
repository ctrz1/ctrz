package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func InjectBinary(src, dest string) error {
	zip := filepath.Base(src)
	_, isZip := strings.CutSuffix(zip, ".tar.gz")
	if isZip {
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("Error reading zip archive: %v", err)
		}
		if err := extractTarGz(data, dest); err != nil {
			return err
		}
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(
		filepath.Join(dest, "bin"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		info.Mode(),
	)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}
