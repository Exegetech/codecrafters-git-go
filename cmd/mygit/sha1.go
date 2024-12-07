package main

import (
	"crypto/sha1"
	"fmt"
	"os"
)

func readFromSHA1(sha1Hex string) ([]byte, error) {
	dir := sha1Hex[0:2]
	filename := sha1Hex[2:]

	filePath := fmt.Sprintf(".git/objects/%s/%s", dir, filename)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return []byte{}, fmt.Errorf("Error reading file: %s\n", err)
	}

	return contents, nil
}

func writeToSHA1(sha1 []byte, data []byte) error {
	hex := fmt.Sprintf("%x", sha1)
	dir := hex[0:2]

	err := os.Mkdir(".git/objects/"+dir, 0755)
	if err != nil {
		return fmt.Errorf("Error creating directory: %s\n", err)
	}

	filename := hex[2:]

	filePath := fmt.Sprintf(".git/objects/%s/%s", dir, filename)
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("Error writing file: %s\n", err)
	}

	return nil
}

func computeSHA1(data []byte) ([]byte, error) {
	hash := sha1.New()
	if _, err := hash.Write(data); err != nil {
		return []byte{}, fmt.Errorf("Error writing data to hash: %s\n", err)
	}

	summed := hash.Sum(nil)
	return summed, nil
}
