package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		initGit()

	case "cat-file":
		catFile(os.Args[3])

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func initGit() {
	initialDirs := []string{
		".git",
		".git/objects",
		".git/refs",
	}

	for _, dir := range initialDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			os.Exit(1)
		}
	}

	headFileName := ".git/HEAD"
	headFileContents := "ref: refs/heads/main\n"

	if err := os.WriteFile(headFileName, []byte(headFileContents), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Initialized git directory")
}

func catFile(sha1 string) {
	compressed, err := readFromSHA1(sha1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	decompressed, err := zlibDecompress(compressed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decompressing file: %s\n", err)
		os.Exit(1)
	}

	if strings.HasPrefix(decompressed, "blob") {
		blob, err := parseBlob(decompressed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing blob: %s\n", err)
			os.Exit(1)
		}

		fmt.Print(blob.data)
	}
}

var nullByte = '\x00'

type blob struct {
	size int
	data string
}

func parseBlob(str string) (blob, error) {
	startIdx := strings.Index(str, "blob ") + len("blob ")
	j := 0

	for i, ch := range str[startIdx:] {
		if ch == nullByte {
			j = i
			break
		}
	}

	sizeSection := str[startIdx : startIdx+j]
	size, err := strconv.Atoi(sizeSection)
	if err != nil {
		return blob{}, fmt.Errorf("Error converting size to int: %s", err)
	}

	data := str[startIdx+j+1:]

	return blob{
		size,
		data,
	}, nil
}
