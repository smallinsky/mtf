package framework

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	testenv           *TestEnvironment
)

type TestEnvironment struct {
	settings Settings

	components []component.Component
	SUT        component.Component
	network    *docker.Network

	M *testing.M
}

func TestEnv(m *testing.M) *TestEnvironment {
	flag.Parse()

	testenv = &TestEnvironment{
		M: m,
	}
	return testenv
}

func (env *TestEnvironment) Run() {
	ctx := context.Background()

	if err := env.Start(ctx); err != nil {
		log.Fatalf("[ERROR] Failed to prepare testing environment %v", err)
	}

	defer func() {
		if err := env.Stop(ctx); err != nil {
			log.Fatalf("[ERROR] Failed to stop containers: %v", err)
		}
	}()

	code := env.M.Run()

	if err := env.WriteLogs(ctx, ""); err != nil {
		log.Fatalf("[ERROR] Failed to write containers logs: %v", err)
	}

	if core.Settings.Wait {
		sig := make(chan os.Signal)
		signal.Notify(sig, os.Interrupt)
		fmt.Println("waiting for signal...")
		<-sig
	}

	os.Exit(code)
}

func (env *TestEnvironment) Start(ctx context.Context) error {
	fmt.Println("=== PREPARING TEST ENV")
	start := time.Now()
	if err := env.genCerts(); err != nil {
		log.Fatalf("[ERROR] Failed to generate tls certs: %v", err)
	}

	cli, err := docker.New()
	if err != nil {
		log.Fatalf("[ERROR] Failed to create docker client: %v", err)
	}
	env.network, err = cli.CreateNetwork("mtf_net")
	if err != nil {
		log.Fatalf("[ERROR] Failed to create docker network: %v", err)
	}

	if err := env.Prepare(cli); err != nil {
		log.Fatalf("[ERROR] Failed to prepare env: %v", err)
	}

	for _, container := range env.components {
		start := time.Now()
		fmt.Printf("  - Starting %s ", getComponentName(container))
		err := container.Start(ctx)
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

func (env *TestEnvironment) Stop(ctx context.Context) error {
	defer env.network.Remove()
	for _, container := range env.components {
		err := container.Stop(ctx)
		if err != nil {
			log.Fatalf("stop err: %v", err)
		}
	}
	return nil
}

func (env *TestEnvironment) StartSutInCommandMode() error {
	err := env.SUT.Start(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (env *TestEnvironment) StopSutInCommandMode(tcName string) error {
	ctx := context.Background()
	if err := env.WriteComponentLogs(ctx, env.SUT, fmt.Sprintf("%s-", tcName)); err != nil {
		log.Printf("[ERROR] Failed to write sut logs: %v", err)
	}
	if err := env.SUT.Stop(ctx); err != nil {
		return err
	}
	return nil
}

func (env *TestEnvironment) WriteLogs(ctx context.Context, tcName string) error {
	if err := os.MkdirAll("runlogs/components", os.ModePerm); err != nil {
		return err
	}
	for _, container := range env.components {
		err := env.WriteComponentLogs(ctx, container, "components/")
		if err != nil {
			return err
		}
	}
	return nil
}

func (env *TestEnvironment) WriteComponentLogs(ctx context.Context, cpnt component.Component, prefix string) error {
	v, ok := cpnt.(component.Loggable)
	if !ok {
		return nil
	}
	r, err := v.Logs(ctx)
	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("runlogs/%s%s.log", prefix, v.Name()))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func (env *TestEnvironment) Prepare(cli *docker.Docker) error {
	var components []component.Component

	conf := env.settings

	if conf.Redis != nil {
		comp, err := redis.New(cli, redis.RedisConfig{
			Password: conf.Redis.Password,
			Port:     conf.Redis.Port,
		})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if conf.PubSub != nil {
		cfg := pubsub.Config{
			ProjectID: conf.PubSub.ProjectID,
		}
		for _, v := range conf.PubSub.TopicSubscriptions {
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

	if cfg := conf.MySQL; cfg != nil {
		comp, err := mysql.New(cli, mysql.MySQLConfig{
			Database:      cfg.DatabaseName,
			Databases:     cfg.Databases,
			Password:      cfg.Password,
			AttachIfExist: true,
		})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if len(conf.Migration) != 0 {
		for _, mig := range conf.Migration {
			comp, err := migrate.New(cli, migrate.MigrateConfig{
				Path:     mig.Dir,
				Password: mig.Password,
				Port:     mig.Port,
				Hostname: "mysql_mtf",
				Database: mig.DBName,
			})
			if err != nil {
				return err
			}
			components = append(components, comp)
		}
	}

	if false {
		if cfg := conf.MySQL; cfg != nil && cfg.MigrationDir != "" {
			dirs := strings.Split(cfg.MigrationDir, ";")
			for _, dir := range dirs {
				comp, err := migrate.New(cli, migrate.MigrateConfig{
					Path:     dir,
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
		}
	}

	if cfg := conf.FTP; cfg != nil {
		comp, err := ftp.New(cli, ftp.FTPConfig{})
		if err != nil {
			return err
		}
		components = append(components, comp)
	}

	if conf.SUT != nil {
		conf.SUT.Envs = append(conf.SUT.Envs, "PUBSUB_EMULATOR_HOST="+GetDockerHostAddr(8085))
		comp, err := sut.New(cli, sut.SutConfig{
			Path:               conf.SUT.Dir,
			Env:                conf.SUT.Envs,
			ExposedPorts:       conf.SUT.Ports,
			RuntimeTypeCommand: conf.SUT.RuntimeType == RuntimeTypeCommand,
		})
		if err != nil {
			return err
		}

		env.SUT = comp
		if conf.SUT.RuntimeType != RuntimeTypeCommand {
			components = append(components, comp)
		}
	}

	env.components = components
	return nil
}

func (env *TestEnvironment) genCerts() error {
	if env.settings.TLS == nil {
		return nil
	}
	_, err := cert.GenCert(env.settings.TLS.Hosts)
	return err
}

func GetTLSCertPath() string {
	return cert.ServerCertFile
}

func GetTLSKeyPath() string {
	return cert.ServerKeyFile
}
