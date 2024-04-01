# wp-go-epub
Go app to take wordpress post code, and convert to ePub

rm -rf test01.epub; rm -rf EPUB

go run main.go -author "Hingle McCringleberry" -title "My EPUB" -epubfile "test01.epub" -epubfolder . -wpfile test.html -wpfolder .

java -jar epubcheck-4.2.5/epubcheck.jar ./test01.epub