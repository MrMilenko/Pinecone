package main

import (
	"fmt"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

func checkDataFolder(dataFolder string) error {
	// Ensure data folder exists
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		fmt.Println("Data folder not found. Creating...")
		if mkDirErr := os.Mkdir(dataFolder, 0755); mkDirErr != nil {
			return fmt.Errorf("Error creating data folder: %v", mkDirErr)
		}
	}
	return nil
}

func checkDatabaseFile(jsonFilePath string, jsonURL string, updateFlag bool, window ...fyne.Window) error {
	// Check if JSON file exists
	if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
		// Prompt for download if JSON file doesn't exist
		if guiEnabled {
			if len(window) != 1 {
				output1 := canvas.NewText("ERROR: Your local developer did not use the a function correctly!", theme.ErrorColor())
				output2 := canvas.NewText("Please open a GitHub issue and show them this output", theme.ErrorColor())
				outputContainer.Add(output1)
				outputContainer.Add(output2)
			}

			guiShowDownloadConfirmation(window[0], jsonFilePath, jsonURL)
		} else {
			if cliPromptForDownload(jsonURL) {
				err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", "data/id_database.json", &titles, true)
				if err != nil {
					return fmt.Errorf("error downloading data: %v ", err)
				}
			} else {
				return fmt.Errorf("download aborted by user")
			}
		}
	} else if updateFlag {
		// Handle manual update
		err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", jsonFilePath, &titles, true)
		if err != nil {
			return fmt.Errorf("error updating data: %v", err)
		}
	} else {
		// Load existing JSON data
		err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", jsonFilePath, &titles, false)
		if err != nil {
			return fmt.Errorf("error loading data: %v", err)
		}
		if guiEnabled {
			guiScanDump()
		}
	}
	return nil
}

func checkDumpFolder(dumpLocation string) error {
	if dumpLocation != "dump" {
		if _, err := os.Stat(dumpLocation); os.IsNotExist(err) {
			return fmt.Errorf("Directory does not exist, exiting...")
		}
	} else {
		if _, err := os.Stat(dumpLocation); os.IsNotExist(err) {
			fmt.Println("Default dump folder not found. Creating...")
			if mkDirErr := os.Mkdir(dumpLocation, 0755); mkDirErr != nil {
				return fmt.Errorf("Error creating dump folder: %v", mkDirErr)
			}
			return fmt.Errorf("Please place TDATA folder in the \"dump\" folder")
		}
	}
	return nil
}

func checkParsingSettings() error {
	if titleIDFlag != "" {
		// if the titleID flag is set, print stats for that title
		printStats(titleIDFlag, false)
	} else if summarizeFlag {
		// if the summarize flag is set, print stats for all titles
		printStats("", true)
	} else if fatxplorer {
		if runtime.GOOS == "windows" {
			if _, err := os.Stat(`X:\`); os.IsNotExist(err) {
				return fmt.Errorf(`FatXplorer's X: drive not found`)
			} else {
				fmt.Println("Checking for Content...")
				fmt.Println("====================================================================================================")
				err := checkForContent("X:\\TDATA")
				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("FatXplorer mode is only available on Windows.")
		}
	} else {
		// If no flag is set, proceed normally
		// Check if TDATA folder exists
		if _, err := os.Stat(dumpLocation + "/TDATA"); os.IsNotExist(err) {
			return fmt.Errorf("TDATA folder not found. Please place TDATA folder in the dump folder.")
		}
		fmt.Println("Checking for Content...")
		fmt.Println("====================================================================================================")
		err := checkForContent(dumpLocation + "/TDATA")
		if err != nil {
			return err
		}
	}

	return nil
}
