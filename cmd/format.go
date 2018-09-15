package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	utils "markovgenerator/internal"

	"github.com/spf13/cobra"
)

var outDir = "training_proc"

// Some helpful regexes that we'll need later
var directionLine = regexp.MustCompile(`^[\[\(]`)  // Match a line that is a stage direction
var dialogueLine = regexp.MustCompile(`^[A-Z]*:`)  // Match a line of dialogue
var directionTag = regexp.MustCompile(`\((.*?)\)`) // Match a direction (within another line)
var speakerTag = regexp.MustCompile(`([A-Z]*:)`)   // Match speaker at start of dialogue line

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format raw data to be parsable by the trainer",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reading training files...")
		// Read each file into a string array
		files := utils.ReadFiles(args[0])

		// Format the text & write each file in parallel
		ch := make(chan bool)
		for path, text := range files {
			go func(path string, text string, ch chan bool) {
				fmt.Printf("Formatting %s\n", path)
				formattedText := FormatText(text)
				fmt.Printf("Writing %s\n", path)
				writeFormattedText(path, formattedText)
				ch <- true
			}(path, text, ch)
		}

		// Join all the goroutines and finish the program
		for range files {
			_ = <-ch
		}
		close(ch)
		fmt.Println("Finished!")
	},
}

// FormatText takes in a string of training data (one episode of seinfeld)
// and applies preprocessing/formatting rules so it can be used to build
// a model. Returns the formatted string to the given channel.
func FormatText(text string) string {
	text = "<start>\n<scene>\n" + text + "\n</scene>\n<end>" // Add start and end tags to text

	lines := strings.Split(text, "\n")                   // Split text on newlines
	lines[2] = "<lsetting> " + lines[2] + " </lsetting>" // Add <lsetting> tag to setting of first scene
	for j := 0; j < len(lines); j++ {
		// If we find 3 blank lines, its a scene change
		if lines[j] == "" && lines[j+1] == "" && lines[j+2] == "" {
			// Add closing scene tag
			lines[j] = "</scene>"
			for lines[j+1] == "" { // Skip blank lines until we reach next scene
				j = j + 1
			}
			lines[j] = "<scene>"
			lines[j+1] = "<lsetting> " + lines[j+1] + " </lsetting>"
			continue
		}

		if dialogueLine.MatchString(lines[j]) { // If this line is dialogue
			lines[j] = "<ldialogue> " + lines[j] + " </ldialogue>"
			lines[j] = directionTag.ReplaceAllString(lines[j], "<direction> ${1} </direction>")
			lines[j] = speakerTag.ReplaceAllString(lines[j], "<speaker> ${1} </speaker>")
		} else if directionLine.MatchString(lines[j]) { // If this line is a stage direction
			lines[j] = "<ldirection> " + lines[j] + " </ldirection>"
		}
	}
	return strings.Join(lines, "\n")
}

// WriteModel writes the probability model P to the given file
func writeFormattedText(path string, text string) {
	temp := strings.Split(path, "/")
	temp[0] = outDir
	directory := filepath.Join(temp[0 : len(temp)-1]...)
	filename := temp[len(temp)-1]

	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	file, err := os.Create(filepath.Join(directory, filename))
	defer file.Close()
	if err != nil {
		fmt.Println(err)
	}

	_, err = file.WriteString(text)
	if err != nil {
		fmt.Println(err)
	}
}
