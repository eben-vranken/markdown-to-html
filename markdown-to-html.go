package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

	fmt.Println(files[index_as_int])

	generateHtml("in/" + files[index_as_int].Name())
}

func findFiles() []os.DirEntry {
	entries, err := os.ReadDir("./in")

	handleErr(err)

	return entries
}

func generateHtml(fileLocation string) {
	fmt.Println("Generating file:", fileLocation)
	
	f, err := os.Open(fileLocation)
	r := bufio.NewReader(f)
	handleErr(err)

	for {
		line, err := r.ReadString('\n')

		if err != nil {
			if errors.Is(err, io.EOF) {
				if len(line) > 0 {
					fmt.Println(parseLine(line))
				}
				break
			}

			break
		}

		fmt.Println(parseLine(line))
	}
}

func parseLine(line string) string {
	// Headings
	if strings.HasPrefix(line, "######") {
		content := strings.TrimPrefix(line, "######")
		return "<h6>" + strings.TrimSpace(content) + "</h6>"
	}

	if strings.HasPrefix(line, "#####") {
		content := strings.TrimPrefix(line, "#####")
		return "<h5>" + strings.TrimSpace(content) + "</h5>"
	}

		if strings.HasPrefix(line, "####") {
		content := strings.TrimPrefix(line, "####")
		return "<h4>" + strings.TrimSpace(content) + "</h4>"
	}

	if strings.HasPrefix(line, "###") {
		content := strings.TrimPrefix(line, "###")
		return "<h3>" + strings.TrimSpace(content) + "</h3>"
	}

		if strings.HasPrefix(line, "##") {
		content := strings.TrimPrefix(line, "##")
		return "<h2>" + strings.TrimSpace(content) + "</h2>"
	}


		if strings.HasPrefix(line, "#") {
		content := strings.TrimPrefix(line, "#")
		return "<h1>" + strings.TrimSpace(content) + "</h1>"
	}


	return ""
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}