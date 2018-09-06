package main

import (
	"regexp"
	"strings"

	utils "markovgenerator/internal"
)

// N : value of n to use for n-grams
var N = 3

// OutputFilePath : path to the file where the probability model will be written
var OutputFilePath = "./outputbooyah.json"

func main() {
	// Read each file into a string array
	files := utils.ReadFiles("../../training_raw/test")

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

	// Collect the frequency data from all the FrequencyObjs
	tokens, n1grams, ngrams := mergeFreqObjs(frequencies...)

	// Use the frequencies to calculate the probability
	// of each ngram given each (n-1)gram
	// I wish this part could be parallelized, it's
	// the part that takes the majority of the runtime
	P := CalculateProbabilities(tokens, n1grams, ngrams)

	P.WriteModel(OutputFilePath)
}

// FormatText takes in a string of training data (one episode of seinfeld)
// and applies preprocessing/formatting rules so it can be used to build
// a model. Returns the formatted string to the given channel.
func FormatText(text string) []string {
	// Some helpful regexes that we'll need later
	directionLine := regexp.MustCompile(`^[\[\(]`)  // Match a line that is a stage direction
	dialogueLine := regexp.MustCompile(`^[A-Z]*:`)  // Match a line of dialogue
	directionTag := regexp.MustCompile(`\((.*?)\)`) // Match a direction (within another line)
	speakerTag := regexp.MustCompile(`([A-Z]*:)`)   // Match speaker at start of dialogue line

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
	return lines
}

// CountFrequencies takes in an array of the lines of a training file and a value of N
// to use, and generates all the n-grams and (n-1)-grams it can find, returning the
// information in a new FrequencyObj
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
func CalculateProbabilities(tokens []string, n1grams *map[string]int, ngrams *map[string]int) *ProbabilityModel {
	P := make(ProbabilityModel)
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
