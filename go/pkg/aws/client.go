package aws

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type client struct {
	s3 *s3.S3
}

type AWSClient interface {
	UploadImageToS3(bucket, key string, image []byte) (string, error)
	DeleteObjectFromS3(bucket, key string) error
	GetObjectLastModified(bucket, key string) (*time.Time, error)
	GeneratePresignedURL(bucket, key string, expiry time.Duration) (string, error)
	GeneratePresignedUploadURL(bucket, key string, expiry time.Duration) (string, error)
	GeneratePresignedUploadURLWithContentType(bucket, key string, contentType string, expiry time.Duration) (string, error)
}

// ObjectClient extends the shared S3 client without changing the older media
// signing interface used by existing services and test doubles.
type ObjectClient interface {
	PutObject(context.Context, string, string, string, []byte) error
	ReadObject(context.Context, string, string, int64) ([]byte, error)
	DeleteObject(context.Context, string, string) error
}

func (client *client) PutObject(ctx context.Context, bucket, key, contentType string, data []byte) error {
	_, err := client.s3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(data),
		ContentType: aws.String(contentType), CacheControl: aws.String("private, no-store"),
	})
	return err
}

func (client *client) ReadObject(ctx context.Context, bucket, key string, limit int64) ([]byte, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("object read limit must be positive")
	}
	object, err := client.s3.GetObjectWithContext(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	defer object.Body.Close()
	if object.ContentLength != nil && *object.ContentLength > limit {
		return nil, fmt.Errorf("object exceeds read limit")
	}
	data, err := io.ReadAll(io.LimitReader(object.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("object exceeds read limit")
	}
	return data, nil
}

func (client *client) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := client.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	return err
}

func NewAWSClient(region string) AWSClient {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	return &client{
		s3: s3.New(sess),
	}
}

func (client *client) UploadImageToS3(bucket, key string, image []byte) (string, error) {
	_, err := client.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(image),
	})
	if err != nil {
		return "", err
	}
	return "https://" + bucket + ".s3.amazonaws.com/" + key, nil
}

func (client *client) DeleteObjectFromS3(bucket, key string) error {
	_, err := client.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (client *client) GetObjectLastModified(bucket, key string) (*time.Time, error) {
	resp, err := client.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case "NotFound", "NoSuchKey":
				return nil, nil
			}
		}
		return nil, err
	}
	return resp.LastModified, nil
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

func (client *client) GeneratePresignedUploadURLWithContentType(bucket, key string, contentType string, expiry time.Duration) (string, error) {
	req, _ := client.s3.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	})
	urlStr, err := req.Presign(expiry)
	return urlStr, err
}
