package components

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type MigrateDB struct {
	migrationDirPath string
}

func (c *MigrateDB) Start() {
	c.migrationDirPath = "../e2e/migrations"

	var err error
	if c.migrationDirPath, err = filepath.Abs(c.migrationDirPath); err != nil {
		log.Printf("[ERROR]: Failed to get absolute path for %v path", c.migrationDirPath)
	}
	if _, err := os.Stat(c.migrationDirPath); os.IsNotExist(err) {
		log.Printf("[ERROR]: Migraitn path: %v doesn't exist\n", c.migrationDirPath)
		return
	}
	cmd := []string{
		"docker", "run", "--rm", "-d",
		"-v", fmt.Sprintf("%s:/migrations", c.migrationDirPath),
		"--network=mtf_net",
		"migrate/migrate",
		"-path", "/migrations",
		"-database", "mysql://root:test@tcp(mysql_mtf:3306)/test_db", "up",
	}
	runCmd(cmd)
}

func (c *MigrateDB) Stop() {
}

func (c *MigrateDB) Ready() {
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
