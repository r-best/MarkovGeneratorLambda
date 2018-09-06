package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// ReadFiles takes in a list of files/folders and recursively
// iterates through them, reading in the text from
// each file as a string and finally returning a
// string array of all of them
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
