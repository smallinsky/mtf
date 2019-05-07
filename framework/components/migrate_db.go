package components

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	cmd := fmt.Sprintf(`docker run --rm -v %s:/migrations --network=mtf_net migrate/migrate -path /migrations -database mysql://root:test@tcp(mysql_mtf:3306)/test_db up`, c.migrationDirPath)
	log.Printf("[INFO] %T running '%s'\n", c, cmd)
	run(cmd)
}

func (c *MigrateDB) Stop() {
}

func (c *MigrateDB) Ready() {
}
