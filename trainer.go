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
	formatText(files)
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
func formatText(files []string) {
	for i := range files {
		b, err := ioutil.ReadFile(files[i])
		if err != nil {
			fmt.Printf("Error reading file %s, skipping it", files[i])
			continue
		}

		text := string(b)
		text = "<start>\n<scene>\n" + text + "\n</scene>\n<end>"

		dialogueLine := regexp.MustCompile(`^[A-Z]*:`)
		directionLine := regexp.MustCompile(`^[\[\(]`)
		directionTag := regexp.MustCompile(`(\(.*?\))`)

		lines := strings.Split(text, "\n")
		lines[2] = "<lsetting>" + lines[2] + "</lsetting>"
		for j := 0; j < len(lines); j++ {
			// If we find 3 blank lines, its a scene change
			if lines[j] == "" && lines[j+1] == "" && lines[j+2] == "" {
				lines[j] = "</scene>"
				for lines[j+1] == "" {
					j = j + 1
				}
				lines[j] = "<scene>"
				lines[j+1] = "<lsetting>" + lines[j+1] + "</lsetting>"
				continue
			}

			if dialogueLine.MatchString(lines[j]) {
				lines[j] = "<ldialogue>" + lines[j] + "</ldialogue>"
				lines[j] = directionTag.ReplaceAllString(lines[j], "<direction>${1}</direction>")
			}

			if directionLine.MatchString(lines[j]) {
				lines[j] = "<ldirection>" + lines[j] + "</ldirection>"
			}
		}

		fmt.Println("RESULTS:")
		for _, v := range lines {
			fmt.Println(v)
		}
	}
}

// func buildModel(files) {

// }
