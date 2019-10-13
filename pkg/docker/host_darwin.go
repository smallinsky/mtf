package docker

// HostIP return osx host IP
// https://docs.docker.com/docker-for-mac/networking/
func HostIP() (string, error) {
	return "host.docker.internal", nil
}
