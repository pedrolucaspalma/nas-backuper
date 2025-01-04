package s3

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Uploader struct {
	region   string
	fileName string
	bucket   string
}

type S3UploaderOption func(*Uploader) S3UploaderOption

func NewUploader(bucket string) *Uploader {
	now := time.Now()
	defaultFileName := fmt.Sprintf("./backup-%d-%d-%d.zip", now.Year(), now.Month(), now.Day())
	return &Uploader{
		region:   "us-east-1",
		fileName: defaultFileName,
		bucket:   bucket,
	}
}

func Region(r string) S3UploaderOption {
	return func(u *Uploader) S3UploaderOption {
		previous := u.region
		if r == "" {
			return Region(previous)
		}
		u.region = r
		return Region(previous)
	}
}

func FileName(f string) S3UploaderOption {
	return func(u *Uploader) S3UploaderOption {
		previous := u.fileName
		if f == "" {
			return FileName(previous)
		}
		u.fileName = f
		return Region(previous)
	}
}

func (u *Uploader) Option(opts ...S3UploaderOption) (previous S3UploaderOption) {
	for _, opt := range opts {
		previous = opt(u)
	}
	return previous
}

func (u *Uploader) Send(reader io.ReadSeeker) error {
	log.Print("Creating aws session...")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(u.region),
	})
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	log.Print("Uploading folder...")
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(u.fileName),
		Body:   reader,
	})
	if err != nil {
		return err
	}
	return nil
}
