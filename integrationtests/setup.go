package integrationtests

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

type postgresDBContainer struct {
	testcontainers.Container
	URI string
}

func startPSQLContainer(ctx context.Context) (*postgresDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		Name:         "postgres",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_USER": "root", "POSTGRES_PASSWORD": "secret", "POSTGRES_DB": "bank"},
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("postgres://root@%s:%s", hostIP, mappedPort.Port())

	return &postgresDBContainer{Container: container, URI: uri}, nil
}
