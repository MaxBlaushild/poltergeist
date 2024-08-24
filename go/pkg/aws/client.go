package aws

import (
	"bytes"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type client struct {
	s3 *s3.S3
}

type AWSClient interface {
	UploadImageToS3(bucket, key string, image []byte) error
	GeneratePresignedURL(bucket, key string, expiry time.Duration) (string, error)
	GeneratePresignedUploadURL(bucket, key string, expiry time.Duration) (string, error)
}

func NewAWSClient(region string) AWSClient {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	return &client{
		s3: s3.New(sess),
	}
}

func (client *client) UploadImageToS3(bucket, key string, image []byte) error {
	_, err := client.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(image),
		ACL:    aws.String("public-read"),
	})
	return err
}

func (client *client) GeneratePresignedURL(bucket, key string, expiry time.Duration) (string, error) {
	req, _ := client.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(expiry)
	return urlStr, err
}

func (client *client) GeneratePresignedUploadURL(bucket, key string, expiry time.Duration) (string, error) {
	req, _ := client.s3.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(expiry)
	return urlStr, err
}
