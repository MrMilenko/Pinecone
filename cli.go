package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
)

const (
	headerWidth = 100
	separator   = ""
)

type CLIOptions struct {
	DataFolder   string
	JSONFilePath string
	JSONUrl      string
}

func printHeader(title string) {
	title = strings.TrimSpace(title)
	if len(title) > headerWidth-6 { // -6 to account for spaces and equals signs
		title = title[:headerWidth-9] + "..."
	}
	formattedTitle := "== " + title + " =="
	padLen := (headerWidth - len(formattedTitle)) / 2
	color.New(color.FgCyan).Println(strings.Repeat("=", padLen) + formattedTitle + strings.Repeat("=", headerWidth-padLen-len(formattedTitle)))
}

func printInfo(colorCode color.Attribute, format string, args ...interface{}) {
	color.New(colorCode).Printf("    "+format, args...)
}

// Prints statistics for a specific title or for all titles if batch is true.
func printStats(titleID string, batch bool) {
	if batch {
		printTotalStats()
	} else {
		data, ok := titles.Titles[titleID]
		if !ok {
			fmt.Printf("No data found for title ID %s\n", titleID)
			return
		}
		fmt.Printf("Statistics for title ID %s:\n", titleID)
		printTitleStats(&data)
	}
}

// Prints statistics for TitleData.
func printTitleStats(data *TitleData) {
	fmt.Println("Title:", data.TitleName)
	fmt.Println("Total number of Content IDs:", len(data.ContentIDs))
	fmt.Println("Total number of Title Updates:", len(data.TitleUpdates))
	fmt.Println("Total number of Known Title Updates:", len(data.TitleUpdatesKnown))
	fmt.Println("Total number of Archived items:", len(data.Archived))
	fmt.Println()
}

func printTotalStats() {
	totalTitles := len(titles.Titles)
	totalContentIDs := 0
	totalTitleUpdates := 0
	totalKnownTitleUpdates := 0
	totalArchivedItems := 0

	// Set to store unique hashes of known title updates and archived items
	knownTitleUpdateHashes := make(map[string]struct{})
	archivedItemHashes := make(map[string]struct{})

	for _, data := range titles.Titles {
		totalContentIDs += len(data.ContentIDs)
		totalTitleUpdates += len(data.TitleUpdates)

		// Count unique known title updates
		for _, knownUpdate := range data.TitleUpdatesKnown {
			for hash := range knownUpdate {
				knownTitleUpdateHashes[hash] = struct{}{}
			}
		}

		// Count unique archived items
		for _, archivedItem := range data.Archived {
			for hash := range archivedItem {
				archivedItemHashes[hash] = struct{}{}
			}
		}
	}

	totalKnownTitleUpdates = len(knownTitleUpdateHashes)
	totalArchivedItems = len(archivedItemHashes)

	fmt.Println("Total Titles:", totalTitles)
	fmt.Println("Total Content IDs:", totalContentIDs)
	fmt.Println("Total Title Updates:", totalTitleUpdates)
	fmt.Println("Total Known Title Updates:", totalKnownTitleUpdates)
	fmt.Println("Total Archived Items:", totalArchivedItems)
}

func cliPromptForDownload(url string) bool {
	var response string
	fmt.Printf("The required JSON data is not found. It can be downloaded from %s\n", url)
	fmt.Print("Do you want to download it now? (yes/no): ")
	fmt.Scanln(&response)

	return strings.ToLower(response) == "yes"
}

func startCLI(options CLIOptions) {
	err := checkDataFolder(options.DataFolder)
	if err != nil {
		log.Fatalln(err)
	}

	err = checkDatabaseFile(options.JSONFilePath, options.JSONUrl, updateFlag, nil)
	if err != nil {
		log.Fatalln(err)
	}

	err = checkDumpFolder(dumpLocation)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("Pinecone v%s\n", version)
	fmt.Println("Please share output of this program with the Pinecone team if you find anything interesting!")

	err = checkParsingSettings()
	if err != nil {
		log.Fatalln(err)
	}
}
