package framework

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/smallinsky/mtf/framework/component/ftp"
	"github.com/smallinsky/mtf/framework/component/migrate"
	"github.com/smallinsky/mtf/framework/component/mysql"
	"github.com/smallinsky/mtf/framework/component/pubsub"
	"github.com/smallinsky/mtf/framework/component/redis"
	"github.com/smallinsky/mtf/framework/component/sut"
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

	containers []*docker.ContainerType
	network    *docker.Network

	M *testing.M
}

func TestEnv(m *testing.M) *TestEnviorment {
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
	return env.M.Run()
}

func genCerts() {
	kvpair, err := cert.GenCert([]string{"localhost", "host.docker.internal"})
	if err != nil {
		log.Fatalf("[ERR] failed to generate certs: %v", err)
	}

	if err := cert.WriteCert(kvpair); err != nil {
		log.Fatalf("[ERR] failed to write certs: %v ", err)
	}
}

func (env *TestEnviorment) Start() error {
	fmt.Println("=== PREPERING TEST ENV")
	start := time.Now()
	genCerts()

	cli, err := docker.New()
	if err != nil {
		panic(err)
	}
	env.network, err = cli.CreateNetwork("mtf_net")
	if err != nil {
		log.Fatalf("faield to get docker client: %v", err)
	}

	containersConfig, err := env.Prepare()
	if err != nil {
		panic(err)
	}

	for _, conf := range containersConfig {
		container, err := cli.NewContainer(*conf)
		if err != nil {
			log.Fatalf("[ERR] failed to run container: %v", err)
		}
		env.containers = append(env.containers, container)
	}

	for _, container := range env.containers {
		start := time.Now()
		fmt.Printf("  - Starting %s ", container.Name())
		err := container.Start()
		if err != nil {
			log.Fatalf("\nstart err: %v", err)
		}
		fmt.Printf("-  %v\n", time.Now().Sub(start))
	}
	fmt.Printf("=== TEST RUN DONE - %v\n\n", time.Now().Sub(start))

	return nil
}

func (env *TestEnviorment) Stop() error {
	defer env.network.Remove()

	for _, container := range env.containers {
		err := container.Stop()
		if err != nil {
			log.Fatalf("stop err: %v", err)
		}
	}

	return nil
}

func (env *TestEnviorment) Prepare() ([]*docker.ContainerConfig, error) {
	var components []*docker.ContainerConfig

	if env.Redis != nil {
		conf, err := redis.BuildContainerConfig(redis.RedisConfig{
			Password: env.Redis.Password,
			Port:     env.Redis.Port,
		})
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	if env.PubSub != nil {
		conf, err := pubsub.BuildContainerConfig()
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	if cfg := env.MySQL; cfg != nil {
		conf, err := mysql.BuildContainerConfig(mysql.MySQLConfig{
			Database:      cfg.DatabaseName,
			Password:      cfg.Password,
			AttachIfExist: true,
		})
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	if cfg := env.MySQL; cfg != nil && cfg.MigrationDir != "" {
		conf, err := migrate.BuildContainerConfig(migrate.MigrateConfig{
			Path:     cfg.MigrationDir,
			Password: cfg.Password,
			Port:     "3306",
			Hostname: "mysql_mtf",
			Database: cfg.DatabaseName,
		})
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	if cfg := env.FTP; cfg != nil {
		conf, err := ftp.BuildContainerConfig(ftp.FTPConfig{})
		if err != nil {
			panic(err)
		}
		components = append(components, conf)
	}

	if env.SUT != nil {
		env.SUT.Envs = append(env.SUT.Envs, "PUBSUB_EMULATOR_HOST="+GetDockerHostAddr(8085))
		conf, err := sut.BuildContainerConfig(sut.SutConfig{
			Path:         env.SUT.Dir,
			Env:          env.SUT.Envs,
			ExposedPorts: env.SUT.Ports,
		})
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	return components, nil
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
