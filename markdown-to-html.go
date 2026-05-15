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
	"unicode"
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

	if index_as_int < 0 || index_as_int > len(files)-1 {
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

	f, err := os.Open(fileLocation)
	r := bufio.NewReader(f)
	handleErr(err)

	outputFileLocation := strings.TrimPrefix(fileLocation, "in/")
	outputFileFormat := strings.TrimSuffix(outputFileLocation, ".md") + ".html"
	outputFile, err := os.Create("out/" + outputFileFormat)
	fmt.Println("Generating file:", "out/"+outputFileFormat)

	handleErr(err)

	defer outputFile.Close()

	inParagraph := false
	inUnorderedList := false
	inOrderedList := false
	inCodeBlock := false

	for {
		line, err := r.ReadString('\n')

		if err != nil {
			if errors.Is(err, io.EOF) {
				if len(line) > 0 {
					processLine(line, &inParagraph, &inUnorderedList, &inOrderedList, &inCodeBlock, outputFile)
				}

				if inParagraph {
					fmt.Fprint(outputFile, "</p>")
				}
				break
			}

			break
		}

		processLine(line, &inParagraph, &inUnorderedList, &inOrderedList, &inCodeBlock, outputFile)
	}
}

func parseLine(line string) string {
	// Headings
	if strings.HasPrefix(line, "######") {
		content := strings.TrimPrefix(line, "######")
		return "<h6>" + strings.TrimSpace(parseInline(content)) + "</h6>"
	}

	if strings.HasPrefix(line, "#####") {
		content := strings.TrimPrefix(line, "#####")
		return "<h5>" + strings.TrimSpace(parseInline(content)) + "</h5>"
	}

	if strings.HasPrefix(line, "####") {
		content := strings.TrimPrefix(line, "####")
		return "<h4>" + strings.TrimSpace(parseInline(content)) + "</h4>"
	}

	if strings.HasPrefix(line, "###") {
		content := strings.TrimPrefix(line, "###")
		return "<h3>" + strings.TrimSpace(parseInline(content)) + "</h3>"
	}

	if strings.HasPrefix(line, "##") {
		content := strings.TrimPrefix(line, "##")
		return "<h2>" + strings.TrimSpace(parseInline(content)) + "</h2>"
	}

	if strings.HasPrefix(line, "#") {
		content := strings.TrimPrefix(line, "#")
		return "<h1>" + strings.TrimSpace(parseInline(content)) + "</h1>"
	}

	return parseInline(line)
}

func processLine(line string, inParagraph *bool, inUnorderedList *bool, inOrderedList *bool, inCodeBlock *bool, outputFile *os.File) {
	if *inCodeBlock && !strings.HasPrefix(strings.TrimSpace(line), "```") {
		fmt.Fprint(outputFile, line)
		return
	}

	if strings.HasPrefix(strings.TrimSpace(line), "```") {
		if !*inCodeBlock {
			*inCodeBlock = true
			fmt.Fprintln(outputFile, "<pre><code>")
		} else {
			*inCodeBlock = false
			fmt.Fprintln(outputFile, "</code></pre>")
		}

		return
	}

	if strings.TrimSpace(line) == "---" {
		if *inParagraph {
			fmt.Fprint(outputFile, "</p>")
			*inParagraph = false
		}

		fmt.Fprintln(outputFile, "<hr>")
	} else if len(strings.TrimSpace(line)) >= 2 && strings.TrimSpace(line)[0] == '>' {
		quote := strings.TrimPrefix(strings.TrimSpace(line), ">")

		fmt.Fprintln(outputFile, "<blockquote>"+strings.TrimSpace(parseInline(quote))+"</blockquote>")
	} else if len(strings.TrimSpace(line)) >= 2 && unicode.IsDigit(rune(strings.TrimSpace(line)[0])) && strings.TrimSpace(line)[1] == '.' {
		if !*inOrderedList {
			fmt.Fprintln(outputFile, "<ol>")
			*inOrderedList = true
		}

		trimmedListItem := strings.Split(strings.TrimSpace(line), ".")[1]

		fmt.Fprintln(outputFile, "<li>"+strings.TrimSpace(parseInline(trimmedListItem))+"</li>")
	} else if *inOrderedList {
		*inOrderedList = false
		fmt.Fprintln(outputFile, "</ol>")
		processLine(line, inParagraph, inUnorderedList, inOrderedList, inCodeBlock, outputFile)
	} else if strings.HasPrefix(strings.TrimSpace(line), "*") || strings.HasPrefix(strings.TrimSpace(line), "-") {
		if !*inUnorderedList {
			fmt.Fprintln(outputFile, "<ul>")
			*inUnorderedList = true
		}

		trimmedListItem := ""
		if strings.HasPrefix(strings.TrimSpace(line), "*") {
			trimmedListItem = strings.TrimPrefix(strings.TrimSpace(line), "*")
		} else {
			trimmedListItem = strings.TrimPrefix(strings.TrimSpace(line), "-")
		}

		fmt.Fprintln(outputFile, "<li>"+strings.TrimSpace(parseInline(trimmedListItem))+"</li>")
	} else if *inUnorderedList {
		*inUnorderedList = false
		fmt.Fprintln(outputFile, "</ul>")
		processLine(line, inParagraph, inUnorderedList, inOrderedList, inCodeBlock, outputFile)
	} else if strings.TrimSpace(line) == "" {
		if *inParagraph {
			fmt.Fprint(outputFile, "</p>")
			*inParagraph = false
		}
	} else if strings.HasPrefix(line, "#") {
		if *inParagraph {
			fmt.Fprint(outputFile, "</p>")
			*inParagraph = false
		}
		fmt.Fprintln(outputFile, parseLine(line))
	} else {
		if !*inParagraph {
			fmt.Fprint(outputFile, "<p>")
			*inParagraph = true
		}
		fmt.Fprint(outputFile, parseInline(line))
	}
}

func parseInline(line string) string {
	for strings.Contains(line, "`") {
		line = strings.Replace(line, "`", "<code>", 1)
		line = strings.Replace(line, "`", "</code>", 1)
	}

	for strings.Contains(line, "***") {
		line = strings.Replace(line, "***", "<strong><em>", 1)
		line = strings.Replace(line, "***", "</em></strong>", 1)
	}

	for strings.Contains(line, "**") {
		line = strings.Replace(line, "**", "<strong>", 1)
		line = strings.Replace(line, "**", "</strong>", 1)
	}

	for strings.Contains(line, "*") {
		line = strings.Replace(line, "*", "<em>", 1)
		line = strings.Replace(line, "*", "</em>", 1)
	}

	for strings.Contains(line, "![") {
		linkTextStart := strings.Index(line, "![")
		linkTextEnd := strings.Index(line, "](")
		urlEnd := strings.Index(line, ")")

		linkText := line[linkTextStart+2 : linkTextEnd]
		linkUrl := line[linkTextEnd+2 : urlEnd]

		line = line[:linkTextStart] + "<img src=\"" + linkUrl + "\" alt=\"" + linkText + "\">" + line[urlEnd+1:]
	}

	for strings.Contains(line, "](") {
		linkTextStart := strings.Index(line, "[")
		linkTextEnd := strings.Index(line, "](")
		urlEnd := strings.Index(line, ")")

		linkText := line[linkTextStart+1 : linkTextEnd]
		linkUrl := line[linkTextEnd+2 : urlEnd]

		line = line[:linkTextStart] + "<a href=\"" + linkUrl + "\">" + linkText + "</a>" + line[urlEnd+1:]
	}

	return line
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
