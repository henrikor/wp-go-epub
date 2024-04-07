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
	author := flag.String("author", "", "the author of the EPUB")
	title := flag.String("title", "", "the title of the EPUB")
	wpFile := flag.String("wpfile", "", "the name of the file to be added as a section")
	epubFile := flag.String("epubfile", "", "the name of the file to be added as a section")
	wpFolder := flag.String("wpfolder", "", "the path to a folder containing the file")
	epubFolder := flag.String("epubfolder", "", "the path to a folder containing the file")
	flag.Parse()

	// Check if required flags are provided
	if *author == "" || *title == "" || *wpFolder == "" || *wpFile == "" || *epubFolder == "" || *epubFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create a new EPUB
	e, err := epub.NewEpub(*title)
	if err != nil {
		log.Fatal(err)
	}
	e.SetAuthor(*author)
	e.SetTitle(*title)

	// Get the path of the file
	wpFilePath := filepath.Join(*wpFolder, *wpFile)
	epubFilePath := filepath.Join(*epubFolder, *epubFile)

	// Read the content of the file
	content, err := ioutil.ReadFile(wpFilePath)
	if err != nil {
		log.Fatal(err)
	}
	ncontent := strings.Replace(string(content), "&nbsp;", " ", -1)
	ncontent = strings.ReplaceAll(ncontent, "<!-- wp:heading -->", "<!-- post2epub:section-end -->\n<!-- post2epub:section-start -->\n<!-- wp:heading -->")
	ncontent = strings.Replace(ncontent, "<!-- post2epub:section-end -->", "", 1)

	re := regexp.MustCompile(`(?s)<!-- wp:heading -->(.*?)<!-- wp:heading -->`)
	matches := re.FindAllStringSubmatch(ncontent, -1)

	// Finn siste del:
	re = regexp.MustCompile(`(?s)<!--\s*wp:heading\s*-->`)
	lastIndex := re.FindAllStringIndex(ncontent, -1)

	color.Yellow("\n\n--------------------------------------------------------\n\n")

	var contentAfterLastMatch string
	if len(lastIndex) > 0 {
		lastMatchIndex := lastIndex[len(lastIndex)-1][1]  // Slutten av siste forekomst av <!-- wp:heading -->
		contentAfterLastMatch = ncontent[lastMatchIndex:] // Alt etter siste forekomst av <!-- wp:heading -->
		// fmt.Println(contentAfterLastMatch)

	} else {
		fmt.Println("Ingen treff på <!-- wp:heading --> i teksten.")
	}
	color.Yellow("\n\n--------------------------------------------------------\n\n")
	var contentAfterLastMatchSlice []string
	contentAfterLastMatchSlice = append(contentAfterLastMatchSlice, contentAfterLastMatch)
	matches = append(matches, contentAfterLastMatchSlice)
	fmt.Printf("Matches lengde: %v", len(matches))
	// os.Exit(0)

	// reh2 := regexp.MustCompile(`<h2.*?>(.*?)<\/h2>`)
	// reh3 := regexp.MustCompile(`<h3.*?>(.*?)<\/h3>`)
	for i, match := range matches {
		color.Yellow("\n\n======================================================\n\n")
		color.Yellow("Prints Index: %d", i)
		color.Yellow("\n\n======================================================\n\n")
		for _, x := range match {
			color.Yellow("\n\n======================================================\n\n")
			color.Yellow("Prints x: %d", i)
			color.Yellow("\n\n======================================================\n\n")

			fmt.Printf("x: %v", x)
		}
	}

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EPUB created successfully.")
}
