package main

import (
	"bytes"
	"testing"
)

const testFullResult = `├───project
│	├───file.txt (19b)
│	└───gopher.png (70372b)
├───static
│	├───a_lorem
│	│	├───dolor.txt (empty)
│	│	├───gopher.png (70372b)
│	│	└───ipsum
│	│		└───gopher.png (70372b)
│	├───css
│	│	└───body.css (28b)
│	├───empty.txt (empty)
│	├───html
│	│	└───index.html (57b)
│	├───js
│	│	└───site.js (10b)
│	└───z_lorem
│		├───dolor.txt (empty)
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
├───zline
│	├───empty.txt (empty)
│	└───lorem
│		├───dolor.txt (empty)
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
└───zzfile.txt (empty)
`

func TestTreeFull(t *testing.T) {
	out := new(bytes.Buffer)
	err := dirTree(out, "testdata", true)
	if err != nil {
		t.Errorf("test for OK Failed - error")
	}
	result := out.String()
	if result != testFullResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testFullResult)
	}
}

const testDirResult = `├───project
├───static
│	├───a_lorem
│	│	└───ipsum
│	├───css
│	├───html
│	├───js
│	└───z_lorem
│		└───ipsum
└───zline
	└───lorem
		└───ipsum
`

func TestTreeDir(t *testing.T) {
	out := new(bytes.Buffer)
	err := dirTree(out, "testdata", false)
	if err != nil {
		t.Errorf("test for OK Failed - error")
	}
	result := out.String()
	if result != testDirResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testDirResult)
	}
}

func TestTree(t *testing.T) {
	node1 := TreeNode{
		Parent: nil,
		Str:    "node 1",
	}
	node2 := TreeNode{
		Parent: nil,
		Str:    "node 2",
	}
	node3 := TreeNode{
		Parent: &node1,
		Str:    "node 3",
	}
	node4 := TreeNode{
		Parent: &node3,
		Str:    "node 4",
	}
	node5 := TreeNode{
		Parent: &node2,
		Str:    "node 5",
	}

	stubTree := TreeNodeList{node1, node2, node3, node4, node5}

	t.Run("test checking depth of tree node", func(t *testing.T) {

		testTable := []struct {
			node  TreeNode
			depth int
		}{
			{node1, 0},
			{node2, 0},
			{node3, 1},
			{node4, 2},
			{node5, 1},
		}

		for _, test := range testTable {
			got := stubTree.GetDepth(test.node)

			if got != test.depth {
				t.Errorf("got %d, want %d", got, test.depth)
			}
		}
	})
}
