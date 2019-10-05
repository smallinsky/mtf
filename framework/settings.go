package framework

type Settings struct {
	mysql  *MysqlSettings
	sut    *SutSettings
	pubsub *PubSubSettings
	redis  *RedisSettings
	ftp    *FTPSettings
}

type MysqlSettings struct {
	DatabaseName string
	MigrationDir string
	Password     string
	Port         string
}

func (s *Suite) WithMySQL(c MysqlSettings) *Suite {
	s.settings.mysql = &c
	return s
}

type SutSettings struct {
	Envs []string
	Dir  string
}

func (s *Suite) WithSut(c SutSettings) *Suite {
	s.settings.sut = &c
	return s
}

type PubSubSettings struct{}

func (s *Suite) WithPubSub(c PubSubSettings) *Suite {
	s.settings.pubsub = &c
	return s
}

type RedisSettings struct {
	Port     string
	Password string
}

func (s *Suite) WithRedis(c RedisSettings) *Suite {
	s.settings.redis = &c
	return s
}

type FTPSettings struct {
	Addr string
	User string
	Pass string
}

func (s *Suite) WithFTP(c FTPSettings) *Suite {
	s.settings.ftp = &c
	return s
}

type Comper interface {
	Start() error
	Stop() error
	Ready() error
	StartPriority() int
}
