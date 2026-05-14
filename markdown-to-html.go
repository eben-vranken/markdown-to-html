package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	files := findFiles()

	fmt.Println("Files found:")
	for i, e := range files {
		fmt.Print(i)
		fmt.Println(":", e.Name())
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("File to generate: ")
	index, err := reader.ReadString('\n')

	handleErr(err)

	index_as_int, err := strconv.Atoi(strings.TrimSpace(index))

	handleErr(err)

	if index_as_int < 0 || index_as_int > len(files) - 1 {
		log.Fatal("Index out of bounds!")
	}

	generateHtml(files[index_as_int])
}

func findFiles() []os.DirEntry {
	entries, err := os.ReadDir("./in")

	handleErr(err)

	return entries
}

func generateHtml(file os.DirEntry) {
	fmt.Println("Generating file:", file.Name())
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}