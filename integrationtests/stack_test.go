package integrationtests

import (
	"context"
	"testing"
)

func Test_PSQLContainer(t *testing.T) {
	ctx := context.Background()
	psqlContainer, err := startPSQLContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := psqlContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
}
