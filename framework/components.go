package framework

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/smallinsky/mtf/framework/components/ftp"
	"github.com/smallinsky/mtf/framework/components/migrate"
	"github.com/smallinsky/mtf/framework/components/mysql"
	"github.com/smallinsky/mtf/framework/components/network"
	"github.com/smallinsky/mtf/framework/components/pubsub"
	"github.com/smallinsky/mtf/framework/components/redis"
	"github.com/smallinsky/mtf/framework/components/sut"
	"github.com/smallinsky/mtf/pkg/docker"
)

func (s *Suite) getComponents() Components {
	cli, err := docker.NewClient()
	if err != nil {
		log.Fatalf("faield to get docker client: %v", err)
	}

	components := Components{
		m: make(map[startPriority][]Comper),
	}

	components.Add(network.New(cli, network.NetworkConfig{
		Name:          "mtf_net",
		AttachIfExist: true,
	}))

	if s.settings.redis != nil {
		components.Add(redis.NewRedis(cli, redis.RedisConfig{
			Password: s.settings.redis.Password,
			Port:     s.settings.redis.Port,
		}))
	}

	if s.settings.pubsub != nil {
		components.Add(pubsub.NewPubsub(cli))
	}

	if s.settings.sut != nil {
		s.settings.sut.Envs = append(s.settings.sut.Envs, "PUBSUB_EMULATOR_HOST=host.docker.internal:8085")
		components.Add(sut.NewSUT(cli, sut.SutConfig{
			Path: s.settings.sut.Dir,
			Env:  s.settings.sut.Envs,
			//IsCmd: s.settings.sut.RunForEachTest,
		}))
	}

	if cfg := s.settings.mysql; cfg != nil {
		components.Add(mysql.NewMySQL(cli, mysql.MySQLConfig{
			Database:      cfg.DatabaseName,
			Password:      cfg.Password,
			AttachIfExist: true,
		}))
	}

	if cfg := s.settings.mysql; cfg != nil && cfg.MigrationDir != "" {
		components.Add(migrate.NewMigrate(cli, migrate.MigrateConfig{
			Path:     cfg.MigrationDir,
			Password: cfg.Password,
			Port:     "3306",
			Hostname: "mysql_mtf",
			Database: cfg.DatabaseName,
		}))
	}

	if cfg := s.settings.ftp; cfg != nil {
		components.Add(ftp.NewFTP(cli))
	}

	return components
}

type startPriority int

type Components struct {
	m       map[startPriority][]Comper
	maxPrio int
}

func StartWithSamePrio(cc []Comper) {
	var wg sync.WaitGroup
	for _, c := range cc {
		wg.Add(1)
		go func(c Comper) {
			defer wg.Done()
			fmt.Printf("%T is starting \n", c)
			start := time.Now()
			if err := c.Start(); err != nil {
				log.Fatalf("failed to start %T, err %v\n", c, err)
			}
			if err := c.Ready(); err != nil {
				log.Fatalf("faield to call ready for %T, err: %v", c, err)
			}
			fmt.Printf("%T is ready - %v\n", c, time.Since(start))
		}(c)
	}
	wg.Wait()
}

func StopWithSamePrio(cc []Comper) {
	var wg sync.WaitGroup
	for _, c := range cc {
		if err := WriteLogs(c); err != nil {
			log.Panicf("Failed to write logs: %v", err)
		}
		wg.Add(1)
		go func(cp Comper) {
			defer wg.Done()
			if err := cp.Stop(); err != nil {
				fmt.Printf("failed to stop %T: %v", cp, err)
			}
		}(c)
	}
	wg.Wait()
}

func (c *Components) Add(v Comper) {
	if v.StartPriority() > c.maxPrio {
		c.maxPrio = v.StartPriority()
	}
	c.m[startPriority(v.StartPriority())] = append(c.m[startPriority(v.StartPriority())], v)
}

func (c *Components) Start() {
	for i := 0; i < c.maxPrio+1; i++ {
		cc, ok := c.m[startPriority(i)]
		if !ok {
			continue
		}
		StartWithSamePrio(cc)
	}
}

func (co *Components) Stop() {
	for i := co.maxPrio; i >= 0; i-- {
		cc, ok := co.m[startPriority(i)]
		if !ok {
			continue
		}
		StopWithSamePrio(cc)
	}
}

func WriteLogs(c Comper) error {
	sutC, ok := c.(*sut.SUT)
	if !ok {
		return nil
	}
	buff, err := sutC.Logs()
	if err != nil {
		return fmt.Errorf("got error during sut logs call: %v", err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/sut.logs", "runlogs"), buff, 0644)
	if err != nil {
		return fmt.Errorf("failed to write sut logs: %v", err)
	}
	return nil
}
