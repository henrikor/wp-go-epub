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

## Convert to kepub:
To get popup footnotes to be working on Kobo readers, we must convert to kepub (using -kepub flag). This requires that you download kepubify from: https://pgaskin.net/kepubify/dl/, renames the binary to "kepubify", set x bit on it (ie: "chmod 755 kepubify"), and keep that file in same folder as the wp-go-epub binary.

Example of using:
go run main.go -author "Proletarian Perspectives" -title "Party theory" -epubfile "test03.epub" -epubfolder . -wpfile ../tmp/test02.html -wpfolder . -logdir ./log -br -kepub

## Check epub:
java -jar epubcheck-4.2.5/epubcheck.jar ./test01.epub