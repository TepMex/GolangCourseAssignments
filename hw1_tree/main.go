package main

import (
	"io"
	"log"
	"os"
	"strings"
)

const SymbolLineH = "─"
const SymbolLineV = "│"
const SymbolBranch = "├"
const SymbolAngle = "└"

type TreeNode struct {
	Parent *TreeNode
	Str    string
}

type TreeNodeList []TreeNode

func (t *TreeNodeList) GetDepth(node TreeNode) int {
	if node.Parent == nil {
		return 0
	} else if node.Parent.Parent == nil {
		return 1
	} else {
		return 1 + t.GetDepth(*node.Parent)
	}
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	var sb strings.Builder

	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	names, err := dir.Readdirnames(0)
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range names {
		sb.WriteString(name)
		sb.WriteString("\n")
	}

	out.Write([]byte(sb.String()))
	return nil
}
