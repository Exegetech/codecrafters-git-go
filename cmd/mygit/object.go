package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type gitObject interface {
	getType() string
	getSize() int
}

type blob struct {
	content []byte
}

func (b blob) getType() string {
	return "blob"
}

func (b blob) getSize() int {
	return len(b.content)
}

func (b blob) Bytes() []byte {
	return []byte(fmt.Sprintf("blob %d\x00%s", b.getSize(), b.content))
}

type tree struct {
	nodes []treeNode
}

type treeNode struct {
	mode string
	name string
	hash [20]byte
}

func (t tree) getType() string {
	return "tree"
}

func (t tree) getSize() int {
	content := t.getContent()
	return len(content)
}

func (t tree) getContent() []byte {
	data := []byte{}

	for _, node := range t.nodes {
		entry := fmt.Sprintf("%s %s\x00", node.mode, node.name)
		data = append(data, []byte(entry)...)
		data = append(data, node.hash[:]...)
	}

	return data
}

func (t tree) Bytes() []byte {
	data := []byte{}

	content := t.getContent()
	header := fmt.Sprintf("tree %d\x00", len(content))

	data = append(data, []byte(header)...)
	data = append(data, content...)

	return data
}

type commit struct {
	treeSha1Hex    string
	parentsSha1Hex []string
	message        string
}

func (c commit) getType() string {
	return "commit"
}

func (c commit) getSize() int {
	content := c.getContent()
	return len(content)
}

func (c commit) getContent() []byte {
	data := []byte{}

	treeSection := fmt.Sprintf("tree %s\n", c.treeSha1Hex)
	data = append(data, []byte(treeSection)...)

	for _, parent := range c.parentsSha1Hex {
		parentSection := fmt.Sprintf("parent %s\n", parent)
		data = append(data, []byte(parentSection)...)
	}

	messageSection := fmt.Sprintf("\n%s\n", c.message)
	data = append(data, []byte(messageSection)...)

	return data
}

func (c commit) Bytes() []byte {
	data := []byte{}

	content := c.getContent()
	header := fmt.Sprintf("commit %d\x00", len(content))

	data = append(data, []byte(header)...)
	data = append(data, content...)

	return data
}

func parseObjectContent(data []byte) (gitObject, error) {
	indexZeroByte := bytes.IndexByte(data, 0)
	if indexZeroByte == -1 {
		return blob{}, fmt.Errorf("Byte zero not found")
	}

	parts := bytes.Fields(data[:indexZeroByte])
	if len(parts) != 2 {
		return blob{}, fmt.Errorf("Header does not contain exactly two parts")
	}

	objectType := string(parts[0])

	size, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		return blob{}, fmt.Errorf("Invalid size: %s", err)
	}

	content := data[indexZeroByte+1:]

	if size > len(content) {
		return blob{}, fmt.Errorf("Data length beyond slice boundary")
	}

	if objectType == "blob" {
		return blob{content}, nil
	}

	if objectType == "tree" {
		nodes, err := parseTreeEntry(content)
		if err != nil {
			return tree{}, fmt.Errorf("Failed to parse tree nodes: %s", err)
		}

		return tree{
			nodes,
		}, nil
	}

	return blob{}, nil
}

func parseTreeEntry(data []byte) ([]treeNode, error) {
	var entries []treeNode

	for len(data) > 0 {
		indexZeroByte := bytes.IndexByte(data, 0)
		if indexZeroByte == -1 {
			return nil, fmt.Errorf("Byte zero not found")
		}

		parts := bytes.Fields(data[:indexZeroByte])

		if len(parts) != 2 {
			return nil, fmt.Errorf("Header does not contain exactly two parts")
		}

		mode := string(parts[0])
		name := string(parts[1])

		var hash [20]byte

		copy(hash[:], data[indexZeroByte+1:indexZeroByte+21])

		entries = append(entries, treeNode{
			mode: mode,
			name: name,
			hash: hash,
		})

		data = data[indexZeroByte+21:]
	}

	return entries, nil
}

func runHashBlob(filepath string) ([]byte, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return []byte{}, fmt.Errorf("Error reading file: %s\n", err)
	}

	b := blob{
		content: data,
	}

	serialized := b.Bytes()
	compressed, err := zlibCompress(serialized)
	if err != nil {
		return []byte{}, fmt.Errorf("Error compressing data: %s\n", err)
	}

	sha1, err := computeSHA1(serialized)
	if err != nil {
		return []byte{}, fmt.Errorf("Error computing SHA-1: %s\n", err)
	}

	if err = writeToSHA1(sha1, compressed); err != nil {
		return []byte{}, fmt.Errorf("Error writing file: %s\n", err)
	}

	return sha1, nil
}

func runHashTree(path string) ([]byte, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to list files in dir: %s", err)
	}

	treeObj := tree{
		nodes: []treeNode{},
	}

	for _, file := range files {
		name := file.Name()
		if name == ".git" {
			continue
		}

		fullPath := filepath.Join(path, name)

		if file.IsDir() == false {
			sha1, err := runHashBlob(fullPath)
			if err != nil {
				return []byte{}, fmt.Errorf("Failed to hash blob for file %s: %s", fullPath, err)
			}

			hashByte := [20]byte{}
			copy(hashByte[:], sha1)

			node := treeNode{
				mode: "100644",
				name: name,
				hash: hashByte,
			}

			treeObj.nodes = append(treeObj.nodes, node)
			continue
		}

		sha1, err := runHashTree(fullPath)
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to hash tree for dir %s: %s", fullPath, err)
		}

		hashByte := [20]byte{}
		copy(hashByte[:], sha1)

		node := treeNode{
			mode: "40000",
			name: name,
			hash: hashByte,
		}

		treeObj.nodes = append(treeObj.nodes, node)
	}

	serialized := treeObj.Bytes()
	compressed, err := zlibCompress(serialized)
	if err != nil {
		return []byte{}, fmt.Errorf("Error compressing data: %s\n", err)
	}

	sha1, err := computeSHA1(serialized)
	if err != nil {
		return []byte{}, fmt.Errorf("Error computing SHA-1: %s\n", err)
	}

	if err = writeToSHA1(sha1, compressed); err != nil {
		return []byte{}, fmt.Errorf("Error writing file: %s\n", err)
	}

	return sha1, nil
}

func runHashCommit(treeSha1Hex string, parentsCommitSha1Hex []string, message string) ([]byte, error) {
	c := commit{
		treeSha1Hex:    treeSha1Hex,
		parentsSha1Hex: parentsCommitSha1Hex,
		message:        message,
	}

	serialized := c.Bytes()
	compressed, err := zlibCompress(serialized)
	if err != nil {
		return []byte{}, fmt.Errorf("Error compressing data: %s\n", err)
	}

	sha1, err := computeSHA1(serialized)
	if err != nil {
		return []byte{}, fmt.Errorf("Error computing SHA-1: %s\n", err)
	}

	if err = writeToSHA1(sha1, compressed); err != nil {
		return []byte{}, fmt.Errorf("Error writing file: %s\n", err)
	}

	return sha1, nil
}
