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
	inputPath := parseInput()
	f, err := getFolderCompressed(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	err = sendToS3(f)
	if err != nil {
		log.Fatal(err)
	}
}

func parseInput() string {
	if len(os.Args) != 2 {
		log.Fatal("Invalid args. Only provide a relative path to the directory.")
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
	return inputPath
}

func sendToS3(file []byte) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	now := time.Now()
	objectName := fmt.Sprintf("./backup-%d-%d-%d-%d.zip", now.Year(), now.Month(), now.Day(), now.Second())

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String(objectName),
		Body:   bytes.NewReader(file),
	})
	if err != nil {
		return err
	}
	return nil
}

func getFolderCompressed(folderPath string) ([]byte, error) {
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
