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
}

func (c *MigrateConfig) DBConnString() string {
	return fmt.Sprintf("mysql://root:%s@tcp(%s:%s)/%s", c.Password, c.Hostname, c.Port, c.Database)
}

func BuildContainerConfig(config MigrateConfig) (*docker.ContainerConfig, error) {
	var err error
	if config.Path, err = filepath.Abs(config.Path); err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %v path", config.Path)
	}
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("migraitn path: %v doesn't exist\n", config.Path)
	}

	var (
		image   = "migrate/migrate"
		name    = "migrate_mtf"
		network = "mtf_net"
	)

	return &docker.ContainerConfig{
		Image:       image,
		Name:        name,
		NetworkName: network,
		CapAdd:      []string{"NET_RAW", "NET_ADMIN"},
		Mounts: docker.Mounts{
			docker.Mount{
				Source: config.Path,
				Target: "/migrations",
			},
		},
		Cmd: []string{
			"-path", "/migrations",
			"-database", config.DBConnString(), "up",
		},
	}, nil
}
