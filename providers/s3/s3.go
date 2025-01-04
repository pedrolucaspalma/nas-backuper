package s3

import (
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Uploader struct {
	region   string
	fileName string
}

type S3UploaderOption func(*Uploader) S3UploaderOption

func Region(r string) S3UploaderOption {
	return func(u *Uploader) S3UploaderOption {
		previous := u.region
		u.region = r
		return Region(previous)
	}
}

func FileName(f string) S3UploaderOption {
	return func(u *Uploader) S3UploaderOption {
		previous := u.fileName
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

func (u *Uploader) Send(reader io.ReadSeeker, bucket string) error {
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
		Bucket: aws.String(bucket),
		Key:    aws.String(u.fileName),
		Body:   reader,
	})
	if err != nil {
		return err
	}
	return nil
}
