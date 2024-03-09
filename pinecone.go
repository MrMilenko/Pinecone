package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

var (
	titles        TitleList
	updateFlag    = false
	summarizeFlag = false
	titleIDFlag   = ""
	fatxplorer    = false
	dumpLocation  = "./dump/"
	helpFlag      = false
	guiEnabled    = true
	version       = "0.5.0"
	dataPath      = "data"
)

func main() {
	flag.BoolVar(&updateFlag, "update", false, "Update the JSON data from the source URL")
	flag.BoolVar(&updateFlag, "u", false, "Update the JSON data from the source URL")
	flag.BoolVar(&summarizeFlag, "summarize", false, "Print summary statistics for all titles")
	flag.BoolVar(&summarizeFlag, "s", false, "Print summary statistics for all titles")
	flag.StringVar(&titleIDFlag, "titleid", "", "Filter statistics by Title ID")
	flag.StringVar(&titleIDFlag, "tID", "", "Filter statistics by Title ID")
	flag.BoolVar(&fatxplorer, "fatxplorer", false, "Use FatXplorer's X: drive")
	flag.BoolVar(&fatxplorer, "f", false, "Use FatXplorer's X: drive")
	flag.StringVar(&dumpLocation, "location", "./dump/", "Directory to search for TDATA/UDATA directories")
	flag.StringVar(&dumpLocation, "l", "./dump/", "Directory to search for TDATA/UDATA directories")
	flag.BoolVar(&helpFlag, "help", false, "Display help information")
	flag.BoolVar(&helpFlag, "h", false, "Display help information")
	flag.BoolVar(&guiEnabled, "gui", true, "Enable GUI")
	flag.BoolVar(&guiEnabled, "g", true, "Enable GUI")

	flag.Parse() // Parse command line flags

	// Check for help flag
	if helpFlag {
		fmt.Println("Usage of Pinecone:")
		fmt.Println("  -u, --update:     Update the JSON data from the source URL. If not set, uses local copies of data.")
		fmt.Println("  -s, --summarize:  Print summary statistics for all titles. If not set, checks for content in the TDATA folder.")
		fmt.Println("  -tID, --titleid:  Filter statistics by Title ID (-titleID=ABCD1234). If not set, statistics are computed for all titles.")
		fmt.Println("  -f, --fatxplorer: Use FATXPlorer's X drive as the root directory. If not set, runs as normal. (Windows Only)")
		fmt.Println("  -l --location:    Directory where TDATA/UDATA folders are stored. If not set, checks in \"dump\"")
		fmt.Println("  -h, --help:       Display this help information.")
		fmt.Println("  -g, --gui:        Enable the GUI interface (default = true)")
		return
	}

	switch runtime.GOOS {
	case "linux":
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dataPath = path.Clean(homedir + "/.config/Pinecone/data/")
	case "darwin":
		if dumpLocation == "dump" {
			dumpLocation = path.Clean("../dump")
		}
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dataPath = path.Clean(homedir + "/.config/Pinecone/data/")
	default:
		dataPath = path.Clean("data")
	}

	_, err := os.Stat(dataPath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dataPath, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	jsonFilePath := filepath.Join(dataPath, "id_database.json")
	jsonDataFolder := dataPath
	jsonURL := "https://api.github.com/repos/Xbox-Preservation-Project/Pinecone/contents/data/id_database.json"
	ignoreListURL := "https://api.github.com/repos/Xbox-Preservation-Project/Pinecone/contents/data/ignorelist.json"

	if guiEnabled {
		guiOpts := GUIOptions{
			DataFolder:   jsonDataFolder,
			JSONFilePath: jsonFilePath,
			JSONUrl:      jsonURL,
			IgnoreURL:    ignoreListURL,
		}

		startGUI(guiOpts)
	} else {
		cliOpts := CLIOptions{
			DataFolder:   jsonDataFolder,
			JSONFilePath: jsonFilePath,
			JSONUrl:      jsonURL,
			IgnoreURL:    ignoreListURL,
		}

		startCLI(cliOpts)
	}
}
