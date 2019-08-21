package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
)

type PullImageConfig struct {
	Ref string
}

func (c *Client) PullImage(ref string) error {
	_, err := c.cli.ImagePull(context.Background(), ref, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	return nil
}

type BuildImageConfig struct {
	Path string
	Tag  string
}

func (c *Client) BuildImage(path, tag string) error {
	r, err := tarDir(path)
	if err != nil {
		return fmt.Errorf("failed to tar dir: %v", err)
	}
	_, err = c.cli.ImageBuild(context.Background(), r, types.ImageBuildOptions{
		Tags: []string{tag},
	})

	if err != nil {
		return err
	}

	return nil
}

func tarDir(path string) (io.Reader, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("failed to stat file: %v", err)
	}
	tw := tar.NewWriter(&buff)
	defer tw.Close()

	err := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return fmt.Errorf("failed to get file header: %v", err)
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write file header: %v", err)
		}
		if fi.IsDir() {
			return nil
		}
		data, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed read file '%v': %v", file, err)
		}
		if _, err := io.Copy(tw, data); err != nil {
			return fmt.Errorf("failed to copy file content to buff: %v", err)
		}
		return nil

	})
	if err != nil {
		return nil, err
	}
	return &buff, nil
}
