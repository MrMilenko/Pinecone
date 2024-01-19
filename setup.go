package main

import (
	"fmt"
	"os"
)

func checkDataFolder(dataFolder string) bool {
	// Ensure data folder exists
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		fmt.Println("Data folder not found. Creating...")
		if mkDirErr := os.Mkdir(dataFolder, 0755); mkDirErr != nil {
			fmt.Println("Error creating data folder:", mkDirErr)
			return false
		}
	}
	return true
}

func checkDatabaseFile(jsonFilePath string, jsonURL string, updateFlag bool) bool {
	// Check if JSON file exists
	if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
		// Prompt for download if JSON file doesn't exist
		if promptForDownload(jsonURL) {
			err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", "data/id_database.json", &titles, true)
			if err != nil {
				fmt.Println("Error downloading data:", err)
				return false
			}
		} else {
			fmt.Println("Download aborted by user.")
			return false
		}
	} else if updateFlag {
		// Handle manual update
		err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", jsonFilePath, &titles, true)
		if err != nil {
			fmt.Println("Error updating data:", err)
			return false
		}
	} else {
		// Load existing JSON data
		err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", jsonFilePath, &titles, false)
		if err != nil {
			fmt.Println("Error loading data:", err)
			return false
		}
	}
	return true
}

func checkDumpFolder(dumpLocation string) bool {
	if dumpLocation != "dump" {
		if _, err := os.Stat(dumpLocation); os.IsNotExist(err) {
			fmt.Println("Directory does not exist, exiting...")
			return false
		}
	} else {
		if _, err := os.Stat(dumpLocation); os.IsNotExist(err) {
			fmt.Println("Default dump folder not found. Creating...")
			if mkDirErr := os.Mkdir(dumpLocation, 0755); mkDirErr != nil {
				fmt.Println("Error creating dump folder:", mkDirErr)
				return false
			}
			fmt.Println("Please place TDATA folder in the \"dump\" folder")
			return false
		}
	}
	return true
}
