package components

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type MigrateDB struct {
	migrationDirPath string
}

func (c *MigrateDB) Start() error {
	c.migrationDirPath = "../e2e/migrations"

	var err error
	if c.migrationDirPath, err = filepath.Abs(c.migrationDirPath); err != nil {
		return fmt.Errorf("failed to get absolute path for %v path", c.migrationDirPath)
	}
	if _, err := os.Stat(c.migrationDirPath); os.IsNotExist(err) {
		return fmt.Errorf("migraitn path: %v doesn't exist\n", c.migrationDirPath)
	}
	cmd := []string{
		"docker", "run", "--rm", "-d",
		"-v", fmt.Sprintf("%s:/migrations", c.migrationDirPath),
		"--network=mtf_net",
		"migrate/migrate",
		"-path", "/migrations",
		"-database", "mysql://root:test@tcp(mysql_mtf:3306)/test_db", "up",
	}
	return runCmd(cmd)
}

func (c *MigrateDB) Stop() error {
	return nil
}

func (c *MigrateDB) Ready() error {
	return nil
}

func networkExists(name string) bool {
	return cmdExitStatus(fmt.Sprintf("docker inspect %s", name))
}

func containerIsRunning(name string) bool {
	return cmdExitStatus(fmt.Sprintf("docker top %s", name))
}

func cmdExitStatus(command string) bool {
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:len(args)]...)
	err := cmd.Start()
	if err != nil {
		return false
	}
	err = cmd.Wait()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode() == 0
		}
	}
	return true
}
