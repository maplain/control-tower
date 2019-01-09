package error

import (
	"log"
)

const (
	FileNotFoundError = Error("file not found")
)

func Check(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

type Error string

func (e Error) Error() string {
	return string(e)
}
