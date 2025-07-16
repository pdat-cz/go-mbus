package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// MBusData represents the root element of the M-Bus XML data
type MBusData struct {
	XMLName          xml.Name         `xml:"MBusData" json:"-"`
	SlaveInformation SlaveInformation `xml:"SlaveInformation" json:"SlaveInformation,omitempty"`
	DataRecords      []DataRecord     `xml:"DataRecord" json:"DataRecords"`
}

// SlaveInformation represents the slave device information
type SlaveInformation struct {
	Id           string `xml:"Id" json:"Id,omitempty"`
	Manufacturer string `xml:"Manufacturer" json:"Manufacturer,omitempty"`
	Version      string `xml:"Version" json:"Version,omitempty"`
	ProductName  string `xml:"ProductName" json:"ProductName,omitempty"`
	Medium       string `xml:"Medium" json:"Medium,omitempty"`
	AccessNumber string `xml:"AccessNumber" json:"AccessNumber,omitempty"`
	Status       string `xml:"Status" json:"Status,omitempty"`
	Signature    string `xml:"Signature" json:"Signature,omitempty"`
}

// DataRecord represents a single data record in the M-Bus data
type DataRecord struct {
	Id            string `xml:"id,attr" json:"id,omitempty"`
	Function      string `xml:"Function" json:"Function,omitempty"`
	StorageNumber string `xml:"StorageNumber" json:"StorageNumber,omitempty"`
	Tariff        string `xml:"Tariff" json:"Tariff,omitempty"`
	Device        string `xml:"Device" json:"Device,omitempty"`
	Unit          string `xml:"Unit" json:"Unit,omitempty"`
	Value         string `xml:"Value" json:"Value,omitempty"`
}

func main() {
	// Directory containing XML files
	testdataDir := "../../test/testdata"

	// Get all XML files in the directory
	xmlFiles, err := filepath.Glob(filepath.Join(testdataDir, "*.xml"))
	if err != nil {
		fmt.Printf("Error finding XML files: %v\n", err)
		os.Exit(1)
	}

	// Also include .norm.xml files
	normXmlFiles, err := filepath.Glob(filepath.Join(testdataDir, "*.norm.xml"))
	if err != nil {
		fmt.Printf("Error finding .norm.xml files: %v\n", err)
		os.Exit(1)
	}

	// Combine both file lists
	xmlFiles = append(xmlFiles, normXmlFiles...)

	// Convert each XML file to JSON
	for _, xmlFile := range xmlFiles {
		// Read XML file
		xmlData, err := ioutil.ReadFile(xmlFile)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", xmlFile, err)
			continue
		}

		// Parse XML
		var mbusData MBusData

		// Create a decoder with ISO-8859-1 charset reader
		decoder := xml.NewDecoder(strings.NewReader(string(xmlData)))
		decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
			if charset == "ISO-8859-1" {
				return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
			}
			return input, nil
		}

		err = decoder.Decode(&mbusData)
		if err != nil {
			fmt.Printf("Error parsing XML file %s: %v\n", xmlFile, err)
			continue
		}

		// Convert to JSON
		jsonData, err := json.MarshalIndent(mbusData, "", "    ")
		if err != nil {
			fmt.Printf("Error converting to JSON for file %s: %v\n", xmlFile, err)
			continue
		}

		// Create JSON file name
		jsonFile := strings.TrimSuffix(xmlFile, filepath.Ext(xmlFile))
		if strings.HasSuffix(jsonFile, ".norm") {
			jsonFile = strings.TrimSuffix(jsonFile, ".norm")
		}
		jsonFile = jsonFile + ".json"

		// Write JSON file
		err = ioutil.WriteFile(jsonFile, jsonData, 0644)
		if err != nil {
			fmt.Printf("Error writing JSON file %s: %v\n", jsonFile, err)
			continue
		}

		fmt.Printf("Converted %s to %s\n", xmlFile, jsonFile)
	}

	fmt.Println("Conversion complete!")
}
