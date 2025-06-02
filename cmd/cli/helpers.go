package main

import (
	"os"
	"path/filepath"
	"strings"
)

func updateSourceFiles(path string, fi os.FileInfo, err error) error {
	// check for an error before doing anything else
	if err != nil {
		return err
	}

	// check if current file is directory
	if fi.IsDir() {
		return nil
	}

	// only check go files
	matched, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}

	// we have a matching file
	if matched {
		// read file contents
		read, err := os.ReadFile(path)
		if err != nil {
			exitGracefully(err)
		}

		newContents := strings.Replace(string(read), "myapp", appURL, -1)

		// write the changed file
		err = os.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			exitGracefully(err)
		}
	}

	return nil
}

func updateSource() {
	// walk entire project folder, including subfolders
	err := filepath.Walk(".", updateSourceFiles)
	if err != nil {
		exitGracefully(err)
	}
}
