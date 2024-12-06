package main

import (
	"fmt"
	"os"
)

func readFromSHA1(sha1 string) ([]byte, error) {
	dir := sha1[0:2]
	filename := sha1[2:]

	filePath := fmt.Sprintf(".git/objects/%s/%s", dir, filename)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return []byte{}, fmt.Errorf("Error reading file: %s\n", err)
	}

	return contents, nil
}
