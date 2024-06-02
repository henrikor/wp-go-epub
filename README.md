# Convert WordPress to epub
This repo is part of wordpress to epub conversion project. We got the following repos:

## https://github.com/henrikor/wp-epub-converter
This is the WorPress plugin. This plugin exports the global function "display_epub_link". This function returns a text with link to a epub version of the post. This can be used in your theme like this: "$link = display_epub_link();". Example of child theme is:

## https://github.com/henrikor/tortuga-child-pp
This child theme puts link to epub in div.entry-meta (with links to date, "Leave comment" etc.). This child also filter frontpage, so only post from category 19 will show up here.

## https://github.com/henrikor/wp-go-epub
This is the Go app which converts the wp post to epub. It created Table of Content (TOC) based on heading elements (starting with h2 as top element), and creats endnotes based on what you have in dubble paranteses "((like this))".

# wp-go-epub
Go app to take wordpress post code, and convert to ePub

rm -rf test01.epub; rm -rf EPUB

go run main.go -author "Hingle McCringleberry" -title "My EPUB" -epubfile "test01.epub" -epubfolder . -wpfile test.html -wpfolder .

java -jar epubcheck-4.2.5/epubcheck.jar ./test01.epub