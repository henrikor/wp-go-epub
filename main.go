package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-shiori/go-epub"
	"github.com/gookit/color"
	"golang.org/x/net/html"
)

var (
	footnoteFilePath = "footnotes.xhtml"
	colorRed         = color.FgRed.Render
	colorBlue        = color.FgBlue.Render
	colorYellow      = color.FgYellow.Render
	colorGreen       = color.Green.Render
)

func main() {
	// Parse command line flags
	author, title, wpFile, epubFile, wpFolder, epubFolder, headingType, removeBr := manageFlag()

	// Create a new EPUB
	e, err := epub.NewEpub(*title)
	if err != nil {
		log.Fatalf("Failed to create EPUB: %v", err)
	}
	e.SetAuthor(*author)
	e.SetTitle(*title)

	// Add CSS to the EPUB
	css := `
		body {
			font-family: Arial, sans-serif;
			line-height: 1.6;
			margin: 0;
			padding: 0;
			white-space: pre-wrap;
		}
		h1, h2, h3, h4, h5, h6 {
			font-family: 'Georgia', serif;
			line-height: 1.4;
			margin-top: 1em;
			margin-bottom: 0.5em;
		}
		h1 {
			font-size: 2em;
			border-bottom: 2px solid #000;
		}
		h2 {
			font-size: 1.75em;
			border-bottom: 1px solid #000;
		}
		h3 {
			font-size: 1.5em;
		}
		h4 {
			font-size: 1.25em;
		}
		h5 {
			font-size: 1em;
		}
		h6 {
			font-size: 0.875em;
		}
		p {
			margin: 0.5em 0;
		}
		a {
			color: #007BFF;
			text-decoration: none;
		}
		a:hover {
			text-decoration: underline;
		}
	`
	cssFilePath := "styles.css"
	err = os.WriteFile(cssFilePath, []byte(css), 0644)
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
	content, err := os.ReadFile(wpFilePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Replace HTML entities
	ncontent := strings.Replace(string(content), "&nbsp;", " ", -1)
	ncontent = removePTags(ncontent)

	// Remove <br> elements if the flag is set
	if *removeBr {
		ncontent = removeBrElements(ncontent)
	}

	// Remove unnecessary line breaks
	ncontent = removeExtraLineBreaks(ncontent)

	// Process content based on the specified heading type
	footnotes := processContent(ncontent, e, cssPath, *headingType, "h3", "h4", "h5", "h6")

	footnotes, err = cleanHTML(footnotes)
	if err != nil {
		fmt.Println("Error cleaning HTML:", err)
	}

	// Add footnotes to the end of the book

	if footnotes != "" {
		_, err := e.AddSection(footnotes, "Footnotes", footnoteFilePath, "")
		fmt.Printf("Footnotes section: %s\n", colorBlue(footnoteFilePath))

		if err != nil {
			log.Fatalf("Error adding footnotes: %v", err)
		}
	}

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

func processContent(content string, e *epub.Epub, cssPath, headingType string, subheadingTypes ...string) string {
	// Find all occurrences of specified heading tags and their positions
	reh := regexp.MustCompile(fmt.Sprintf(`(?s)<%s.*?>.*?</%s>`, headingType, headingType))
	matches := reh.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		fmt.Printf("No <%s> tags found in the text.\n", headingType)
		return ""
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

	// Initialize footnote content
	var footnotes strings.Builder
	footnoteCount := 1

	// Loop through sections and process each one
	for _, section := range sections {
		// Skip empty sections
		if strings.TrimSpace(section) == "" {
			continue
		}
		h, txt := fixHeading(section, rehh)
		var sectionFootnotes string
		txt, sectionFootnotes = replaceFootnotes(txt, &footnoteCount)
		// Fix the HTML structure in the output
		// txt, err := cleanHTML(txt)
		// if err != nil {
		// 	fmt.Println("Error cleaning HTML:", err)
		// }

		re := regexp.MustCompile("<h3.*?>.*")
		newtxt := re.ReplaceAllString(txt, "")
		// Add the main section
		sectionID, _ := e.AddSection(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, newtxt), h, "", "")
		fmt.Printf("sectionID: %s\n", colorYellow(sectionID))

		// Process subsections recursively
		processSubsectionsRecursively(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, txt), sectionID, e, cssPath, subheadingTypes...)

		// Append footnotes to the main footnotes content
		footnotes.WriteString(sectionFootnotes)
	}

	// Return the main content and footnotes
	return footnotes.String()
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

func replaceFootnotes(input string, footnoteCount *int) (string, string) {
	// Compile regex to find footnote patterns
	re := regexp.MustCompile(`\(\((.*?)\)\)`)

	// Initialize a builder for footnotes content
	var footnotes strings.Builder

	// Replace each footnote pattern with the appropriate HTML
	output := re.ReplaceAllStringFunc(input, func(match string) string {
		// Generate footnote ID
		footnoteID := fmt.Sprintf("footnote_%d", *footnoteCount)
		footnoteText := match[2 : len(match)-2] // Extract footnote text

		// Append the footnote to the footnotes builder
		footnotes.WriteString(fmt.Sprintf(`<p id="%s">%d: %s</p>`, footnoteID, *footnoteCount, footnoteText))

		// Increment the footnote count
		(*footnoteCount)++

		// Return the superscript link to the footnote and ePub3 popup
		return fmt.Sprintf(
			`<sup><a href="#%s" epub:type="noteref" id="ref_%d">%d</a></sup><aside id="%s" epub:type="footnote">%s</aside>`,
			footnoteID, *footnoteCount-1, *footnoteCount-1, footnoteID, footnoteText,
		)
	})

	// Return the text with footnote references and the footnotes content
	return output, footnotes.String()
}

// cleanHTML parses and cleans the HTML input.
func cleanHTML(input string) (string, error) {
	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %v", err)
	}

	// Serialize the cleaned HTML
	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		return "", fmt.Errorf("error rendering HTML: %v", err)
	}

	return buf.String(), nil
}
func removePTags(input string) string {
	// Compile regex to find patterns between (( and ))
	re := regexp.MustCompile(`\(\((.*?)\)\)`)

	// Replace function to remove <p> and </p> tags
	replacer := func(match string) string {
		// Extract the content between (( and ))
		content := match[2 : len(match)-2]

		// Remove <p> and </p> tags
		cleanedContent := strings.ReplaceAll(content, "<p.*?>", "")
		cleanedContent = strings.ReplaceAll(cleanedContent, "</p.*?>", "")

		// Return the cleaned content wrapped in (())
		return "((" + cleanedContent + "))"
	}

	// Apply the replacer function to all matches
	output := re.ReplaceAllStringFunc(input, replacer)

	return output
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
	i := 0
	re := regexp.MustCompile("<h2.*?>.*?<h")
	for _, match := range matches {
		i++ // Need to skip the first one, else we got h2 ..> h3 repeted in subsection
		if i == 1 {
			continue
		}
		newtxt := re.ReplaceAllString(content[lastIndex:match[0]], `<h`)
		subsections = append(subsections, newtxt)
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
		h, txt := fixHeading(subsection, rehh)
		txt, _ = replaceFootnotes(txt, new(int))
		txt, err := cleanHTML(txt)
		if err != nil {
			fmt.Println("Error cleaning HTML:", err)
		}

		// Add the subsection under the parent section
		subsectionID, _ := e.AddSubSection(parentSectionID, fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, txt), h, "", "")
		fmt.Printf("subsectionID: %s\n", colorGreen(subsectionID))
		// Process sub-subsections recursively
		processSubsectionsRecursively(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, txt), subsectionID, e, cssPath, remainingSubheadingTypes...)
	}
}

func removeBrElements(input string) string {
	re := regexp.MustCompile(`(?i)<br\s*/?>`)
	return re.ReplaceAllString(input, "")
}

func removeExtraLineBreaks(input string) string {
	re := regexp.MustCompile(`\r?\n`)
	return re.ReplaceAllString(input, " ")
}

func removeHTMLTags(input string) string {
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(input, "")
}

func manageFlag() (*string, *string, *string, *string, *string, *string, *string, *bool) {
	author := flag.String("author", "", "the author of the EPUB")
	title := flag.String("title", "", "the title of the EPUB")
	wpFile := flag.String("wpfile", "", "the name of the file to be added as a section")
	epubFile := flag.String("epubfile", "", "the name of the file to be added as a section")
	wpFolder := flag.String("wpfolder", "", "the path to a folder containing the file")
	epubFolder := flag.String("epubfolder", "", "the path to a folder containing the file")
	headingType := flag.String("headingtype", "h2", "the type of heading to look for (e.g., h2, h3, h4)")
	removeBr := flag.Bool("br", false, "remove <br> elements from the content")
	flag.Parse()

	if *author == "" || *title == "" || *wpFolder == "" || *wpFile == "" || *epubFolder == "" || *epubFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return author, title, wpFile, epubFile, wpFolder, epubFolder, headingType, removeBr
}
