package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Define the test data directory
	testDataDir := "../../../test/testdata"

	// Find all .hex files in the testdata directory
	hexFiles, err := filepath.Glob(filepath.Join(testDataDir, "*.hex"))
	if err != nil {
		fmt.Printf("Failed to find hex files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d .hex files in %s\n", len(hexFiles), testDataDir)

	// Find all .json files in the testdata directory
	jsonFiles, err := filepath.Glob(filepath.Join(testDataDir, "*.json"))
	if err != nil {
		fmt.Printf("Failed to find json files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d .json files in %s\n", len(jsonFiles), testDataDir)

	// Create a map of base names to JSON files for quick lookup
	jsonFileMap := make(map[string]string)
	for _, jsonFile := range jsonFiles {
		baseName := strings.TrimSuffix(filepath.Base(jsonFile), ".json")
		jsonFileMap[baseName] = jsonFile
	}

	// Find matching .hex and .json file pairs
	var matchingPairs []struct {
		hexFile  string
		jsonFile string
	}

	for _, hexFile := range hexFiles {
		baseName := strings.TrimSuffix(filepath.Base(hexFile), ".hex")
		if jsonFile, ok := jsonFileMap[baseName]; ok {
			matchingPairs = append(matchingPairs, struct {
				hexFile  string
				jsonFile string
			}{
				hexFile:  hexFile,
				jsonFile: jsonFile,
			})
		}
	}

	fmt.Printf("Found %d matching .hex and .json file pairs:\n", len(matchingPairs))
	for i, pair := range matchingPairs {
		fmt.Printf("%d. %s <-> %s\n", i+1, filepath.Base(pair.hexFile), filepath.Base(pair.jsonFile))
	}
}
