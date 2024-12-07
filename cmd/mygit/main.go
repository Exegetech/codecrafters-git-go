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

	case "write-tree":
		writeTree()

	case "commit-tree":
		commitTree(os.Args[2], os.Args[4], os.Args[6])

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

func catFile(sha1Hex string) {
	compressed, err := readFromSHA1(sha1Hex)
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
	sha1, err := runHashBlob(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to hash blob: %s\n", err)
		os.Exit(1)
	}

	hex := fmt.Sprintf("%x", sha1)

	fmt.Print(hex)
}

func lsTree(sha1Hex string) {
	compressed, err := readFromSHA1(sha1Hex)
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

func writeTree() {
	currDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current directory: %s", err)
		os.Exit(1)
	}

	sha1, err := runHashTree(currDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write tree: %s", err)
		os.Exit(1)
	}

	hex := fmt.Sprintf("%x", sha1)
	fmt.Print(hex)
}

func commitTree(treeSha1Hex, commitSha1Hex, message string) {
	sha1, err := runHashCommit(treeSha1Hex, []string{commitSha1Hex}, message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write commit: %s", err)
		os.Exit(1)
	}

	hex := fmt.Sprintf("%x", sha1)
	fmt.Print(hex)
}
