package s3

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Client struct {
	session *session.Session
	s3      *s3.S3
	logger  *slog.Logger
}

type Config struct {
	AccessKey       string
	SecretAccessKey string
	Endpoint        string
	Logger          *slog.Logger
}

func NewClient(cfg Config) (*S3Client, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	awsConfig := awsConfig(cfg)
	session, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	s3 := s3.New(session)

	return &S3Client{
		session: session,
		s3:      s3,
		logger:  cfg.Logger,
	}, nil
}

func awsConfig(cfg Config) *aws.Config {
	if cfg.Endpoint == "" {
		return &aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				cfg.AccessKey,
				cfg.SecretAccessKey, ""),
		}
	}

	return &aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(os.Getenv("TOP90_AWS_S3_ENDPOINT")),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			cfg.AccessKey,
			cfg.SecretAccessKey, ""),
	}
}

func (c *S3Client) VerifyConnection(bucketName string) error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := c.s3.HeadBucket(input)
	if err != nil {
		return err
	}
	return nil
}

func (c *S3Client) UploadFile(fileName string, key string, contentType string, bucketName string) error {
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	upInput := &s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	}

	uploader := s3manager.NewUploader(c.session)
	_, err = uploader.UploadWithContext(context.Background(), upInput)
	if err != nil {
		return err
	}

	return nil
}

func (c *S3Client) DownloadFile(key, bucket, outputFilename string) error {
	file, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(c.session)
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (c *S3Client) DownloadFileBytes(key string, bucket string) ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(c.session)
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *S3Client) DeleteObject(key string, bucketName string) error {
	_, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	return err
}

func (c *S3Client) HeadObject(key string, bucketName string) (*s3.HeadObjectOutput, error) {
	out, err := c.s3.HeadObject(&s3.HeadObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(bucketName),
	})
	return out, err
}

func (c *S3Client) ListAllObjects(bucket string) ([]string, error) {
	var keys []string
	i := 0
	err := c.s3.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: &bucket,
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		i++
		for _, obj := range p.Contents {
			keys = append(keys, *obj.Key)
		}
		return true
	})
	return keys, err
}

func (c *S3Client) PresignedUrl(key string, bucket string, expire time.Duration) (string, error) {
	req, _ := c.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	url, err := req.Presign(expire)
	if err != nil {
		return "", err
	}

	// When running in docker locally, host.docker.internal will be in the url
	result := strings.Replace(url, "host.docker.internal", "localhost", -1)
	return result, nil
}
