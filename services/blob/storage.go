package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

const storageDir = "data"

func init() {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0755)
	}
}

func FileDelete(filename string) error {
	filePath := filepath.Join(storageDir, filename)
	return os.Remove(filePath)
}

func FileGet(filename string) (*os.File, error) {
	filePath := filepath.Join(storageDir, filename)
	return os.Open(filePath)
}

func FileSave(filename string, data io.Reader) error {
	filePath := filepath.Join(storageDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	return err
}

func FilesList() ([]string, error) {
	filenames := make([]string, 0)
	if err := filepath.Walk(storageDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				filenames = append(filenames, strings.TrimPrefix(path, storageDir+"/"))
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return filenames, nil
}
