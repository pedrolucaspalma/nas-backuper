package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	input := parseInput()
	f, err := getFolderCompressed(input.inputPath)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	objectName := fmt.Sprintf("./backup-%d-%d-%d.zip", now.Year(), now.Month(), now.Day())

	s3Sender := &S3Sender{
		region:   "us-east-1",
		fileName: objectName,
	}

	reader := bytes.NewReader(f)
	err = s3Sender.SendToS3(reader, input.bucket)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Backup finished successfully!")
}

type input struct {
	inputPath string
	bucket    string
}

func parseInput() input {
	if len(os.Args) != 3 {
		log.Fatal("Invalid args. You must provide a relative path to the directory and the bucket name.")
	}

	inputPath := os.Args[1]

	if string(inputPath[0]) == "/" {
		log.Fatal("Invalid path. It must be a relative path to a directory.")
	}

	_, err := os.Stat(inputPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("Invalid path. Directory not found.")
		}
		log.Fatal("Invalid path:", err)
	}

	bucketName := os.Args[2]

	if len(bucketName) == 0 {
		log.Fatal("Invalid bucket name.")
	}

	return input{
		inputPath: inputPath,
		bucket:    bucketName,
	}
}

type S3Sender struct {
	region   string
	fileName string
}

type S3SenderOption func(*S3Sender) S3SenderOption

func Region(r string) S3SenderOption {
	return func(s *S3Sender) S3SenderOption {
		previous := s.region
		s.region = r
		return Region(previous)
	}
}

func FileName(f string) S3SenderOption {
	return func(s *S3Sender) S3SenderOption {
		previous := s.fileName
		s.fileName = f
		return Region(previous)
	}
}

func (s *S3Sender) Option(opts ...S3SenderOption) (previous S3SenderOption) {
	for _, opt := range opts {
		previous = opt(s)
	}
	return previous
}

func (s *S3Sender) SendToS3(reader io.ReadSeeker, bucket string) error {
	log.Print("Creating aws session...")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.region),
	})
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	log.Print("Uploading folder...")
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s.fileName),
		Body:   reader,
	})
	if err != nil {
		return err
	}
	return nil
}

func getFolderCompressed(folderPath string) ([]byte, error) {
	log.Print("Reading folder and compressing...")
	var b bytes.Buffer
	zipWriter := zip.NewWriter(&b)
	defer zipWriter.Close()

	// Walk through the source folder
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories (they're included implicitly in zip archives)
		if info.IsDir() {
			return nil
		}

		// Create a relative path for the zip header
		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		// Create a zip file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate // Compression method

		// Create the file in the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Open the source file for reading
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Copy the file content to the zip archive
		_, err = io.Copy(writer, srcFile)
		return err
	})

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil

}
