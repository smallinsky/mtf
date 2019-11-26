package framework

type Settings struct {
	MySQL  *MysqlSettings
	SUT    *SutSettings
	PubSub *PubSubSettings
	Redis  *RedisSettings
	FTP    *FTPSettings
	TLS    *TLSSettings
}

type MysqlSettings struct {
	// DatabaseName is name of a database that will be created
	DatabaseName string
	// The source dir from migration files.
	MigrationDir string
	// Password for mysql user.
	Password string
	// Port address used for mysql service.
	Port string
}

// RuntimeType distinguish between different ways of sut execution.
type RuntimeType int

const (
	// RuntimeTypeService indicates that system under test works as a service.
	RuntimeTypeService RuntimeType = iota

	// RuntimeTypeCommand indicates that system under test is a simble binary that will
	// terminate after execution.
	RuntimeTypeCommand
)

// SutSettings system under test settings used to build and executed sut in docker container.
type SutSettings struct {
	// Envs allows to pass custom env to system under test container.
	Envs []string
	// Dir is a path to directory that collect system under test source for which binary shuld be build.
	Dir string
	// Ports is a collection of ports that sut binary require, those ports will be forwarded to local host with the
	// same port mapping.
	Ports []int

	// RuntimeType Type of system under test runtime. In case of service runtime sut component will
	// be executed once, but when runtime type is set to command (terminates after execution) sut component
	// needs to be re-executed for each testcase.
	RuntimeType RuntimeType
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

// TLSSettings allows to pass additional that will be used during generation certs.
type TLSSettings struct {
	// Host is a list of DSName or IP that will be added as a Subject Alternative Names.
	Hosts []string
}

func (env *TestEnviorment) WithMySQL(settings MysqlSettings) *TestEnviorment {
	env.settings.MySQL = &settings
	return env
}

func (env *TestEnviorment) WithSUT(settings SutSettings) *TestEnviorment {
	env.settings.SUT = &settings
	return env
}

func (env *TestEnviorment) WithPubSub(settings PubSubSettings) *TestEnviorment {
	env.settings.PubSub = &settings
	return env
}

func (env *TestEnviorment) WithRedis(settings RedisSettings) *TestEnviorment {
	env.settings.Redis = &settings
	return env
}

func (env *TestEnviorment) WithFTP(settings FTPSettings) *TestEnviorment {
	env.settings.FTP = &settings
	return env
}

func (env *TestEnviorment) WithTLS(settings TLSSettings) *TestEnviorment {
	env.settings.TLS = &settings
	return env
}
