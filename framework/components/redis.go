package components

func NewRedis() *Redis {
	return &Redis{
		ready: make(chan struct{}),
	}
}

type Redis struct {
	Pass     string
	Port     string
	DB       []string
	Hostname string
	Network  string
	ready    chan struct{}
}

func (c *Redis) Start() {
	cmd := `docker run --rm -d --network=mtf_net --name redis_mtf --hostname=redis_mtf --env REDIS_PASSWORD=test -p 6379:6379 bitnami/redis:4.0`
	run(cmd)
	close(c.ready)
}

func (c *Redis) Stop() {
	run("docker stop redis_mtf")
}

func (c *Redis) Ready() {
	<-c.ready
}
