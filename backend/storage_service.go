package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type StorageService interface {
	Store(id string, reader io.Reader) error
	Get(id string) (io.ReadCloser, error)
	Delete(id string) error
}

// LocalStorageService implements StorageService for local file system storage
type LocalStorageService struct {
	storagePath string
}

func newLocalStorageService() *LocalStorageService {
	// Create storage directory if it doesn't exist with sudo
	storagePath := "/tmp/filehost"
	cmd := exec.Command("sudo", "mkdir", "-p", storagePath)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	// Set permissions to allow the application to read/write
	cmd = exec.Command("sudo", "chmod", "777", storagePath)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	return &LocalStorageService{storagePath: storagePath}
}

func (s *LocalStorageService) Store(id string, reader io.Reader) error {
	path := filepath.Join(s.storagePath, id)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	// Set proper permissions with sudo
	cmd := exec.Command("sudo", "chmod", "644", path)
	return cmd.Run()
}

func (s *LocalStorageService) Get(id string) (io.ReadCloser, error) {
	path := filepath.Join(s.storagePath, id)
	return os.Open(path)
}

func (s *LocalStorageService) Delete(id string) error {
	path := filepath.Join(s.storagePath, id)
	cmd := exec.Command("sudo", "rm", "-f", path)
	return cmd.Run()
}

// CloudStorageService is a placeholder implementation for cloud storage
type CloudStorageService struct {
	// Add cloud storage client configuration here
}

func newCloudStorageService() *CloudStorageService {
	return &CloudStorageService{}
}

func (s *CloudStorageService) Store(id string, reader io.Reader) error {
	// TODO: Implement cloud storage upload
	// Example implementation for AWS S3:
	// uploader := s3manager.NewUploader(s.session)
	// _, err := uploader.Upload(&s3manager.UploadInput{
	//     Bucket: aws.String(s.bucketName),
	//     Key:    aws.String(id),
	//     Body:   reader,
	// })
	// return err
	return nil
}

func (s *CloudStorageService) Get(id string) (io.ReadCloser, error) {
	// TODO: Implement cloud storage download
	// Example implementation for AWS S3:
	// result, err := s.s3Client.GetObject(&s3.GetObjectInput{
	//     Bucket: aws.String(s.bucketName),
	//     Key:    aws.String(id),
	// })
	// return result.Body, err
	return nil, nil
}

func (s *CloudStorageService) Delete(id string) error {
	// TODO: Implement cloud storage deletion
	// Example implementation for AWS S3:
	// _, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
	//     Bucket: aws.String(s.bucketName),
	//     Key:    aws.String(id),
	// })
	// return err
	return nil
}