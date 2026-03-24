package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ImageStorage interface {
	SaveImage(ctx context.Context, file multipart.File, filename string) (string, error)
	DeleteImage(ctx context.Context, imagePath string) error
}

type FileSystemStorage struct {
	uploadDir string
}

func NewFileSystemStorage(uploadDir string) *FileSystemStorage {
	os.MkdirAll(uploadDir, 0755)
	return &FileSystemStorage{uploadDir: uploadDir}
}

func (fs *FileSystemStorage) SaveImage(ctx context.Context, file multipart.File, filename string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}

	fileType := http.DetectContentType(buffer)
	if !strings.HasPrefix(fileType, "image/") {
		return "", fmt.Errorf("file is not an image: %s", fileType)
	}

	file.Seek(0, 0)

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)
	uniqueName := fmt.Sprintf("%s_%d%s", baseName, time.Now().Unix(), ext)

	filePath := filepath.Join(fs.uploadDir, uniqueName)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return fmt.Sprintf("/uploads/%s", uniqueName), nil
}

func (fs *FileSystemStorage) DeleteImage(ctx context.Context, imagePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	filePath := filepath.Join(fs.uploadDir, filepath.Base(imagePath))
	return os.Remove(filePath)
}
