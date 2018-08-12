package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

var n = 2

func main() {
	files := validateFiles("./training/test")

	ch := make(chan string)
	for _, v := range files {
		go formatText(v, ch)
	}

	data := make([]string, len(files))
	for i := range files {
		data[i] = <-ch
	}

	fmt.Print(data)
}

/**
* Takes in a list of files/folders and checks to make
* sure they exist, then returns an array of the
* individual files in all the directories
 */
func validateFiles(filepath ...string) []string {
	ret := make([]string, 0, 10)
	for _, v := range filepath {
		_validateFiles(v, &ret)
	}
	return ret
}
func _validateFiles(currentFile string, files *[]string) {
	file, err := os.Stat(currentFile)
	if err != nil {
		fmt.Printf("Error accessing %s\n", currentFile)
		return
	}

	if file.IsDir() {
		if dir, err := ioutil.ReadDir(currentFile); err == nil {
			for _, v := range dir {
				_validateFiles(path.Join(currentFile, v.Name()), files)
			}
		} else {
			fmt.Printf("Error reading directory %s, skipping\n", currentFile)
			return
		}
	} else {
		*files = append(*files, currentFile)
	}
}

/**
* Takes in the list of files from validateFiles() and reads in the
* text of each one, applying necessary preprocessing and formatting
* rules to it so it can be used to build a model, and returns it
 */
func formatText(filepath string, ch chan string) {
	// Some helpful regexes that we'll need later
	directionLine := regexp.MustCompile(`^[\[\(]`)  // Match a line that is a stage direction
	dialogueLine := regexp.MustCompile(`^[A-Z]*:`)  // Match a line of dialogue
	directionTag := regexp.MustCompile(`(\(.*?\))`) // Match a direction (within another line)
	speakerTag := regexp.MustCompile(`([A-Z]*:)`)   // Match speaker at start of dialogue line

	// Read file in as string
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Error reading file %s, skipping it", filepath)
		return
	}

	text := string(b)
	text = "<start>\n<scene>\n" + text + "\n</scene>\n<end>" // Add start and end tags to text

	lines := strings.Split(text, "\n")                 // Split text on newlines
	lines[2] = "<lsetting>" + lines[2] + "</lsetting>" // Add <lsetting> tag to setting of first scene
	for j := 0; j < len(lines); j++ {
		// If we find 3 blank lines, its a scene change
		if lines[j] == "" && lines[j+1] == "" && lines[j+2] == "" {
			// Add closing scene tag
			lines[j] = "</scene>"
			for lines[j+1] == "" { // Skip blank lines until we reach next scene
				j = j + 1
			}
			lines[j] = "<scene>"
			lines[j+1] = "<lsetting>" + lines[j+1] + "</lsetting>"
			continue
		}

		if dialogueLine.MatchString(lines[j]) { // If this line is dialogue
			lines[j] = "<ldialogue>" + lines[j] + "</ldialogue>"
			lines[j] = directionTag.ReplaceAllString(lines[j], "<direction>${1}</direction>")
			lines[j] = speakerTag.ReplaceAllString(lines[j], "<speaker>${1}</speaker>")
		} else if directionLine.MatchString(lines[j]) { // If this line is a stage direction
			lines[j] = "<ldirection>" + lines[j] + "</ldirection>"
		}
	}
	ch <- strings.Join(lines, "\n")
}

// func buildModel(files) {

// }
