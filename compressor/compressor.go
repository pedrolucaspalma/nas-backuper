package compressor

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

func GetFolder(path string) ([]byte, error) {
	log.Print("Reading folder and compressing...")
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	defer w.Close()

	// Walk through the source folder
	err := zipFolder(w, path)

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil

}

func zipFolder(w *zip.Writer, folderPath string) error {
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
		writer, err := w.CreateHeader(header)
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
	return err
}
