package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gookit/color" //https://github.com/gookit/color
	"github.com/henrikor/unzip"
)

type Config struct {
	ServiceName string
	AppHome     string
}

// Unmarshal yaml file - returns struct
func read_file(filename string) (txt string) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	txt = string(source)
	return txt
}

// Unzip ePub file
func unzipEpub(epub string, zipout string) []string {
	zipFiles, err := unzip.Unzip(epub, zipout)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Unzip ok")
	}
	if len(zipFiles) > 20 {
		fmt.Println("Unziped ", len(zipFiles), "zipfiles")
	} else {
		fmt.Println("Unzipped:\n" + strings.Join(zipFiles, "\n"))
	}
	return zipFiles
}
func fix_xml(filename string) {
	oldtxt := read_file(filename)
	matched, _ := regexp.MatchString(`footnote_plugin_reference_`, oldtxt)

	//  CSS to add:
	// .footnote_backlink{
	// 	font-family: Arial, Helvetica, sans-serif;
	// 	font-weight: 900;
	// 	border-top: double;
	//   }

	// r := regexp.MustCompile(`(?s)<span class="footnote_referrer"><a role="button" tabindex="0" onkeypress="footnote_moveToReference_\d+_\d+\('footnote_plugin_reference_(\d+_\d+_\d+)'\);"><sup id="footnote_plugin_tooltip_\d+_\d+_\d+" class="footnote_plugin_tooltip_text">\[(\d+)\].*?<span id="footnote_plugin_tooltip_text_\d+_\d+_\d+".*?<span class="footnote_tooltip_continue">Continue reading</span></span></span>`)
	r := regexp.MustCompile(`<span class="footnote_tooltip_continue">Continue reading</span>`)
	r2 := regexp.MustCompile(`(?s)<span class="footnote_referrer"><a role="button" tabindex="0" onkeypress="footnote_moveToReference_\d+_\d+\('footnote_plugin_reference_(\d+_\d+_\d+)'\);"><sup id="footnote_plugin_tooltip_\d+_\d+_\d+" class="footnote_plugin_tooltip_text">\[(\d+)\].*?<span id="footnote_plugin_tooltip_text_(\d+_\d+_\d+)".*?</span></span>`)
	r5 := regexp.MustCompile(`(?s)<span class="footnote_referrer"></span></a><a role="button" tabindex="0" onkeypress="footnote_moveToReference_\d+_\d+\('footnote_plugin_reference_(\d+_\d+_\d+)'\);"><sup id="footnote_plugin_tooltip_\d+_\d+_\d+" class="footnote_plugin_tooltip_text">\[(\d+)\].*?</li>`)
	r3 := regexp.MustCompile(`<a id="footnote_plugin_reference_(\d+_\d+_\d+)" class="footnote_backlink"><span class="footnote_index_arrow">↑</span>(\d+)(,?)</a>`)
	r4 := regexp.MustCompile(`<button class="rtoc_open_close rtoc_open"></button>`)
	r6 := regexp.MustCompile(`(?s)<table class="footnotes_table.*?(<tr class="footnotes_plugin_reference_row">.*)</tbody> </table>`)
	r7 := regexp.MustCompile(`(?s)<tr class="footnotes_plugin_reference_row">.*?(<a id="footnote_plugin_reference.*?)</td></tr>`)
	r8 := regexp.MustCompile(`(?s)</th> <td class="footnote_plugin_text">`)
	r9 := regexp.MustCompile(`<div class="pdfprnt-buttons.*print=print  --></div></div>`)
	r10 := regexp.MustCompile(`<figure class="wp-block-table.*?>`)
	r11 := regexp.MustCompile(`</figure>`)

	// footnote := r.FindString(oldtxt)
	// newtxt := r.ReplaceAllString(oldtxt, `<span class="footnote_referrer"><a href="#footnote_plugin_reference_$1"><sup id="footnote_plugin_tooltip_$1" class="footnote_plugin_tooltip_text">[$2]</sup></a></span>`)
	newtxt := r.ReplaceAllString(oldtxt, ``)
	newtxt = r2.ReplaceAllString(newtxt, `<span class="footnote_referrer"><a href="#footnote_plugin_reference_$1"><sup id="footnote_plugin_tooltip_$1" class="footnote_plugin_tooltip_text">[$2]</sup></a></span>`)
	newtxt = r5.ReplaceAllString(newtxt, `</span></a><span class="footnote_referrer"><a href="#footnote_plugin_reference_$1"><sup id="footnote_plugin_tooltip_$1" class="footnote_plugin_tooltip_text">[$2]</sup></a></span></li>`)
	newtxt = r3.ReplaceAllString(newtxt, `<a id="footnote_plugin_reference_$1" class="footnote_backlink" href="#footnote_plugin_tooltip_$1"><span class="footnote_index_arrow">↑</span>$2$3</a>`)
	newtxt = r4.ReplaceAllString(newtxt, ``)
	newtxt = r6.ReplaceAllString(newtxt, `$1`)
	newtxt = r7.ReplaceAllString(newtxt, `<p class="footnote_rkmg">$1</p>`)
	newtxt = r8.ReplaceAllString(newtxt, ` `)
	newtxt = r9.ReplaceAllString(newtxt, ``)
	newtxt = r10.ReplaceAllString(newtxt, ``)
	newtxt = r11.ReplaceAllString(newtxt, ``)

	// color.Info.Println("Footnote: " + footnote)

	// fmt.Printf("%q\n", r.FindString(oldtxt))
	// re := regexp.MustCompile(`foo.?`)
	// fmt.Printf("%q\n", re.FindString("seafood fool"))
	// fmt.Printf("%q\n", re.FindString("meat"))

	if matched {
		color.Info.Println("Fant fotnote")
		// fmt.Println(newtxt)
		err := ioutil.WriteFile(filename, []byte(newtxt), 0777)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		color.Warn.Println("Fant ingen fotnote")
	}
}
func addFiles(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	color.Info.Println("basePath: " + basePath)
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		fmt.Println(err)
	}

	re := regexp.MustCompile(`.*\.pdf$`)
	for _, file := range files {
		fmt.Println(basePath + file.Name())
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(basePath + file.Name())
			if err != nil {
				fmt.Println(err)
			}
			if re.MatchString(file.Name()) {
				continue
			}

			// Add some files to the archive.
			f, err := w.Create(baseInZip + "/" + file.Name())
			if err != nil {
				fmt.Println(err)
			}
			_, err = f.Write(dat)
			if err != nil {
				fmt.Println(err)
			}
		} else if file.IsDir() {

			// Recurse
			newBase := basePath + "/" + file.Name() + "/"
			fmt.Println("Recursing and Adding SubDir: " + file.Name())
			fmt.Println("Recursing and Adding SubDir: " + newBase)
			if !re.MatchString(file.Name()) {
				addFiles(w, newBase, baseInZip+file.Name()+"/")
			}
		}
	}
}
func MkEpub(baseFolder string, epub string) {
	// Get a Buffer to Write To
	outFile, err := os.Create(epub)
	if err != nil {
		fmt.Println(err)
	}
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	// Add some files to the archive.
	addFiles(w, baseFolder, "")

	if err != nil {
		fmt.Println(err)
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		fmt.Println(err)
	}
}
func main() {
	version := "0.2"
	var xmlfile string
	var newepub string
	var zipout string
	flag.Usage = func() {
		usagemsg := "epub-fix, version: " + version + " - Fix referneces in ePub"
		fmt.Fprintf(os.Stderr, usagemsg+"%s -x=forord.xml", os.Args[0])
		flag.PrintDefaults()
		color.Error.Println("Prøv igjen!")
	}
	x := flag.String("x", "", "Hvilke xml fil skal du fikse?")
	e := flag.String("e", "", "Hvilke ePub fil skal du fikse?")

	// imp := flag.Bool("imp", false, "If imp is used - script will be importing and not exporting pdb")

	flag.Parse()
	if *x == "" && *e == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *x != "" && *e != "" {
		color.Error.Println("Bruk kun 'x' eller 'v', ikke begge samtidig")
		flag.Usage()
		os.Exit(1)
	}
	color.Info.Println("epub-fix version: " + version)
	if *x != "" {
		color.Info.Println("xml fil:", *x)
	}
	if *x != "" {
		xmlfile = *x
		fix_xml(xmlfile)
	}
	if *e != "" {
		color.Info.Println("ePub fil:", *e)
		re := regexp.MustCompile(".epub$")
		newepub = re.ReplaceAllString(*e, "_rkmg.epub")
		zipout = "tmpepub"
		files := unzipEpub(*e, zipout)
		htmlreg := regexp.MustCompile(`0000_.*xhtml$`)
		cssreg := regexp.MustCompile(`stylesheet.css`)
		for _, file := range files {
			htmlmatch := htmlreg.MatchString(file)
			cssmatch := cssreg.MatchString(file)
			if htmlmatch {
				color.Info.Println("Fount xhtml file: " + file)
				xmlfile = file
			}
			if cssmatch {
				color.Info.Println("Fount css file: " + file)
				oldtxt := read_file(file)
				newtxt := oldtxt + `
				.footnote_backlink{
					font-family: Arial, Helvetica, sans-serif;
					font-weight: 900;
				  }
				  .footnote_rkmg{
				  border-top: 2px solid;
				  }
				  .has-gray-background-color{
					background-color: rgb(248, 188, 173)  
					}
					.has-drop-cap:not(:focus):first-letter {
			float:left;
			font-size:8.4em;
			line-height:.68;
			font-weight:100;
			margin:.05em .1em 0 0;
			text-transform:uppercase;
			font-style:normal
		   }
		   p.has-drop-cap.has-background {
			overflow:hidden;
		   }
		   .has-small-font-size {
			   font-size: .8125em;
		   }
		   .has-gray-background-color, .has-gray-background-color[class] {
			   background-color: #f8bcad;
			   font-family: open-quote;
			   font-style: italic;
		   }
	  `
				err := ioutil.WriteFile(file, []byte(newtxt), 0777)
				if err != nil {
					log.Fatal(err)
				}
			}

		}
		fix_xml(xmlfile)
		MkEpub(zipout, newepub)
	}
}
