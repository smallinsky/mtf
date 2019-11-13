package framework

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"testing"
	"time"

	"github.com/smallinsky/mtf/framework/component"
	"github.com/smallinsky/mtf/framework/component/ftp"
	"github.com/smallinsky/mtf/framework/component/migrate"
	"github.com/smallinsky/mtf/framework/component/mysql"
	"github.com/smallinsky/mtf/framework/component/pubsub"
	"github.com/smallinsky/mtf/framework/component/redis"
	"github.com/smallinsky/mtf/framework/component/sut"
	"github.com/smallinsky/mtf/framework/core"
	"github.com/smallinsky/mtf/pkg/cert"
	"github.com/smallinsky/mtf/pkg/docker"
)

var (
	GetDockerHostAddr = docker.HostAddr
	testenv           *TestEnviorment
)

type TestEnviorment struct {
	MySQL  *MysqlSettings
	SUT    *SutSettings
	PubSub *PubSubSettings
	Redis  *RedisSettings
	FTP    *FTPSettings
	TLS    *TLSSettings

	components []component.Component
	network    *docker.Network

	M *testing.M
}

func TestEnv(m *testing.M) *TestEnviorment {
	flag.Parse()

	testenv = &TestEnviorment{
		M: m,
	}
	return testenv
}

func (env *TestEnviorment) Run() int {
	if err := env.Start(); err != nil {
		log.Fatalf("failed to prepare testing environment %v", err)
	}
	defer env.Stop()

	code := env.M.Run()

	if core.Settings.Wait {
		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt)
		fmt.Println("waiting for signal...")
		<-sig
	}

	return code
}

func (env *TestEnviorment) Start() error {
	fmt.Println("=== PREPERING TEST ENV")
	start := time.Now()
	if err := env.genCerts(); err != nil {
		panic(err)
	}

	cli, err := docker.New()
	if err != nil {
		panic(err)
	}
	env.network, err = cli.CreateNetwork("mtf_net")
	if err != nil {
		log.Fatalf("faield to get docker client: %v", err)
	}

	if err := env.Prepare(); err != nil {
		panic(err)
	}

	for _, container := range env.components {
		start := time.Now()
		fmt.Printf("  - Starting %s ", getComponentName(container))
		err := container.Start()
		if err != nil {
			log.Fatalf("\nstart err: %v", err)
		}
		fmt.Printf("-  %v\n", time.Now().Sub(start))
	}
	fmt.Printf("=== TEST RUN DONE - %v\n\n", time.Now().Sub(start))

	return nil
}

func getComponentName(c component.Component) string {
	name := fmt.Sprintf("%T", c)
	name = strings.ReplaceAll(name, "*", "")
	ss := strings.Split(name, ".")
	if len(ss) != 2 {
		return name
	}
	return fmt.Sprintf("[%s %s]", strings.ToUpper(ss[0]), ss[1])
}

func (env *TestEnviorment) Stop() error {
	defer env.network.Remove()

	for _, container := range env.components {
		err := container.Stop()
		if err != nil {
			log.Fatalf("stop err: %v", err)
		}
	}

	return nil
}

func (env *TestEnviorment) Prepare() error {
	cli, err := docker.New()
	if err != nil {
		panic(err)
	}
	var components []component.Component

	if env.Redis != nil {
		comp, err := redis.New(cli, redis.RedisConfig{
			Password: env.Redis.Password,
			Port:     env.Redis.Port,
		})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if env.PubSub != nil {
		cfg := pubsub.Config{
			ProjectID: env.PubSub.ProjectID,
		}
		for _, v := range env.PubSub.TopicSubscriptions {
			cfg.TopicSubscriptions = append(cfg.TopicSubscriptions, pubsub.TopicSubscriptions{
				Topic:         v.Topic,
				Subscriptions: v.Subscriptions,
			})
		}
		comp, err := pubsub.New(cli, cfg)
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if cfg := env.MySQL; cfg != nil {
		comp, err := mysql.New(cli, mysql.MySQLConfig{
			Database:      cfg.DatabaseName,
			Password:      cfg.Password,
			AttachIfExist: true,
		})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if cfg := env.MySQL; cfg != nil && cfg.MigrationDir != "" {
		comp, err := migrate.New(cli, migrate.MigrateConfig{
			Path:     cfg.MigrationDir,
			Password: cfg.Password,
			Port:     "3306",
			Hostname: "mysql_mtf",
			Database: cfg.DatabaseName,
		})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if cfg := env.FTP; cfg != nil {
		comp, err := ftp.New(cli, ftp.FTPConfig{})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if env.SUT != nil {
		env.SUT.Envs = append(env.SUT.Envs, "PUBSUB_EMULATOR_HOST="+GetDockerHostAddr(8085))
		comp, err := sut.New(cli, sut.SutConfig{
			Path:         env.SUT.Dir,
			Env:          env.SUT.Envs,
			ExposedPorts: env.SUT.Ports,
		})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	env.components = components
	return nil
}

func (env *TestEnviorment) genCerts() error {
	var hosts []string
	if env.TLS != nil {
		hosts = env.TLS.Hosts
	}
	_, err := cert.GenCert(hosts)
	return err
}

func (env *TestEnviorment) WithMySQL(settings MysqlSettings) *TestEnviorment {
	env.MySQL = &settings
	return env
}

func (env *TestEnviorment) WithSUT(settings SutSettings) *TestEnviorment {
	env.SUT = &settings
	return env
}

func (env *TestEnviorment) WithPubSub(settings PubSubSettings) *TestEnviorment {
	env.PubSub = &settings
	return env
}

func (env *TestEnviorment) WithRedis(settings RedisSettings) *TestEnviorment {
	env.Redis = &settings
	return env
}

func (env *TestEnviorment) WithFTP(settings FTPSettings) *TestEnviorment {
	env.FTP = &settings
	return env
}

func (env *TestEnviorment) WithTLS(settings TLSSettings) *TestEnviorment {
	env.TLS = &settings
	return env
}
