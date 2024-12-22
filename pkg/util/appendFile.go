package util

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func AppendFile(filePath string, fileContent string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Panicf("Error closing file: %v", err)
		}
	}(file)
	if _, err := file.WriteString(fileContent); err != nil {
		panic(err)
	}
}
