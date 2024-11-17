package zipservice

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"zip-api/internal/core/entities"
)

// Errors
var (
	ErrIncorrectMimeType = errors.New("not allowed mimetype provided in archive file")
)

var (
	AllowedMimeTypes = []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/xml",
		"image/jpeg",
		"image/png",
	}
)

type zipService struct {
}

func NewZipService() *zipService {
	return &zipService{}
}

func (s *zipService) ZipInfo(zipArchiveBinaryReader io.Reader, zipName string) (*entities.Archive, error) {
	zipFile, err := os.CreateTemp("", "*.zip")
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	_, err = io.Copy(zipFile, zipArchiveBinaryReader)
	if err != nil {
		return nil, err
	}

	stat, err := zipFile.Stat()
	if err != nil {
		return nil, err
	}

	// Read the ZIP archive from the io.Reader
	fmt.Println(stat.Size())
	zipReader, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return nil, err
	}

	// Prepare the Archive struct
	archive := &entities.Archive{
		FileName:   zipName,
		Size:       uint32(stat.Size()),         // size of the ZIP file in bytes
		TotalSize:  0,                           // Total uncompressed size of files in the archive
		TotalFiles: uint32(len(zipReader.File)), // Total number of files
	}

	sniff := make([]byte, 512)
	// Iterate through each file in the ZIP archive
	for _, file := range zipReader.File {
		fileReader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer fileReader.Close()

		fileReader.Read(sniff)

		contentType := http.DetectContentType(sniff)
		if !In(contentType, AllowedMimeTypes) {
			return nil, ErrIncorrectMimeType
		}

		// Update the total uncompressed size
		archive.TotalSize += uint32(file.UncompressedSize64)

		// Extract file metadata
		file := entities.File{
			FilePath: file.Name,
			Size:     uint32(file.UncompressedSize64),
			MimeType: contentType,
		}

		// Append the file metadata to the archive
		archive.Files = append(archive.Files, file)
	}

	return archive, nil

}

func (s *zipService) ZipArchive(files []io.Reader) ([]byte, error) {
	return []byte{}, nil
}

func In(s string, strs []string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}