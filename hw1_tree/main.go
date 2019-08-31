package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

const LastPrefix = "└───"
const NoLastPrefix = "├───"
const DepthPrefix = "│	"
const EmptySpacePrefix = "	"

type FileWithStat struct {
	Name string
	Stat os.FileInfo
	Path string
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

	for _, name := range getAllFiles(path, printFiles, "") {
		out.Write([]byte(name))
		out.Write([]byte("\n"))
	}

	return nil
}

func getAllFiles(path string, printFiles bool, parentPrefix string) []string {
	var result []string

	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	stat, err := dir.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if !stat.IsDir() {
		return result
	}

	names, err := dir.Readdirnames(0)
	if err != nil {
		log.Fatal(err)
	}

	sort.Strings(names)

	var items []FileWithStat
	for _, name := range names {
		filepath := path + string(os.PathSeparator) + name
		stat, err := os.Stat(filepath)
		if err != nil {
			log.Fatal(err)
		}

		if stat.IsDir() || printFiles {
			items = append(items, FileWithStat{
				Name: name,
				Stat: stat,
				Path: filepath,
			})
		}
	}

	for i, file := range items {

		isLastNode := i == len(items)-1

		var nextPrefix string
		var selfPrefix string
		if isLastNode {
			selfPrefix = LastPrefix
			nextPrefix = EmptySpacePrefix
		} else {
			selfPrefix = NoLastPrefix
			nextPrefix = DepthPrefix
		}

		renderStr := parentPrefix + selfPrefix + file.Name

		if printFiles && !file.Stat.IsDir() {
			size := file.Stat.Size()
			if size > 0 {
				renderStr += fmt.Sprintf(" (%db)", size)
			} else {
				renderStr += " (empty)"
			}
		}

		result = append(result, renderStr)

		next := getAllFiles(file.Path, printFiles, parentPrefix+nextPrefix)
		result = append(result, next...)

	}

	return result
}
