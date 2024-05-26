package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-shiori/go-epub"
)

func main() {
	// Parse command line flags
	author, title, wpFile, epubFile, wpFolder, epubFolder := manageFlag()

	// Create a new EPUB
	e, err := epub.NewEpub(*title)
	if err != nil {
		log.Fatalf("Error creating EPUB: %v", err)
	}
	e.SetAuthor(*author)
	e.SetTitle(*title)

	// Get the path of the file
	wpFilePath := filepath.Join(*wpFolder, *wpFile)
	epubFilePath := filepath.Join(*epubFolder, *epubFile)

	// Read the content of the file
	content, err := ioutil.ReadFile(wpFilePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Replace HTML entities
	ncontent := strings.Replace(string(content), "&nbsp;", " ", -1)

	// Find all occurrences of <h2> tags and their positions
	reh2 := regexp.MustCompile(`(?s)<h2.*?>.*?</h2>`)
	matches := reh2.FindAllStringIndex(ncontent, -1)

	if len(matches) == 0 {
		fmt.Println("No <h2> tags found in the text.")
		return
	}

	// Extract content between <h2> tags
	sections := make([]string, 0, len(matches)+1)
	lastIndex := 0
	for _, match := range matches {
		sections = append(sections, ncontent[lastIndex:match[0]])
		lastIndex = match[0]
	}
	// Add the remaining content after the last <h2> tag
	sections = append(sections, ncontent[lastIndex:])

	// Compile regex for extracting text within <h2> tags
	reh2h := regexp.MustCompile(`<h2.*?>(.*?)<\/h2>`)

	// Loop through sections and process each one
	for _, section := range sections {
		// Skip empty sections
		if strings.TrimSpace(section) == "" {
			continue
		}
		h2, txt := fixh2(section, reh2h)
		e.AddSection(h2+txt, h2, "", "")
	}

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatalf("Error writing EPUB: %v", err)
	}
	fmt.Println("EPUB created successfully.")
}

func fixh2(section string, reh2h *regexp.Regexp) (string, string) {
	// Find <h2> content
	matches := reh2h.FindStringSubmatch(section)
	var h2 string
	if len(matches) > 0 {
		h2 = matches[1]
		fmt.Println(h2)
	} else {
		fmt.Println("No match found")
	}
	return h2, section
}

func manageFlag() (*string, *string, *string, *string, *string, *string) {
	author := flag.String("author", "", "the author of the EPUB")
	title := flag.String("title", "", "the title of the EPUB")
	wpFile := flag.String("wpfile", "", "the name of the file to be added as a section")
	epubFile := flag.String("epubfile", "", "the name of the file to be added as a section")
	wpFolder := flag.String("wpfolder", "", "the path to a folder containing the file")
	epubFolder := flag.String("epubfolder", "", "the path to a folder containing the file")
	flag.Parse()

	if *author == "" || *title == "" || *wpFolder == "" || *wpFile == "" || *epubFolder == "" || *epubFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return author, title, wpFile, epubFile, wpFolder, epubFolder
}
