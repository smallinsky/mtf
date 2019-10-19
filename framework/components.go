package framework

import (
	"github.com/smallinsky/mtf/framework/components/ftp"
	"github.com/smallinsky/mtf/framework/components/migrate"
	"github.com/smallinsky/mtf/framework/components/mysql"
	"github.com/smallinsky/mtf/framework/components/pubsub"
	"github.com/smallinsky/mtf/framework/components/redis"
	"github.com/smallinsky/mtf/framework/components/sut"
	"github.com/smallinsky/mtf/pkg/docker"
)

func (s *Suite) GetContainersConfig() ([]*docker.ContainerConfig, error) {
	var components []*docker.ContainerConfig

	if s.settings.redis != nil {
		conf, err := redis.BuildContainerConfig(redis.RedisConfig{
			Password: s.settings.redis.Password,
			Port:     s.settings.redis.Port,
		})
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	if s.settings.pubsub != nil {
		conf, err := pubsub.BuildContainerConfig()
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	if cfg := s.settings.mysql; cfg != nil {
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

	if cfg := s.settings.mysql; cfg != nil && cfg.MigrationDir != "" {
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

	if cfg := s.settings.ftp; cfg != nil {
		conf, err := ftp.BuildContainerConfig(ftp.FTPConfig{})
		if err != nil {
			panic(err)
		}
		components = append(components, conf)
	}

	if s.settings.sut != nil {
		s.settings.sut.Envs = append(s.settings.sut.Envs, "PUBSUB_EMULATOR_HOST="+GetDockerHostAddr(8085))
		conf, err := sut.BuildContainerConfig(sut.SutConfig{
			Path:         s.settings.sut.Dir,
			Env:          s.settings.sut.Envs,
			ExposedPorts: s.settings.sut.Ports,
			//IsCmd: s.settings.sut.RunForEachTest,
		})
		if err != nil {
			return nil, err
		}
		components = append(components, conf)
	}

	return components, nil
}
