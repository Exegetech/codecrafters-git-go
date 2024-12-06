package main

import (
	"bytes"
	"fmt"
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

func (b blob) String() string {
	return fmt.Sprintf("blob %d\x00%s", b.size, b.content)
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
