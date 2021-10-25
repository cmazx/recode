package storage

import (
	"net/http"
	"os"
)

type Storage interface {
	Put(localPath string, storagePath string, isPublic bool) (string, error)
	GetObjectUrl(storagePath string) string
}

func GetFileContentType(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
