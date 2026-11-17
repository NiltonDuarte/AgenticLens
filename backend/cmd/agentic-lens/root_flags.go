package main

import "strings"

type repeatStringFlag []string

func (s *repeatStringFlag) String() string         { return strings.Join(*s, ",") }
func (s *repeatStringFlag) Set(value string) error { *s = append(*s, value); return nil }
