package error

import (
	"fmt"
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

func Warn(w string) {
	fmt.Printf("warning: %s\n", w)
}

func Warnf(format string, a ...interface{}) {
	fmt.Printf("warning: "+format, a...)
}
