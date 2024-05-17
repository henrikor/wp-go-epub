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
	// Convert the designated charset HTML to utf-8 encoded HTML.
	// `charset` being one of the charsets known by the iconv package.
	// utfBody, err := iconv.NewReader(file, charset, "utf-8")
	// if err != nil {
	// 	// handler error
	// }
	// doc, err := goquery.NewDocumentFromReader(utfBody)
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		// sectionTitle, err := s.Html()
		// if err != nil {
		// 	color.RedString("Err: %v", err)
		// }
		title := s.Closest("h2")

		fmt.Println("-------------- title2 ----------------- \n" + title.Text())
		// fmt.Println("-------------- SECTION ----------------- \n" + sectionTitle)

		s.NextUntil("h2").Filter("h3").Each(func(j int, s *goquery.Selection) {
			// Print out the text of the h2 element
			title := s.Closest("h3")
			fmt.Println("-------------- title3 ----------------- \n" + title.Text())

			fmt.Println("h3 element text:", s.Prev().Text())

			// Now you can process h3 elements
		})

		// s.NextUntil("h2").Each(func(j int, s *goquery.Selection) {
		// 	sub1 := s.Text()
		// 	fmt.Println("-------------- SUBSECTION 1 ----------------\n" + sub1)

		// 	// s.NextUntil("h2, h3").Each(func(k int, s *goquery.Selection) {
		// 	// 	title := s.Text()
		// 	// 	fmt.Println("---------------- SUBSECTION 2 / 3 -----------------\n" + title)
		// 	// })
		// })
	})

}
