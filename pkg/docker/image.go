package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/smallinsky/mtf/pkg/tar"
)

type PullImageConfig struct {
	Ref string
}

func (c *Docker) PullImage(ref string) error {
	_, err := c.cli.ImagePull(context.Background(), ref, types.ImagePullOptions{})
	return err
}

type BuildImageConfig struct {
	Path string
	Tag  string
}

func (c *Docker) BuildImage(path, tag string) error {
	r, err := tar.DirReader(path)
	if err != nil {
		return fmt.Errorf("failed to tar dir: %v", err)
	}
	_, err = c.cli.ImageBuild(context.Background(), r, types.ImageBuildOptions{
		Tags: []string{tag},
	})
	return err
}
