package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-shiori/go-epub"
	"github.com/gookit/color"
)

var (
	footnoteFilePath = "footnotes.xhtml"
	colorRed         = color.FgRed.Render
	colorBlue        = color.FgBlue.Render
	colorYellow      = color.FgYellow.Render
	colorGreen       = color.Green.Render
	logger           *log.Logger
)

func main() {

	author, title, wpFile, epubFile, wpFolder, epubFolder, headingType, removeBr, logdir := manageFlag()
	setupLogging(*logdir)

	e, err := createEpub(*title, *author)
	if err != nil {
		logger.Fatalf("Failed to create EPUB: %v", err)
	}

	cssFilePath, err := createCSSFile()
	if err != nil {
		logger.Fatalf("Error writing CSS file: %v", err)
	}
	defer os.Remove(cssFilePath)

	cssPath, err := e.AddCSS(cssFilePath, "")
	if err != nil {
		logger.Fatalf("Error adding CSS: %v", err)
	}

	wpFilePath := filepath.Join(*wpFolder, *wpFile)
	epubFilePath := filepath.Join(*epubFolder, *epubFile)

	content, err := os.ReadFile(wpFilePath)
	if err != nil {
		logger.Fatalf("Error reading file: %v", err)
	}

	ncontent := prepareContent(string(content), *removeBr)

	footnotes := processContent(ncontent, e, cssPath, *headingType, "h3", "h4", "h5", "h6")

	if footnotes != "" {
		_, err := e.AddSection(footnotes, "Footnotes", footnoteFilePath, "")
		logger.Printf("Footnotes section: %s\n", colorBlue(footnoteFilePath))
		if err != nil {
			logger.Fatalf("Error adding footnotes: %v", err)
		}
	}

	err = e.Write(epubFilePath)
	if err != nil {
		logger.Fatalf("Error writing EPUB: %v", err)
	}
	logger.Println("EPUB created successfully.")
}

func setupLogging(logDir string) {
	// Get current weekday
	weekday := time.Now().Weekday().String()

	// Create log directory if it doesn't exist
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, os.ModePerm)
	}

	// Define log file path
	logFilePath := filepath.Join(logDir, fmt.Sprintf("%s.log", weekday))

	// Check if log file exists and its modification time
	var logFile *os.File
	if stat, err := os.Stat(logFilePath); err == nil {
		if time.Since(stat.ModTime()).Hours() > 24 {
			// Overwrite if older than one day
			logFile, err = os.Create(logFilePath)
			if err != nil {
				log.Fatalf("Failed to create log file: %v", err)
			}
		} else {
			// Append if modified within the last day
			logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if err != nil {
				log.Fatalf("Failed to open log file: %v", err)
			}
		}
	} else {
		// Create new log file if it doesn't exist
		logFile, err = os.Create(logFilePath)
		if err != nil {
			log.Fatalf("Failed to create log file: %v", err)
		}
	}

	// Set up the logger
	logger = log.New(logFile, "", log.LstdFlags)
}

func createEpub(title, author string) (*epub.Epub, error) {
	e, err := epub.NewEpub(title)
	if err != nil {
		return nil, fmt.Errorf("failed to create EPUB: %w", err)
	}
	e.SetAuthor(author)
	e.SetTitle(title)
	return e, nil
}

func createCSSFile() (string, error) {
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
	err := os.WriteFile(cssFilePath, []byte(css), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing CSS file: %w", err)
	}
	return cssFilePath, nil
}

func prepareContent(content string, removeBr bool) string {
	ncontent := strings.Replace(content, "&nbsp;", " ", -1)
	ncontent = removePTags(ncontent)
	if removeBr {
		ncontent = removeBrElements(ncontent)
	}
	return removeExtraLineBreaks(ncontent)
}

func processContent(content string, e *epub.Epub, cssPath, headingType string, subheadingTypes ...string) string {
	matches := findMatches(content, headingType)

	sections, _ := extractSections(content, matches, true)
	if len(sections) == 0 {
		logger.Printf("No <%s> tags found in the text.\n", headingType)
		return ""
	}
	rehh := compileHeadingRegex(headingType)
	var footnotes strings.Builder
	footnoteCount := 1

	for nr, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}
		h, txt := fixHeading(section, rehh)
		logger.Printf("%s nr %v: %s\n", colorRed("------------------ Section"), nr, h)

		var sectionFootnotes string
		txt, sectionFootnotes = replaceFootnotes(txt, &footnoteCount)

		var appendTxt string
		if nr == 0 {
			appendTxt = txt
			logger.Printf("txt: %s\n", txt)
		} else {
			appendTxt = getOnlyStartSection(txt)
		}
		sectionID, _ := e.AddSection(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, appendTxt), h, "", "")
		logger.Printf("sectionID: %s\n", colorYellow(sectionID))

		processSubsectionsRecursively(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, txt), sectionID, e, cssPath, h, subheadingTypes...)

		footnotes.WriteString(sectionFootnotes)
	}

	return footnotes.String()
}

func getOnlyStartSection(htmlContent string) string {
	// Regex for å matche første overskrift og all tekst fram til neste overskrift eller slutten av strengen
	re := regexp.MustCompile(`(?s)(<h[1-6][^>]*>.*?</h[1-6]>.*?)(<h[1-6][^>]*>|$)`)
	matches := re.FindStringSubmatch(htmlContent)

	if len(matches) > 1 {
		return matches[1]
	}
	// Hvis vi ikke finner en andre overskrift, returner all teksten
	reAll := regexp.MustCompile(`(?s)(<h[1-6][^>]*>.*?</h[1-6]>.*)`)
	matchesAll := reAll.FindStringSubmatch(htmlContent)
	if len(matchesAll) > 1 {
		return matchesAll[1]
	}
	return ""
}

func processSubsectionsRecursively(content string, parentSectionID string, e *epub.Epub, cssPath string, previousHeading string, subheadingTypes ...string) {
	if len(subheadingTypes) == 0 {
		return
	}

	subheadingType := subheadingTypes[0]
	remainingSubheadingTypes := subheadingTypes[1:]

	matches := findMatches(content, subheadingType)
	if len(matches) == 0 {
		return
	}

	subsections, _ := extractSections(content, matches, false)
	rehh := compileHeadingRegex(subheadingType)

	for nr, subsection := range subsections {
		if strings.TrimSpace(subsection) == "" || nr == 0 {
			continue
		}
		h, txt := fixHeading(subsection, rehh)
		logger.Printf("%s: %v nr: %s\n", colorYellow("------------------ SubSection"), nr, h)
		if h == "" && strings.Contains(txt, previousHeading) {
			logger.Printf("WARNING: SKIPPING: %s\n", h)
		}

		txt, _ = replaceFootnotes(txt, new(int))
		appendTxt := getOnlyStartSection(txt)

		subsectionID, _ := e.AddSubSection(parentSectionID, fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, appendTxt), h, "", "")
		logger.Printf("subsectionID: %s\n", colorGreen(subsectionID))
		processSubsectionsRecursively(fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="%s"/>%s`, cssPath, txt), subsectionID, e, cssPath, h, remainingSubheadingTypes...)
	}
}

func findMatches(content, headingType string) [][]int {
	reh := regexp.MustCompile(fmt.Sprintf(`(?s)<%s.*?>.*?</%s>`, headingType, headingType))
	return reh.FindAllStringIndex(content, -1)
}

func extractSections(content string, matches [][]int, addTxtWithoutHeading bool) ([]string, string) {
	var firstSection string
	var noHeadingTxt string
	if addTxtWithoutHeading && len(matches) > 0 {
		noHeadingTxt = content[:matches[0][0]]
		logger.Printf("noHeadingTxt %s", noHeadingTxt)
	}
	sections := make([]string, 0, len(matches)+1)
	if addTxtWithoutHeading && strings.Contains(noHeadingTxt, `\w`) {
		sections = append(sections, noHeadingTxt)
	}

	lastIndex := 0
	for x, match := range matches {
		if x == 0 {
			firstSection = content[lastIndex:match[0]]
		}
		sections = append(sections, content[lastIndex:match[0]])
		lastIndex = match[0]
	}
	sections = append(sections, content[lastIndex:])
	return sections, firstSection
}

func compileHeadingRegex(headingType string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`<%s.*?>(.*?)</%s>`, headingType, headingType))
}

func fixHeading(section string, rehh *regexp.Regexp) (string, string) {
	matches := rehh.FindStringSubmatch(section)

	var heading string
	if len(matches) > 0 {
		heading = matches[1]
		heading = removeHTMLTags(heading)
	} else {
		logger.Println("No match found")
	}
	return heading, section
}

func replaceFootnotes(input string, footnoteCount *int) (string, string) {
	re := regexp.MustCompile(`\(\((.*?)\)\)`)
	var footnotes strings.Builder

	output := re.ReplaceAllStringFunc(input, func(match string) string {
		footnoteID := fmt.Sprintf("footnote_%d", *footnoteCount)
		footnoteText := match[2 : len(match)-2]

		footnotes.WriteString(fmt.Sprintf(`<p id="%s">%d: %s</p>`, footnoteID, *footnoteCount, footnoteText))

		(*footnoteCount)++

		return fmt.Sprintf(
			`<sup><a href="#%s" epub:type="noteref" id="ref_%d">%d</a></sup><aside id="%s" epub:type="footnote">%s</aside>`,
			footnoteID, *footnoteCount-1, *footnoteCount-1, footnoteID, footnoteText,
		)
	})

	return output, footnotes.String()
}

func removePTags(input string) string {
	re := regexp.MustCompile(`\(\((.*?)\)\)`)

	replacer := func(match string) string {
		content := match[2 : len(match)-2]
		cleanedContent := strings.ReplaceAll(content, "<p.*?>", "")
		cleanedContent = strings.ReplaceAll(cleanedContent, "</p.*?>", "")
		return "((" + cleanedContent + "))"
	}

	return re.ReplaceAllStringFunc(input, replacer)
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

func manageFlag() (*string, *string, *string, *string, *string, *string, *string, *bool, *string) {
	author := flag.String("author", "", "the author of the EPUB")
	title := flag.String("title", "", "the title of the EPUB")
	wpFile := flag.String("wpfile", "", "the name of the file to be added as a section")
	epubFile := flag.String("epubfile", "", "the name of the file to be added as a section")
	wpFolder := flag.String("wpfolder", "", "the path to a folder containing the file")
	epubFolder := flag.String("epubfolder", "", "the path to a folder containing the file")
	headingType := flag.String("headingtype", "h2", "the type of heading to look for (e.g., h2, h3, h4)")
	logdir := flag.String("logdir", "", "Path to the log directory")
	removeBr := flag.Bool("br", false, "remove <br> elements from the content")
	flag.Parse()

	if *author == "" || *title == "" || *wpFolder == "" || *wpFile == "" || *epubFolder == "" || *epubFile == "" || *logdir == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return author, title, wpFile, epubFile, wpFolder, epubFolder, headingType, removeBr, logdir
}
