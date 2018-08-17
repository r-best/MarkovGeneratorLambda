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

func mergeMaps(maps ...*map[string]int) {
	for i := range maps[1:len(maps)] {
		for k, v := range *maps[i+1] {
			(*maps[0])[k] += v
		}
	}
}
