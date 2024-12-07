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
	size    int
	content []byte
}

func (b blob) getType() string {
	return "blob"
}

func (b blob) getSize() int {
	return b.size
}

func (b blob) Bytes() []byte {
	return []byte(fmt.Sprintf("blob %d\x00%s", b.size, b.content))
}

type tree struct {
	size  int
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
	return t.size
}

func (t tree) Bytes() []byte {
	total := []byte{}

	header := fmt.Sprintf("tree %d\x00", t.size)
	total = append(total, []byte(header)...)

	for _, node := range t.nodes {
		entry := fmt.Sprintf("%s %s\x00", node.mode, node.name)
		total = append(total, []byte(entry)...)
		total = append(total, node.hash[:]...)
	}

	return total
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
		return blob{
			size,
			content,
		}, nil
	}

	if objectType == "tree" {
		nodes, err := parseTreeEntry(content)
		if err != nil {
			return tree{}, fmt.Errorf("Failed to parse tree nodes: %s", err)
		}

		return tree{
			size,
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
	stats, err := os.Stat(filepath)
	if err != nil {
		return []byte{}, fmt.Errorf("Error getting file stats: %s\n", err)
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return []byte{}, fmt.Errorf("Error reading file: %s\n", err)
	}

	b := blob{
		size:    int(stats.Size()),
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
		size:  0,
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

			// add 2 for space and null byte
			treeObj.size += len([]byte(node.mode)) + len([]byte(node.name)) + 2 + len(node.hash)
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

		// add 2 for space and null byte
		treeObj.size += len([]byte(node.mode)) + len([]byte(node.name)) + 2 + len(node.hash)
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
