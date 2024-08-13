package aws

import (
	"bytes"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AWSClient struct {
	S3 *s3.S3
}

func NewAWSClient(region string) *AWSClient {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	return &AWSClient{
		S3: s3.New(sess),
	}
}

func (client *AWSClient) UploadImageToS3(bucket, key string, image []byte) error {
	_, err := client.S3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(image),
		ACL:    aws.String("public-read"),
	})
	return err
}

func (client *AWSClient) GeneratePresignedURL(bucket, key string, expiry time.Duration) (string, error) {
	req, _ := client.S3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(expiry)
	return urlStr, err
}

func (client *AWSClient) GeneratePresignedUploadURL(bucket, key string, expiry time.Duration) (string, error) {
	req, _ := client.S3.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	urlStr, err := req.Presign(expiry)
	return urlStr, err
}
