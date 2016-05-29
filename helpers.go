package main

import (
	"fmt"
	"log"
)

// StringSliceValue is a string slice value shortcut.
type StringSliceValue []string

func (f *StringSliceValue) String() string {
	return fmt.Sprintf(`%v`, *f)
}

// Set implemets flag.Value interface Set method.
func (f *StringSliceValue) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func exit(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
