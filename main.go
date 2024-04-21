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

	"github.com/fatih/color"
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

	// Find all occurrences of <!-- wp:heading -->
	reh2 := regexp.MustCompile(`(?s)<!-- wp:heading -->(.*?)<!-- wp:heading -->`)
	matchesh2 := reh2.FindAllStringSubmatch(ncontent, -1)

	// Find the last occurrence of <!-- wp:heading -->
	lastIndex := reh2.FindAllStringIndex(ncontent, -1)

	color.Yellow("\n\n--------------------------------------------------------\n\n")

	var contentAfterLastMatch string
	if len(lastIndex) > 0 {
		lastMatchIndex := lastIndex[len(lastIndex)-1][1]  // End of last occurrence of <!-- wp:heading -->
		contentAfterLastMatch = ncontent[lastMatchIndex:] // All content after last occurrence of <!-- wp:heading -->
	} else {
		fmt.Println("No matches found for <!-- wp:heading --> in the text.")
	}
	color.Yellow("\n\n--------------------------------------------------------\n\n")

	// Append content after last match to matches slice
	matchesh2 = append(matchesh2, []string{contentAfterLastMatch})

	// Compile regex outside of loop for efficiency
	reh2h := regexp.MustCompile(`<h2.*?>(.*?)<\/h2>`)

	// Loop through matches and process each one
	for i, match := range matchesh2 {
		fixh2(match, i, reh2h)
	}

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatalf("Error writing EPUB: %v", err)
	}
	fmt.Println("EPUB created successfully.")
}

func fixh2(match []string, i int, reh2h *regexp.Regexp) {
	txt := match[len(match)-1]

	color.Yellow("\n\n==/////////////////////////////////////////////////////////////==\n\n")
	color.Yellow("$1: %s", `$1`)
	color.Yellow("\n\n/////////////////////////////////////////////////////////////====\n\n")
	color.Yellow("\n\n======================================================\n\n")
	color.Yellow("Prints Index: %d", i)
	color.Yellow("\n\n======================================================\n\n")

	h2 := reh2h.ReplaceAllString(txt, `$1`)
	color.Yellow("H2 : %s", h2)

	fmt.Println(txt)
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
