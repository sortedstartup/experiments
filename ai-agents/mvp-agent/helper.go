package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// copyStarterTemplate copies the starter template to a timestamped output directory
func copyStarterTemplate() (string, error) {
	// Create timestamp for unique folder name
	timestamp := time.Now().Format("20060102_150405")
	outputDirName := fmt.Sprintf("output_%s", timestamp)

	// Define paths
	sourceDir := "./data/starter-template"
	outputBaseDir := "./data/outputs"
	outputDir := filepath.Join(outputBaseDir, outputDirName)

	// Create outputs directory if it doesn't exist
	if err := os.MkdirAll(outputBaseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create outputs directory: %v", err)
	}

	// Copy the starter template
	if err := copyDir(sourceDir, outputDir); err != nil {
		return "", fmt.Errorf("failed to copy directory: %v", err)
	}

	return outputDir, nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Set file permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}

// copyPRDToOutput copies the PRD file to the output directory
func copyPRDToOutput(prdPath, outputDir string) error {
	// Check if PRD file exists
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return fmt.Errorf("PRD file does not exist: %s", prdPath)
	}

	// Get the base filename from the PRD path
	prdFilename := filepath.Base(prdPath)

	// Define destination path in output directory
	dstPath := filepath.Join(outputDir, prdFilename)

	// Copy the PRD file
	if err := copyFile(prdPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy PRD file: %v", err)
	}

	return nil
}
