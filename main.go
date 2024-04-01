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

	// Set the author
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

	re := regexp.MustCompile(`(?s)<!--\s*wp:heading\s*-->(.*?)<!--\s*wp:heading\s*-->`)
	matches := re.FindAllStringSubmatch(ncontent, -1)
	i := 0
	for _, match := range matches {
		heading := fmt.Sprintf("test %v", i)
		_, err := e.AddSection(match[i], heading, "", "")
		if err != nil {
			log.Fatal(err)
		}
	}

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EPUB created successfully.")
}
