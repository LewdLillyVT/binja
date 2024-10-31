package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// findPattern locates all occurrences of the pattern in binary data.
func findPattern(data, pattern []byte) []int64 {
	var positions []int64
	dataLen := int64(len(data))
	patternLen := int64(len(pattern))

	for i := int64(0); i <= dataLen-patternLen; i++ {
		if bytes.Equal(data[i:i+patternLen], pattern) {
			positions = append(positions, i)
		}
	}
	return positions
}

// parsePattern tries to interpret the input as either hex or raw binary bytes.
func parsePattern(input string) ([]byte, error) {
	// Try interpreting as hex
	pattern, err := hex.DecodeString(input)
	if err == nil {
		return pattern, nil
	}

	// Try interpreting as raw bytes (e.g., "0xDE 0xAD 0xBE 0xEF" or "DE AD BE EF")
	parts := strings.Fields(input)
	pattern = make([]byte, len(parts))
	for i, part := range parts {
		// Remove "0x" prefix if it exists
		part = strings.TrimPrefix(part, "0x")
		// Parse each byte as hex
		byteValue, err := hex.DecodeString(part)
		if err != nil || len(byteValue) != 1 {
			return nil, fmt.Errorf("invalid byte format: %s", part)
		}
		pattern[i] = byteValue[0]
	}
	return pattern, nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Prompt user to drag and drop files into the console
	fmt.Println("Please drag and drop files into this console, then press Enter to proceed:")
	filesInput, _ := reader.ReadString('\n')
	filesInput = strings.TrimSpace(filesInput) // Remove whitespace and newline characters
	files := filepath.SplitList(filesInput)

	// Trim quotes from file paths
	for i, filePath := range files {
		files[i] = strings.Trim(filePath, "\"")
	}

	if len(files) == 0 {
		fmt.Println("Error: No files provided. Please drag and drop at least one file.")
		return
	}

	// Prompt user to enter a binary pattern
	var pattern []byte
	for {
		fmt.Print("Enter the binary pattern to search for (e.g., 'deadbeef' or '0xDE 0xAD 0xBE 0xEF'): ")
		patternInput, _ := reader.ReadString('\n')
		patternInput = strings.TrimSpace(patternInput)

		var err error
		pattern, err = parsePattern(patternInput)
		if err != nil {
			fmt.Println("Invalid pattern format. Please try again.")
			continue
		}
		break
	}

	// Iterate over each file path provided
	for _, filePath := range files {
		// Debug output to see the exact file path being processed
		fmt.Printf("\nSearching for pattern in file: '%s'\n", filePath)

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("Error: File %s does not exist. Skipping.\n", filePath)
			continue
		}

		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", filePath, err)
			continue
		}
		defer file.Close()

		// Read and search the file in chunks
		const chunkSize = 4096
		var offset int64
		for {
			buf := make([]byte, chunkSize)
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				break
			}
			if n == 0 {
				break
			}

			// Search for the pattern within the current chunk
			positions := findPattern(buf[:n], pattern)
			for _, pos := range positions {
				fmt.Printf("Pattern found in %s at offset %d\n", filePath, offset+pos)
			}

			// Update the offset for the next chunk
			offset += int64(n)
		}
		fmt.Printf("Pattern search completed for file: %s\n", filepath.Base(filePath))
	}

	// Keep the console open by prompting the user to press Enter to exit
	fmt.Println("\nSearch complete. Press Enter to exit.")
	reader.ReadString('\n')
}
