package integrationtests

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ATMackay/psql-ledger/service"
	_ "github.com/jackc/pgx/v4/stdlib"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresUsr  = "root"
	postgresPswd = "secret"
	postgresDB   = "testdb"
)

type postgresDBContainer struct {
	testcontainers.Container
	host string
	port int
}

func startPSQLContainer(ctx context.Context) (*postgresDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		Name:         "postgres",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_USER": postgresUsr, "POSTGRES_PASSWORD": postgresPswd, "POSTGRES_DB": postgresDB},
		WaitingFor:   wait.ForLog("database system is ready to accept connections").WithStartupTimeout(5 * time.Second),
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

	p, _ := strconv.Atoi(mappedPort.Port())

	return &postgresDBContainer{Container: container, host: hostIP, port: p}, nil
}

type stack struct {
	psql       *postgresDBContainer
	psqlLedger *service.Service
}

func createStack(t *testing.T) *stack {
	ctx := context.Background()

	// start postgres container
	psqlContainer, err := startPSQLContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cfg := service.DefaultConfig

	// get config details
	cfg.PostgresHost = psqlContainer.host
	cfg.PostgresPort = psqlContainer.port
	cfg.PostgresUser = postgresUsr
	cfg.PostgresPassword = postgresPswd
	cfg.PostgresDB = postgresDB

	time.Sleep(500 * time.Millisecond) // TODO - code smell, fix

	psqlLedger, err := service.BuildService(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {

		// Kill service
		psqlLedger.Stop(os.Kill)

		// kill PSQL container
		if err := psqlContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Start HTTP service
	psqlLedger.Start()

	return &stack{psql: psqlContainer, psqlLedger: psqlLedger}
}
