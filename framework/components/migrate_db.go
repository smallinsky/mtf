package components

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type MigrateDB struct {
	migrationDirPath string
	start            time.Time
}

func (c *MigrateDB) Start() {
	c.migrationDirPath = "../e2e/migrations"
	c.start = time.Now()

	var err error
	if c.migrationDirPath, err = filepath.Abs(c.migrationDirPath); err != nil {
		log.Printf("[ERROR]: Failed to get absolute path for %v path", c.migrationDirPath)
	}
	if _, err := os.Stat(c.migrationDirPath); os.IsNotExist(err) {
		log.Printf("[ERROR]: Migraitn path: %v doesn't exist\n", c.migrationDirPath)
		return
	}
	cmd := fmt.Sprintf(`docker run --rm -v %s:/migrations --network=mtf_net migrate/migrate -path /migrations -database mysql://root:test@tcp(mysql_mtf:3306)/test_db up`, c.migrationDirPath)
	run(cmd)
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

func (c *MigrateDB) Stop() {
}

func (c *MigrateDB) Ready() {
	fmt.Printf("%T start time %v\n", c, time.Now().Sub(c.start))
}
