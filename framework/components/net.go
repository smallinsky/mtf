package components

type Net struct {
}

func (c *Net) Start() {
	run("docker network create --driver bridge mtf_net")
}

func (c *Net) Stop() {
	run("docker network rm mtf_net")
}

func (c *Net) Ready() {
}
