package dao

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func createTestDb() (dao Top90DAO, pool *dockertest.Pool, res *dockertest.Resource, err error) {
	pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, &dockertest.Pool{}, &dockertest.Resource{}, fmt.Errorf("could not construct pool: %v", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, pool, &dockertest.Resource{}, fmt.Errorf("could not connect to docker: %v", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, pool, nil, fmt.Errorf("could not start resource: %v", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	resource.Expire(120)

	var db *sql.DB

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second
	err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	})
	if err != nil {
		pool.Purge(resource)
		return nil, pool, resource, fmt.Errorf("could not connect to docker: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	var mig *migrate.Migrate
	mig, err = migrate.NewWithDatabaseInstance(
		"file://../migrations",
		"postgres", driver)

	err = mig.Up()
	if err != nil {
		pool.Purge(resource)
		return nil, pool, resource, fmt.Errorf("could not migrate database: %v", err)
	}

	dbx := sqlx.NewDb(db, "postgres")
	return NewPostgresDAO(dbx), pool, resource, nil
}
