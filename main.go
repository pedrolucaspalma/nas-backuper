package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pedrolucaspalma/nas-backuper/compressor"
	"github.com/pedrolucaspalma/nas-backuper/providers/s3"
)

func main() {
	input := parseInput()
	fmt.Printf("%+v\n", input)
	f, err := compressor.GetFolder(input.inputPath)
	if err != nil {
		log.Fatal(err)
	}

	uploader := &s3.Uploader{}
	uploader.Option(
		s3.Region(input.region),
		s3.FileName(input.fileName),
	)

	reader := bytes.NewReader(f)
	err = uploader.Send(reader, input.bucket)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Backup finished successfully!")
}

type input struct {
	inputPath       string
	bucket          string
	region          string
	fileName        string
	mustUseTempFile bool
}

func parseInput() input {
	// Parsing optional flags
	// TODO feature
	mustUseTempFile := flag.Bool("tempFile", false, "If the program must temporarily write the compressed file to the disk before sending to cloud provider. This should reduce memory usage when handling large files, but makes the program bound to disk I/O.")
	region := flag.String("region", "us-east-1", "The cloud provider region")

	now := time.Now()
	defaultFileName := fmt.Sprintf("./backup-%d-%d-%d.zip", now.Year(), now.Month(), now.Day())
	fileName := flag.String("name", defaultFileName, "The compressed file's name to be used when uploading to cloud provider")

	flag.Parse()

	// Parsing mandatory arguments
	args := flag.Args()
	if len(args) != 2 {
		log.Fatal("Invalid args. You must provide a relative path to the directory and the bucket name.")
	}

	inputPath := args[0]

	if filepath.IsAbs(inputPath) {
		log.Fatal("Invalid path. It must be a relative path to a directory.")
	}

	_, err := os.Stat(inputPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("Invalid path. Directory not found.")
		}
		log.Fatal("Invalid path:", err)
	}

	bucketName := args[1]

	if len(bucketName) == 0 {
		log.Fatal("Invalid bucket name.")
	}

	return input{
		inputPath:       inputPath,
		bucket:          bucketName,
		region:          *region,
		fileName:        *fileName,
		mustUseTempFile: *mustUseTempFile,
	}
}
