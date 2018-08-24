package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

// FrequencyObj stores a list of all tokens (1-grams)
// that appear in a file, as well as counts of the number
// of times the n-grams and (n-1)-grams made from those
// tokens appear
type FrequencyObj struct {
	tokens  []string
	n1grams *map[string]int
	ngrams  *map[string]int
}

// N : value of n to use for n-grams
var N = 3

// OutputFilePath : path to the file where the probability model will be written
var OutputFilePath = "./outputbooyah.json"

func main() {
	// Read each file into a string array
	files := ReadFiles("../../training/test")

	// Format the text & calculate the ngram
	// frequencies of each file in parallel
	ch := make(chan *FrequencyObj)
	for _, text := range files {
		go func(text string, ch chan *FrequencyObj) {
			formattedLines := FormatText(text)
			freq := CountFrequencies(formattedLines, N)
			ch <- freq
		}(text, ch)
	}

	// Read the frequency data from each goroutine
	// as they complete, closing the channel when done
	frequencies := make([]*FrequencyObj, len(files))
	for i := range files {
		frequencies[i] = <-ch
	}
	close(ch)

	// Sum all the frequency data together
	frequency := mergeFreqObjs(frequencies...)

	// Use the frequencies to calculate the probability
	// of each ngram given each (n-1)gram
	// I wish this part could be parallelized, it's
	// the part that takes the majority of the runtime
	P := CalculateProbabilities(&frequency)

	fmt.Println(P)

	WriteModel(P, OutputFilePath)
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
func FormatText(text string) []string {
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
	return lines
}

func CountFrequencies(lines []string, N int) *FrequencyObj {
	tokens := make(map[string]int, 0) // Array of all tokens (1-grams)
	n1grams := make(map[string]int)   // Map of all (n-1)-grams to their frequencies
	ngrams := make(map[string]int)    // Map of all n-grams to their frequencies
	for _, line := range lines {
		words := strings.Fields(line)

		if len(words) < N {
			continue
		}

		for i := N - 1; i < len(words); i++ {
			tokens[words[i]]++
			if i >= N-1 {
				n1grams[strings.Join(words[i-N+1:i], ` `)]++
				ngrams[strings.Join(words[i-N+1:i+1], ` `)]++
			}
		}
	}
	return &FrequencyObj{tokens: keys(&tokens), n1grams: &n1grams, ngrams: &ngrams}
}

// CalculateProbabilities takes in a FrequencyObj and uses the frequency data
// it holds to calculate for every (n-1)-gram, what is the probability each
// possible token has of occurring next
func CalculateProbabilities(freq *FrequencyObj) *map[string]map[string]float64 {
	tokens := (*freq).tokens
	n1grams := (*freq).n1grams
	ngrams := (*freq).ngrams

	P := make(map[string]map[string]float64)
	for n1gram := range *n1grams {
		P[n1gram] = make(map[string]float64)
		for _, token := range tokens {
			ngram := n1gram + " " + token
			if v, ok := (*ngrams)[ngram]; ok {
				P[n1gram][token] = float64(v) / float64((*n1grams)[n1gram])
			}
		}
	}
	return &P
}

// WriteModel writes the probability model P to the given file
func WriteModel(P *map[string]map[string]float64, filepath string) {
	obj, err := json.MarshalIndent(P, "", "	")

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(obj))

	if err := ioutil.WriteFile(filepath, obj, 0644); err != nil {
		fmt.Println(err)
	}
}