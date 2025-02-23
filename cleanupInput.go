package main

import "strings"

func cleanupInput(s string) string {
	s = strings.ReplaceAll(s, "kerfuffle", "****")
	s = strings.ReplaceAll(s, "sharbert", "****")
	s = strings.ReplaceAll(s, "fornax", "****")
	s = strings.ReplaceAll(s, "Fornax", "****")
	return s
}
