package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/smallinsky/mtf/pkg/docker"
)

type MigrateDB struct {
	migrationDirPath string

	container *docker.Container
}

func (m *MigrateDB) Start() error {
	m.migrationDirPath = "../../e2e/migrations"

	var err error
	if m.migrationDirPath, err = filepath.Abs(m.migrationDirPath); err != nil {
		return fmt.Errorf("failed to get absolute path for %v path", m.migrationDirPath)
	}
	if _, err := os.Stat(m.migrationDirPath); os.IsNotExist(err) {
		return fmt.Errorf("migraitn path: %v doesn't exist\n", m.migrationDirPath)
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	container, err := docker.NewContainer(cli, docker.Config{
		Name:     "mtf_migrate",
		Image:    "migrate/migrate",
		Hostname: "run_sut",
		CapAdd:   []string{"NET_RAW", "NET_ADMIN"},
		Labels: map[string]string{
			"mtf": "mtf",
		},
		Mounts: docker.Mounts{
			docker.Mount{
				Source: m.migrationDirPath,
				Target: "/migrations",
			},
		},
		NetworkName: "mtf_net",
		Cmd: []string{
			"-path", "/migrations",
			"-database", "mysql://root:test@tcp(mysql_mtf:3306)/test_db", "up",
		},
	})
	if err != nil {
		return err
	}

	if err := container.Start(); err != nil {
		fmt.Println("stage 1")
		return err
	}
	m.container = container
	fmt.Println("container asssigned")
	return nil
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
