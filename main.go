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

	// Add CSS to the EPUB
	css := `
		body {
			white-space: pre-wrap; /* ensures that white space is preserved, but wraps text */
		}
	`
	cssFilePath := "styles.css"
	err = ioutil.WriteFile(cssFilePath, []byte(css), 0644)
	if err != nil {
		log.Fatalf("Error writing CSS file: %v", err)
	}

	cssPath, err := e.AddCSS(cssFilePath, "")
	if err != nil {
		log.Fatalf("Error adding CSS: %v", err)
	}

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

	// Remove unnecessary line breaks
	ncontent = removeExtraLineBreaks(ncontent)

	// Process content based on the specified heading type
	processContent(ncontent, e, cssPath, *headingType, "h3", "h4", "h5", "h6")

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatalf("Error writing EPUB: %v", err)
	}
	fmt.Println("EPUB created successfully.")

	// Clean up the temporary CSS file
	err = os.Remove(cssFilePath)
	if err != nil {
		log.Printf("Warning: Unable to remove temporary CSS file: %v", err)
	}
}

func processContent(content string, e *epub.Epub, cssPath, headingType string, subheadingTypes ...string) {
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
		sectionID, _ := e.AddSection(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, txt), h, "", "")

		// Process subsections recursively
		processSubsectionsRecursively(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, txt), sectionID, e, cssPath, subheadingTypes...)
	}
}

func processSubsectionsRecursively(content string, parentSectionID string, e *epub.Epub, cssPath string, subheadingTypes ...string) {
	if len(subheadingTypes) == 0 {
		return
	}

	subheadingType := subheadingTypes[0]
	remainingSubheadingTypes := subheadingTypes[1:]

	// Find all occurrences of specified subheading tags and their positions
	reh := regexp.MustCompile(fmt.Sprintf(`(?s)<%s.*?>.*?</%s>`, subheadingType, subheadingType))
	matches := reh.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		return
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

	// Compile regex for extracting text within specified subheading tags
	rehh := regexp.MustCompile(fmt.Sprintf(`<%s.*?>(.*?)</%s>`, subheadingType, subheadingType))

	// Loop through subsections and process each one
	for _, subsection := range subsections {
		// Skip empty subsections
		if strings.TrimSpace(subsection) == "" {
			continue
		}
		sh, stxt := fixHeading(subsection, rehh)
		if strings.TrimSpace(sh) != "" {
			subSectionID, _ := e.AddSubSection(parentSectionID, fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, stxt), sh, "", "")
			// Recursively process further sub-subsections
			processSubsectionsRecursively(stxt, subSectionID, e, cssPath, remainingSubheadingTypes...)
		}
	}
}

func fixHeading(section string, rehh *regexp.Regexp) (string, string) {
	// Find heading content
	matches := rehh.FindStringSubmatch(section)
	var heading string
	if len(matches) > 0 {
		heading = matches[1]
		heading = removeHTMLTags(heading)
		fmt.Println(heading)
	} else {
		fmt.Println("No match found")
	}
	return heading, section
}

func removeHTMLTags(input string) string {
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(input, "")
}

func removeExtraLineBreaks(input string) string {
	// Remove line breaks that are not inside HTML tags
	re := regexp.MustCompile(`(?s)(>)(\n|\r|\r\n)(<)`)
	return re.ReplaceAllString(input, "$1$3")
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
