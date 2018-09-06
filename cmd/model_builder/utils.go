package main

import (
	"encoding/json"
	"fmt"
	"os"
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

// ProbabilityModel is just a 2-layer nested map of the following form:
// 		ProbabilityModel[(n-1)-gram][token] = probability of that (n-1)-gram
// 		being followed by that token
type ProbabilityModel map[string]map[string]float64

// WriteModel writes the probability model P to the given file
func (P *ProbabilityModel) WriteModel(filepath string) {
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Println(err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "	")

	err = encoder.Encode(P)
	if err != nil {
		fmt.Println(err)
	}
}

func keys(m *map[string]int) []string {
	keys := make([]string, len(*m))
	i := 0
	for k := range *m {
		keys[i] = k
		i++
	}
	return keys
}

func mergeArraysUniq(arrays ...[]string) []string {
	temp := make(map[string]int)
	for i := range arrays {
		for j := range arrays[i] {
			temp[arrays[i][j]] = 1
		}
	}

	return keys(&temp)
}

func mergeMaps(maps ...*map[string]int) *map[string]int {
	for i := range maps[1:len(maps)] {
		for k, v := range *maps[i+1] {
			(*maps[0])[k] += v
		}
	}
	return maps[0]
}

// This is a simple merging of an array of FrequencyObj objects
// into one. This would be so much easier in JavaScript.
func mergeFreqObjs(frequencyObjs ...*FrequencyObj) ([]string, *map[string]int, *map[string]int) {
	tokens := make([][]string, len(frequencyObjs))
	for i := range frequencyObjs {
		tokens[i] = frequencyObjs[i].tokens
	}

	n1grams := make([]*map[string]int, len(frequencyObjs))
	for i := range frequencyObjs {
		n1grams[i] = frequencyObjs[i].n1grams
	}

	ngrams := make([]*map[string]int, len(frequencyObjs))
	for i := range frequencyObjs {
		ngrams[i] = frequencyObjs[i].ngrams
	}

	return mergeArraysUniq(tokens...), mergeMaps(n1grams...), mergeMaps(ngrams...)
}
