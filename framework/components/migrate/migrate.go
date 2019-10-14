package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/smallinsky/mtf/pkg/docker"
)

type MigrateDB struct {
	config    MigrateConfig
	cli       *docker.Docker
	container *docker.ContainerType
}

type MigrateConfig struct {
	Path string

	Password string
	Port     string
	Hostname string
	Database string
	Labels   map[string]string
}

func (c *MigrateConfig) DBConnString() string {
	return fmt.Sprintf("mysql://root:%s@tcp(%s:%s)/%s", c.Password, c.Hostname, c.Port, c.Database)
}

func NewMigrate(cli *docker.Docker, config MigrateConfig) *MigrateDB {
	return &MigrateDB{
		config: config,
		cli:    cli,
	}
}

func (m *MigrateDB) Start() error {
	var err error
	if m.config.Path, err = filepath.Abs(m.config.Path); err != nil {
		return fmt.Errorf("failed to get absolute path for %v path", m.config.Path)
	}
	if _, err := os.Stat(m.config.Path); os.IsNotExist(err) {
		return fmt.Errorf("migraitn path: %v doesn't exist\n", m.config.Path)
	}

	var (
		image    = "migrate/migrate"
		name     = "migrate_mtf"
		hostname = "migrate_mtf"
		network  = "mtf_net"
	)

	result, err := m.cli.NewContainer(docker.ContainerConfig{
		Image:       image,
		Name:        name,
		Hostname:    hostname,
		NetworkName: network,
		CapAdd:      []string{"NET_RAW", "NET_ADMIN"},
		Mounts: docker.Mounts{
			docker.Mount{
				Source: m.config.Path,
				Target: "/migrations",
			},
		},
		Cmd: []string{
			"-path", "/migrations",
			"-database", m.config.DBConnString(), "up",
		},
	})
	if err != nil {
		return err
	}

	m.container = result
	return m.container.Start()
}

func (m *MigrateDB) Stop() error {
	return m.container.Stop()
}

func (m *MigrateDB) Ready() error {
	if m.container == nil {
		return fmt.Errorf("got nil container")
	}
	state, err := m.container.GetState()
	if err != nil {
		return err
	}
	if state.ExitCode != 0 {
		return fmt.Errorf("container has finished with status code %v", state.ExitCode)
	}

	return nil
}

func (m *MigrateDB) StartPriority() int {
	return 3
}
