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
	author, title, wpFile, epubFile, wpFolder, epubFolder, headingType := manageFlag()

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

	// Process content based on the specified heading type
	processContent(ncontent, e, *headingType, "h3")

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatalf("Error writing EPUB: %v", err)
	}
	fmt.Println("EPUB created successfully.")
}

func processContent(content string, e *epub.Epub, headingType string, subheadingType string) {
	// Find all occurrences of specified heading tags and their positions
	reh := regexp.MustCompile(fmt.Sprintf(`(?s)<%s.*?>.*?</%s>`, headingType, headingType))
	matches := reh.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		fmt.Printf("No <%s> tags found in the text.\n", headingType)
		return
	}

	// Extract content between specified heading tags
	sections := make([]string, 0, len(matches)+1)
	lastIndex := 0
	for _, match := range matches {
		sections = append(sections, content[lastIndex:match[0]])
		lastIndex = match[0]
	}
	// Add the remaining content after the last heading tag
	sections = append(sections, content[lastIndex:])

	// Compile regex for extracting text within specified heading tags
	rehh := regexp.MustCompile(fmt.Sprintf(`<%s.*?>(.*?)</%s>`, headingType, headingType))

	// Loop through sections and process each one
	for _, section := range sections {
		// Skip empty sections
		if strings.TrimSpace(section) == "" {
			continue
		}
		h, txt := fixHeading(section, rehh)

		// Add the main section
		sectionID, _ := e.AddSection(txt, h, "", "")

		// Process subsections based on subheadingType
		subsections := processSubsections(txt, subheadingType)
		for _, subsection := range subsections {
			sh, stxt := fixHeading(subsection, regexp.MustCompile(fmt.Sprintf(`<%s.*?>(.*?)</%s>`, subheadingType, subheadingType)))
			if strings.TrimSpace(sh) != "" {
				e.AddSubSection(sectionID, stxt, sh, "", "")
			}
		}
	}
}

func processSubsections(content string, subheadingType string) []string {
	// Find all occurrences of specified subheading tags and their positions
	reh := regexp.MustCompile(fmt.Sprintf(`(?s)<%s.*?>.*?</%s>`, subheadingType, subheadingType))
	matches := reh.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		return []string{content}
	}

	// Extract content between specified subheading tags
	subsections := make([]string, 0, len(matches)+1)
	lastIndex := 0
	for _, match := range matches {
		subsections = append(subsections, content[lastIndex:match[0]])
		lastIndex = match[0]
	}
	// Add the remaining content after the last subheading tag
	subsections = append(subsections, content[lastIndex:])

	return subsections
}

func fixHeading(section string, rehh *regexp.Regexp) (string, string) {
	// Find heading content
	matches := rehh.FindStringSubmatch(section)
	var heading string
	if len(matches) > 0 {
		heading = matches[1]
		fmt.Println(heading)
	} else {
		fmt.Println("No match found")
	}
	return heading, section
}

func manageFlag() (*string, *string, *string, *string, *string, *string, *string) {
	author := flag.String("author", "", "the author of the EPUB")
	title := flag.String("title", "", "the title of the EPUB")
	wpFile := flag.String("wpfile", "", "the name of the file to be added as a section")
	epubFile := flag.String("epubfile", "", "the name of the file to be added as a section")
	wpFolder := flag.String("wpfolder", "", "the path to a folder containing the file")
	epubFolder := flag.String("epubfolder", "", "the path to a folder containing the file")
	headingType := flag.String("headingtype", "h2", "the type of heading to look for (e.g., h2, h3, h4)")
	flag.Parse()

	if *author == "" || *title == "" || *wpFolder == "" || *wpFile == "" || *epubFolder == "" || *epubFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return author, title, wpFile, epubFile, wpFolder, epubFolder, headingType
}
