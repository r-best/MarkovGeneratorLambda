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
	files := ReadFiles("./training/test")

	// Process the text from each file in parallel
	ch := make(chan []string)
	for _, v := range files {
		go FormatText(v, ch)
	}

	// Get the data back from the formatText() calls,
	// each call returns a string array of the lines
	// of its file, so datas is an array of those
	// string arrays
	datas := make([][]string, len(files))
	for i := range files {
		datas[i] = <-ch
	}

	fmt.Print(datas)
}

/**
* Takes in a list of files/folders and recursively
* iterates through them, reading in the text from
* each file as a string and finally returning a
* string array of all of them
 */
func ReadFiles(filepath ...string) []string {
	ret := make([]string, 0, 10)
	for _, v := range filepath {
		_readFiles(v, &ret)
	}
	return ret
}
func _readFiles(currentFile string, files *[]string) {
	file, err := os.Stat(currentFile)
	if err != nil {
		fmt.Printf("Error accessing %s\n", currentFile)
		return
	}

	if file.IsDir() {
		if dir, err := ioutil.ReadDir(currentFile); err == nil {
			for _, v := range dir {
				_readFiles(path.Join(currentFile, v.Name()), files)
			}
		} else {
			fmt.Printf("Error reading directory %s, skipping\n", currentFile)
			return
		}
	} else {
		// Read file in as string
		if b, err := ioutil.ReadFile(currentFile); err == nil {
			*files = append(*files, string(b))
		} else {
			fmt.Printf("Error reading file %s, skipping it\n", currentFile)
			return
		}
	}
}

/**
* Takes in a string of training data (one episode of seinfeld) and
* applies preprocessing/formatting rules so it can be used to build
* a model. Returns the formatted string to the given channel.
 */
func FormatText(text string, ch chan []string) {
	// Some helpful regexes that we'll need later
	directionLine := regexp.MustCompile(`^[\[\(]`)  // Match a line that is a stage direction
	dialogueLine := regexp.MustCompile(`^[A-Z]*:`)  // Match a line of dialogue
	directionTag := regexp.MustCompile(`(\(.*?\))`) // Match a direction (within another line)
	speakerTag := regexp.MustCompile(`([A-Z]*:)`)   // Match speaker at start of dialogue line

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
	ch <- lines
}
