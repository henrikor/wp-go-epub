package main

import (
	"fmt"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	inputHTML := "test.html"

	file, err := os.Open(inputHTML)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		sectionTitle := s.Text()
		fmt.Println("-------------- SECTION ----------------- \n" + sectionTitle)

		s.NextUntil("h2").Each(func(j int, s *goquery.Selection) {
			subsectionTitle := s.Text()
			fmt.Println("-------------- SUBSECTION 1 ----------------\n" + subsectionTitle)
			// s.NextUntil("h2, h3").Each(func(k int, s *goquery.Selection) {
			// 	title := s.Text()
			// 	fmt.Println("---------------- SUBSECTION 2 / 3 -----------------\n" + title)
			// })
		})
	})

}
