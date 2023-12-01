package s3

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
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
}

// Create new s3 client
func NewClient(awsAccessKey, awsSecretAccessKey string) S3Client {
	s3Config := getS3Config(awsAccessKey, awsSecretAccessKey)

	session, err := session.NewSession(s3Config)
	if err != nil {
		log.Fatalln("Failed to create aws session", err)
	}

	s3 := s3.New(session)

	var s3Client S3Client
	s3Client.session = session
	s3Client.s3 = s3

	return s3Client
}

func getS3Config(awsAccessKey, awsSecretAccessKey string) *aws.Config {
	// TODO: Should pass in a config to NewClient insttead of doing this here
	if os.Getenv("ENV") == "dev" {
		return &aws.Config{
			Region:           aws.String("us-east-1"),
			Endpoint:         aws.String(os.Getenv("TOP90_AWS_S3_ENDPOINT")),
			S3ForcePathStyle: aws.Bool(true),
			Credentials: credentials.NewStaticCredentials(
				awsAccessKey,
				awsSecretAccessKey, ""),
		}
	} else if os.Getenv("ENV") == "prod" {
		return &aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				awsAccessKey,
				awsSecretAccessKey, ""),
		}
	} else {
		log.Fatalln("environment variable ENV must be set")
		os.Exit(1)
		return nil
	}
}

// Call s3.HeadBucket to verify top90 bucket exists and we have permission to view it
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

// Upload a file to s3
func (c *S3Client) UploadFile(fileName string, key string, contentType string, bucketName string) error {
	uploader := s3manager.NewUploader(c.session)

	fileBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	upInput := &s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	}

	_, err = uploader.UploadWithContext(context.Background(), upInput)
	if err != nil {
		return err
	}

	return nil
}

// Download a file from s3
func (c *S3Client) DownloadFile(key, bucket, outputFilename string) {
	downloader := s3manager.NewDownloader(c.session)

	file, err := os.Create(outputFilename)
	if err != nil {
		log.Println("Unable to open file", outputFilename, err)
	}
	defer file.Close()

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		log.Println("Unable to download item", key, err)
	}

	log.Println("Downloaded", file.Name(), numBytes, "bytes")
}

func (c *S3Client) DownloadFileBytes(key string, bucket string) ([]byte, error) {
	downloader := s3manager.NewDownloader(c.session)

	buf := aws.NewWriteAtBuffer([]byte{})
	numBytes, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Println("Unable to download item", key, err)
		return nil, err
	}

	log.Println("Downloaded", key, numBytes, "bytes")
	return buf.Bytes(), nil
}

// Delete a file on s3
func (c *S3Client) DeleteFile(key string, bucketName string) error {
	_, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

// Create a presigned download url with an expiration time
func (c *S3Client) NewSignedGetURL(key string, bucket string, expire time.Duration) (string, error) {
	req, _ := c.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expire)
	if err != nil {
		return "", fmt.Errorf(key, err)
	}

	// When running in docker locally, host.docker.internal will be in the url
	result := strings.Replace(url, "host.docker.internal", "localhost", -1)
	return result, nil
}

func (c *S3Client) PresignedUrl(objectKey string, bucket string) string {
	url, err := c.NewSignedGetURL(objectKey, bucket, time.Minute*10)
	if err != nil {
		log.Println(err)
	}
	return url
}
