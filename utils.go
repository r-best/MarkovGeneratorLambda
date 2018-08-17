package main

func keys(m *map[string]int) []string {
	keys := make([]string, len(*m))
	i := 0
	for k := range *m {
		keys[i] = k
		i++
	}
	return keys
}

func mergeArrays(arrays ...[]string) []string {
	length := 0
	for i := range arrays {
		length += len(arrays[i])
	}

	array := make([]string, length)
	for i := range arrays {
		for j := range arrays[i] {
			array[i+j] = arrays[i][j]
		}
	}

	return array
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
func mergeFreqObjs(frequencyObjs ...*FrequencyObj) FrequencyObj {
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

	return FrequencyObj{
		tokens:  mergeArrays(tokens...),
		n1grams: mergeMaps(n1grams...),
		ngrams:  mergeMaps(ngrams...)}
}
