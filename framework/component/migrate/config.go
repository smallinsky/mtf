package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/smallinsky/mtf/pkg/docker"
)

type MigrateConfig struct {
	Path     string
	Password string
	Port     string
	Hostname string
	Database string
	Labels   map[string]string

	absolutePath string
}

func (c *MigrateConfig) Build() error {
	stat, err := os.Stat(c.Path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("path is not directory")
	}

	c.absolutePath, err = filepath.Abs(c.Path)
	if err != nil {
		return err
	}
	return nil
}

func (c *MigrateConfig) DBConnString() string {
	return fmt.Sprintf("mysql://root:%s@tcp(%s:%s)/%s", c.Password, c.Hostname, c.Port, c.Database)
}

func BuildContainerConfig(config MigrateConfig) (*docker.ContainerConfig, error) {
	if err := config.Build(); err != nil {
		return nil, err
	}

	var (
		image   = "migrate/migrate"
		name    = "migrate_mtf"
		network = "mtf_net"
	)

	return &docker.ContainerConfig{
		Image:       image,
		Name:        fmt.Sprintf("%s_%s", name, config.Database),
		NetworkName: network,
		CapAdd:      []string{"NET_RAW", "NET_ADMIN"},
		Mounts: docker.Mounts{
			docker.Mount{
				Source: config.absolutePath,
				Target: "/migrations",
			},
		},
		Cmd: []string{
			"-path", "/migrations",
			"-database", config.DBConnString(), "up",
		},
	}, nil
}
