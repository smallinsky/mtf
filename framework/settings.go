package framework

type Settings struct {
	MySQL     *MysqlSettings
	SUT       *SutSettings
	PubSub    *PubSubSettings
	Redis     *RedisSettings
	FTP       *FTPSettings
	TLS       *TLSSettings
	Migration []*MigrationSettings
}

type MigrationSettings struct {
	Password string
	Port     string
	DBName   string
	Dir      string
}

type MysqlSettings struct {
	// DatabaseName is name of a database that will be created
	DatabaseName string
	Databases    []string
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

	// RuntimeTypeCommand indicates that system under test is a simile binary that will
	// terminate after execution.
	RuntimeTypeCommand
)

// SutSettings system under test settings used to build and executed sut in docker container.
type SutSettings struct {
	// Envs allows to pass custom env to system under test container.
	Envs []string
	// Dir is a path to directory that collect system under test source for which binary should be build.
	Dir string
	// Ports is a collection of ports that sut binary require, those ports will be forwarded to local host with the
	// same port mapping.
	Ports []int
	// Mounts allows to pass list of directories which should be mountet
	// inside system under test container in form 'src:dst'.
	Mounts []string

	// RuntimeType Type of system under test runtime. In case of service runtime sut component will
	// be executed once, but when runtime type is set to command (terminates after execution) sut component
	// needs to be re-executed for each test case.
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

func (env *TestEnvironment) WithMySQL(settings MysqlSettings) *TestEnvironment {
	env.settings.MySQL = &settings
	return env
}

func (env *TestEnvironment) WithMigration(settings []*MigrationSettings) *TestEnvironment {
	env.settings.Migration = settings
	return env
}

func (env *TestEnvironment) WithSUT(settings SutSettings) *TestEnvironment {
	env.settings.SUT = &settings
	return env
}

func (env *TestEnvironment) WithPubSub(settings PubSubSettings) *TestEnvironment {
	env.settings.PubSub = &settings
	return env
}

func (env *TestEnvironment) WithRedis(settings RedisSettings) *TestEnvironment {
	env.settings.Redis = &settings
	return env
}

func (env *TestEnvironment) WithFTP(settings FTPSettings) *TestEnvironment {
	env.settings.FTP = &settings
	return env
}

func (env *TestEnvironment) WithTLS(settings TLSSettings) *TestEnvironment {
	env.settings.TLS = &settings
	return env
}
