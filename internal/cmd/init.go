package cmd

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/wweitzel/top90/internal/clients/reddit"
	"github.com/wweitzel/top90/internal/clients/s3"
	"github.com/wweitzel/top90/internal/db"
	"github.com/wweitzel/top90/internal/db/postgres/dao"
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

func (i Init) S3Client(cfg s3.Config, bucket string) s3.S3Client {
	s3Client, err := s3.NewClient(cfg)
	if err != nil {
		i.exit("Failed creating s3 client", err)
	}

	err = s3Client.VerifyConnection(bucket)
	if err != nil {
		i.exit("Failed connecting to s3 bucket", err)
	}
	return *s3Client
}

func (i Init) Dao(user, password, name, host, port string) db.Top90DAO {
	DB, err := db.NewPostgresDB(user, password, name, host, port)
	if err != nil {
		i.exit("Failed setting up database", err)
	}

	return dao.NewPostgresDAO(DB)
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
		i.exit("Failed initializing chromedp", err)
	}
	return ctx
}

func (i Init) RedditClient(timeout time.Duration) reddit.Client {
	client, err := reddit.NewClient(reddit.Config{
		Timeout: timeout,
		Logger:  i.logger,
	})
	if err != nil {
		i.exit("Failed creating reddit client", err)
	}
	return *client
}

func (i Init) exit(msg string, err error) {
	i.logger.Error(msg, "erorr", err)
	os.Exit(1)
}
