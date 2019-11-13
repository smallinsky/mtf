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

type SutSettings struct {
	Envs  []string
	Dir   string
	Ports []int

	// RunForEachTest tells that binary should be executed for each test. Example for that
	// behavior is testing command binary instead of servie that always returns after execution.
	RunForEachTest bool
}

type PubSubSettings struct {
	ProjectID          string
	TopicSubscriptions []TopicSubscriptions
}

type TopicSubscriptions struct {
	Topic         string
	Subscriptions []string
}

type RedisSettings struct {
	Port     string
	Password string
}

type FTPSettings struct {
	Addr string
	User string
	Pass string
}

type TLSSettings struct {
	Hosts []string
}
