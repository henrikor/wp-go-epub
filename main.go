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

	re := regexp.MustCompile(`(?s)<!-- post2epub:section-start -->(.*?)<!-- post2epub:section-end -->`)
	matches := re.FindAllStringSubmatch(ncontent, -1)

	reh2 := regexp.MustCompile(`<h2.*?>(.*?)<\/h2>`)
	reh3 := regexp.MustCompile(`<h3.*?>(.*?)<\/h3>`)
	for i, match := range matches {
		if i == len(matches)-1 {
			break
		}
		color.Yellow("Prints Index: %d", i)
		//color.White("Prints match %v", match)
		section := string(match[i])
		tmp := reh2.FindStringSubmatch(section)
		heading := tmp[1]
		filenameSection := fmt.Sprintf("section%05d", i)
		_, err := e.AddSection(match[i], heading, filenameSection, "")
		if err != nil {
			log.Fatal(err)
		}
		// Subsection:

		reSub := regexp.MustCompile(`(?s)<!-- post2epub:section-sub-start -->(.*?)<!-- post2epub:section-sub-end -->`)
		matchesSub := reSub.FindAllStringSubmatch(section, -1)
		for x, matchSub := range matchesSub {
			section := string(matchSub[i])
			tmp := reh3.FindStringSubmatch(section)
			heading := tmp[1]
			filenameSubSection := fmt.Sprintf("%s-sub-%05d", filenameSection, x)

			_, err := e.AddSubSection(filenameSection, matchSub[i], heading, filenameSubSection, "")
			if err != nil {
				log.Fatal(err)
			}
		}

	}

	// Write the EPUB
	err = e.Write(epubFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("EPUB created successfully.")
}
