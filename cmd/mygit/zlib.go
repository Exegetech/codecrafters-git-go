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

func zlibCompress(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	zlibWriter := zlib.NewWriter(&buffer)

	_, err := zlibWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("Error writing data to zlib writer: %s", err)
	}
	zlibWriter.Close()

	return buffer.Bytes(), nil
}
