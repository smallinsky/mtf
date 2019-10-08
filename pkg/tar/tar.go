package tar

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DirReader tar input dir and returns the io.Reader to dir content.
func DirReader(dir string) (io.Reader, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("failed to stat file: %v", err)
	}
	tw := tar.NewWriter(&buff)
	defer tw.Close()

	if err := filepath.Walk(dir, tarWalkFn(tw)); err != nil {
		return nil, err
	}

	return &buff, nil
}

func tarWalkFn(w *tar.Writer) filepath.WalkFunc {
	return func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return fmt.Errorf("failed to get file header: %v", err)
		}
		if err := w.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write file header: %v", err)
		}
		if fi.IsDir() {
			return nil
		}
		data, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed read file '%v': %v", file, err)
		}
		if _, err := io.Copy(w, data); err != nil {
			return fmt.Errorf("failed to copy file content to buff: %v", err)
		}
		return nil
	}
}
