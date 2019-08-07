package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

func foo() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("faield to create docker client: %v", err)
	}
	portMap := PortMap{
		8001: 8001,
	}
	exposedPorts := make(nat.PortSet)

	for k := range portMap {
		exposedPorts[toNatPort(k)] = struct{}{}
	}

	result, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Hostname:     "run_sut",
			AttachStdin:  true,
			AttachStdout: true,
			ExposedPorts: exposedPorts,
			Env: []string{
				"SUT_BINARY_NAME=echo",
				"ORACLE_ADDR=host.docker.internal:8002",
			},
			Image:      "run_sut",
			Entrypoint: nil,
			Labels: map[string]string{
				"": "",
			},
		},
		&container.HostConfig{
			PortBindings: portMap.toNatPortMap(),
			AutoRemove:   false,
			VolumeDriver: "",
			VolumesFrom:  nil,
			Mounts: []mount.Mount{
				mount.Mount{
					Type:   mount.TypeBind,
					Source: "/Users/Marek/Go/src/github.com/smallinsky/mtf/e2e/service/echo",
					Target: "/component",
				},
				mount.Mount{
					Type:   mount.TypeBind,
					Source: "/tmp/mtf/cert",
					Target: "/usr/local/share/ca-certificates",
				},
			},
			CapAdd: strslice.StrSlice{"NET_RAW", "NET_ADMIN"},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"mtf_net": {
					Aliases: []string{"networkalias"},
				},
			},
		},
		"",
	)
	if err != nil {
		log.Fatalf("faield to create docker container: %v", err)
	}
	if err := cli.ContainerStart(context.Background(), result.ID, types.ContainerStartOptions{}); err != nil {
		_ = cli.Close()
		log.Fatalf("faield to start a container: %v", err)
	}
	log.Printf("docker running %s %s\n", result.ID, result.Warnings)

	iresp, err := cli.ContainerInspect(context.Background(), result.ID)
	if err != nil {
		log.Fatalf("faield to start a container: %v", err)
		_ = cli.Close()
	}
	iresp = iresp

	time.Sleep(time.Second * 5)
	r, err := cli.ContainerLogs(context.Background(), result.ID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
	})
	if err != nil {
		log.Fatalf("failed to get container logs: %v", err)
	}
	defer func() { _ = r.Close() }()

	// Read stdout and stderr to the same buffer.
	var allOutput bytes.Buffer
	if _, err = stdcopy.StdCopy(&allOutput, &allOutput, r); err != nil {
		log.Fatalf("failed to get container logs: %v", err)
	}

	fmt.Println(allOutput.String())

}

type ContainerPort int
type HostPort int
type PortMap map[ContainerPort]HostPort

func (m PortMap) toNatPortMap() nat.PortMap {
	out := make(nat.PortMap)
	for k, v := range m {
		out[toNatPort(k)] = []nat.PortBinding{{HostPort: strconv.Itoa(int(v))}}
	}
	return out
}
func toNatPort(p ContainerPort) nat.Port {
	return nat.Port(fmt.Sprintf("%d/tcp", p))
}

func main() {
	foo()
}
