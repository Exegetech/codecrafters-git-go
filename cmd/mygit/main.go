package main

import (
	"fmt"
	"os"
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

	case "hash-object":
		hashObject(os.Args[3])

	case "ls-tree":
		lsTree(os.Args[3])

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

	gitObj, err := parseObjectContent(decompressed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing object: %s\n", err)
		os.Exit(1)
	}

	if gitObj.getType() != "blob" {
		fmt.Fprintf(os.Stderr, "Object is not a blob\n")
		os.Exit(1)
	}

	fmt.Print(string(gitObj.(blob).content))
}

func hashObject(filepath string) {
	stats, err := os.Stat(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file stats: %s\n", err)
		os.Exit(1)
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
		os.Exit(1)
	}

	b := blob{
		size:    int(stats.Size()),
		content: data,
	}

	serialized := b.String()
	sha1, err := computeSHA1([]byte(serialized))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing SHA-1: %s\n", err)
		os.Exit(1)
	}

	compressed, err := zlibCompress([]byte(serialized))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compressing data: %s\n", err)
		os.Exit(1)
	}

	if err = writeToSHA1(sha1, compressed); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		os.Exit(1)
	}

	fmt.Print(sha1)
}

func lsTree(sha1 string) {
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

	gitObj, err := parseObjectContent(decompressed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing object: %s\n", err)
		os.Exit(1)
	}

	if gitObj.getType() != "tree" {
		fmt.Fprintf(os.Stderr, "Object is not a blob\n")
		os.Exit(1)
	}

	for _, node := range gitObj.(tree).nodes {
		fmt.Println(node.name)
	}
}
