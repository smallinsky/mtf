// +build docker

package mysql

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/smallinsky/mtf/pkg/docker"
)

func TestMysql(t *testing.T) {
	dbConfig := MySQLConfig{
		Database: "test_db",
		Password: "test",
		Labels: map[string]string{
			"mtf": "",
		},
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	dockerCli, err := docker.NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	filter := filters.NewArgs()
	filter.Add("label", "mtf")
	fmt.Println(filter)

	resp, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filter,
	})
	if err != nil {
		t.Fatalf("failed to list containers: %v", err)
	}
	resp = resp

	mysqlCom := NewMySQL(dockerCli, dbConfig)
	err = mysqlCom.Start()
	if err != nil {
		t.Fatalf("faield to start component %v", err)
	}
	return

	err = mysqlCom.Ready()
	if err != nil {
		t.Fatalf("faield to get ready state %v", err)
	}
}
