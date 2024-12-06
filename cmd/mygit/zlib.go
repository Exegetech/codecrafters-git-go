package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
)

func zlibDecompress(data []byte) (string, error) {
	buffer := bytes.NewBuffer(data)
	zlibReader, err := zlib.NewReader(buffer)
	if err != nil {
		return "", fmt.Errorf("Error creating zlib reader: %s", err)
	}
	defer zlibReader.Close()

	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(zlibReader)
	if err != nil {
		return "", fmt.Errorf("Error reading from zlib reader: %s", err)
	}

	return decompressed.String(), nil
}
