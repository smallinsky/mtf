package components

func NewNet() *Net {
	return &Net{
		ready: make(chan struct{}),
	}
}

type Net struct {
	ready chan struct{}
}

func (c *Net) Start() {
	run("docker network create --driver bridge mtf_net")
	close(c.ready)
}

func (c *Net) Stop() {
	run("docker network rm mtf_net")
}

func (c *Net) Ready() {
	<-c.ready
}
