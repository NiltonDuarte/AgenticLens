package main

import (
	"fmt"
	"strings"

	"agenticlens/internal/domain"
)

type tagFlags map[string]string

func (f *tagFlags) String() string { return "" }
func (f *tagFlags) Set(input string) error {
	key, value, ok := strings.Cut(input, "=")
	if !ok {
		return fmt.Errorf("tag must use KEY=VALUE")
	}
	tags, err := domain.NormalizeTags(map[string]string{key: value})
	if err != nil {
		return err
	}
	if *f == nil {
		*f = make(tagFlags)
	}
	for k, v := range tags {
		(*f)[k] = v
	}
	return nil
}
