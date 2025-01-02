package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	now := time.Now()
	dest := fmt.Sprintf("./backup-%d-%d-%d-%d.zip", now.Year(), now.Month(), now.Day(), now.Second())

	inputPath := parseInput()

	compressFolder(inputPath, dest)
	sendToS3(dest)
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

func sendToS3(dest string) {
	_, err := os.ReadFile(dest)
	if err != nil {
		fmt.Errorf(err.Error())
	}
}

func compressFolder(folderPath string, destination string) error {
	// Create the destination zip file
	zipFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through the source folder
	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
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
}
