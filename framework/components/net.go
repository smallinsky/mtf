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
	defer close(c.ready)
	if networkExists("mtf_net") {
		return
	}
	run("docker network create --driver bridge mtf_net")
}

func (c *Net) Stop() {
	return
	run("docker network rm mtf_net")
}

func (c *Net) Ready() {
	return
	<-c.ready
}
