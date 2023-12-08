package cmd

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/clients/top90"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/dao"
)

type Init struct {
	logger *slog.Logger
}

func NewInit(logger *slog.Logger) Init {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return Init{logger: logger}
}

func (i Init) DB(user, password, name, host, port string) *sqlx.DB {
	DB, err := db.NewPostgresDB(user, password, name, host, port)
	if err != nil {
		i.Exit("Failed setting up database", err)
	}
	return DB
}

func (i Init) Dao(db *sqlx.DB) dao.Top90DAO {
	return dao.NewPostgresDAO(db)
}

func (i Init) S3Client(cfg s3.Config, bucket string) s3.S3Client {
	s3Client, err := s3.NewClient(cfg)
	if err != nil {
		i.Exit("Failed creating s3 client", err)
	}

	err = s3Client.VerifyConnection(bucket)
	if err != nil {
		i.Exit("Failed connecting to s3 bucket", err)
	}
	return *s3Client
}

func (i Init) RedditClient(timeout time.Duration) reddit.Client {
	client, err := reddit.NewClient(reddit.Config{
		Timeout: timeout,
		Logger:  i.logger,
	})
	if err != nil {
		i.Exit("Failed creating reddit client", err)
	}
	return *client
}

func (i Init) Top90Client(timeout time.Duration) top90.Client {
	return top90.NewClient(top90.Config{
		Timeout: timeout,
	})
}

func (i Init) ChromeDP() context.Context {
	const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/77.0.3830.0 Safari/537.36"

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
	)
	ctx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ = chromedp.NewContext(ctx)
	err := chromedp.Run(ctx)
	if err != nil {
		i.Exit("Failed initializing chromedp", err)
	}
	return ctx
}

func (i Init) Migrate(db *sql.DB) *migrate.Migrate {
	driver, err := pg.WithInstance(db, &pg.Config{})
	if err != nil {
		i.Exit("Failed setting up database driver", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/db/migrations",
		"postgres", driver)
	if err != nil {
		i.Exit("Could not instantiate migrate", err)
	}
	return m
}

func (i Init) Exit(msg string, err error) {
	i.logger.Error(msg, "error", err)
	os.Exit(1)
}
