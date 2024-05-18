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
		log.Fatalf("Failed to create EPUB: %v", err)
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

	// Define regex pattern for h2 headings
	reh2 := regexp.MustCompile(`<h2.*?>(.*?)<\/h2>`)

	// Split content into sections
	sections := splitIntoSections(ncontent, reh2)

	// Add sections to EPUB
	for _, section := range sections {
		sectionTitle := reh2.FindStringSubmatch(section.title)
		if len(sectionTitle) > 1 {
			e.AddSection(section.title+section.content, sectionTitle[1], "", "")
		}
	}

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatalf("Error writing EPUB: %v", err)
	}
	fmt.Println("EPUB created successfully.")
}

type Section struct {
	title   string
	content string
}

func splitIntoSections(content string, reh2 *regexp.Regexp) []Section {
	sections := []Section{}
	h2Matches := reh2.FindAllStringIndex(content, -1)
	lastIndex := 0

	for _, match := range h2Matches {
		if match[0] > lastIndex {
			sectionTitle := content[match[0]:match[1]]
			sectionContent := content[lastIndex:match[0]]
			section := Section{
				title:   sectionTitle,
				content: sectionContent,
			}
			sections = append(sections, section)
		}
		lastIndex = match[1]
	}
	// Add the last section
	if lastIndex < len(content) {
		lastSectionContent := content[lastIndex:]
		sections = append(sections, Section{
			title:   "",
			content: lastSectionContent,
		})
	}

	return sections
}

func manageFlag() (*string, *string, *string, *string, *string, *string) {
	author := flag.String("author", "", "the author of the EPUB")
	title := flag.String("title", "", "the title of the EPUB")
	wpFile := flag.String("wpfile", "", "the name of the file to be added as a section")
	epubFile := flag.String("epubfile", "", "the name of the EPUB file to be created")
	wpFolder := flag.String("wpfolder", "", "the path to the folder containing the HTML file")
	epubFolder := flag.String("epubfolder", "", "the path to the folder where the EPUB will be saved")
	flag.Parse()

	if *author == "" || *title == "" || *wpFolder == "" || *wpFile == "" || *epubFolder == "" || *epubFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return author, title, wpFile, epubFile, wpFolder, epubFolder
}
