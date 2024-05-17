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

	// Define regex patterns for headings
	reh2 := regexp.MustCompile(`<h2.*?>(.*?)<\/h2>`)
	reh3 := regexp.MustCompile(`<h3.*?>(.*?)<\/h3>`)
	reh4 := regexp.MustCompile(`<h4.*?>(.*?)<\/h4>`)
	reh5 := regexp.MustCompile(`<h5.*?>(.*?)<\/h5>`)
	reh6 := regexp.MustCompile(`<h6.*?>(.*?)<\/h6>`)

	// Split content into sections and subsections
	sections := splitIntoSections(ncontent, reh2, reh3, reh4, reh5, reh6)

	// Add sections and subsections to EPUB
	for _, section := range sections {
		sectionTitle := reh2.FindStringSubmatch(section.title)
		if len(sectionTitle) > 1 {
			sectionID, _ := e.AddSection(section.content, sectionTitle[1], "", "")
			for _, subsection := range section.subsections {
				subsectionTitle := getSubSectionTitle(subsection.title, reh3, reh4, reh5, reh6)
				if subsectionTitle != "" {
					e.AddSubSection(sectionID, subsection.content, subsectionTitle, "", "")
				}
			}
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
	title       string
	content     string
	subsections []SubSection
}

type SubSection struct {
	title   string
	content string
}

func splitIntoSections(content string, reh2, reh3, reh4, reh5, reh6 *regexp.Regexp) []Section {
	sections := []Section{}
	h2Matches := reh2.FindAllStringIndex(content, -1)
	lastIndex := 0

	for _, match := range h2Matches {
		if match[0] > lastIndex {
			sectionContent := content[lastIndex:match[0]]
			sectionTitle := content[match[0]:match[1]]
			section := Section{
				title:   sectionTitle,
				content: sectionContent,
			}
			sections = append(sections, section)
		}
		lastIndex = match[1]
	}
	sections = append(sections, Section{content: content[lastIndex:]})

	for i, section := range sections {
		sectionContent := section.content
		h3Matches := reh3.FindAllStringIndex(sectionContent, -1)
		subSections := []SubSection{}
		lastSubIndex := 0

		for _, match := range h3Matches {
			if match[0] > lastSubIndex {
				subSectionContent := sectionContent[lastSubIndex:match[0]]
				subSectionTitle := sectionContent[match[0]:match[1]]
				subSection := SubSection{
					title:   subSectionTitle,
					content: subSectionContent,
				}
				subSections = append(subSections, subSection)
			}
			lastSubIndex = match[1]
		}
		subSections = append(subSections, SubSection{content: sectionContent[lastSubIndex:]})
		sections[i].subsections = subSections
	}

	return sections
}

func getSubSectionTitle(title string, reh3, reh4, reh5, reh6 *regexp.Regexp) string {
	submatch3 := reh3.FindStringSubmatch(title)
	if len(submatch3) > 1 {
		return submatch3[1]
	}
	submatch4 := reh4.FindStringSubmatch(title)
	if len(submatch4) > 1 {
		return submatch4[1]
	}
	submatch5 := reh5.FindStringSubmatch(title)
	if len(submatch5) > 1 {
		return submatch5[1]
	}
	submatch6 := reh6.FindStringSubmatch(title)
	if len(submatch6) > 1 {
		return submatch6[1]
	}
	return ""
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
