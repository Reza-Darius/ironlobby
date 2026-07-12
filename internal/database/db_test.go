package database

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/reza-darius/ironlobby/internal/utils"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *Queries

func TestMain(m *testing.M) {
	utils.InitLogging()
	ctx := context.Background()

	pg, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}
	// no ctx here on purpose - don't want teardown tied to a
	// context that might get cancelled early
	defer func() {
	}()

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	db, err := NewDB(connStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	testDB = db

	code := m.Run()

	if err := pg.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}
	log.Print("shut down container")

	os.Exit(code)
}

func TestInsert(t *testing.T) {
	ctx := context.Background()
	id, err := testDB.InsertPlayer(ctx, "weixiao")
	if err != nil {
		t.Fatalf("failed to insert player %v", err)
	}
	t.Logf("inserterd player weixiao, id = %v", id)
}
