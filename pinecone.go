package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

var (
	titles        TitleList
	updateFlag    = flag.Bool("update", false, "Update the JSON data from the source URL")
	summarizeFlag = flag.Bool("summarize", false, "Print summary statistics for all titles")
	titleIDFlag   = flag.String("titleid", "", "Filter statistics by Title ID")
	fatxplorer    = flag.Bool("fatxplorer", false, "Use FatXplorer's X: drive")
	helpFlag      = flag.Bool("help", false, "Display help information")
)

type TitleData struct {
	TitleName         string              `json:"Title Name,"`
	ContentIDs        []string            `json:"Content IDs"`
	TitleUpdates      []string            `json:"Title Updates"`
	TitleUpdatesKnown []map[string]string `json:"Title Updates Known"`
	Archived          []map[string]string `json:"Archived"`
}

type TitleList struct {
	Titles map[string]TitleData `json:"Titles"`
}

func removeCommentsFromJSON(jsonStr string) string {
	// remove // style comments
	re := regexp.MustCompile(`(?m)^[ \t]*//.*\n?`)
	jsonStr = re.ReplaceAllString(jsonStr, "")

	// remove /* ... */ style comments
	re = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	jsonStr = re.ReplaceAllString(jsonStr, "")

	return jsonStr
}

func downloadJSONData(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func loadJSONData(jsonFilePath, owner, repo, path string, v interface{}, updateFlag bool) error {
	if updateFlag {

		//Notify we're checking for updates
		fmt.Printf("Checking for PineCone updates..\n")

		// Download JSON data
		jsonData, err := downloadJSONData(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path))

		// Check if downloaded JSON is different from existing JSON
		if _, err := os.Stat(jsonFilePath); err == nil {
			existingData, err := os.ReadFile(jsonFilePath)
			if err != nil {
				return err
			}
			existingHash := fmt.Sprintf("%x", sha1.Sum(existingData))
			newHash := fmt.Sprintf("%x", sha1.Sum(jsonData))
			if existingHash == newHash {
				return json.Unmarshal(existingData, &v)
			}
		}

		// Write the newly downloaded JSON to file
		fmt.Printf("Updating %s...\n", jsonFilePath)
		err = os.WriteFile(jsonFilePath, jsonData, 0644)
		if err != nil {
			return err
		}

		// Load the newly downloaded JSON data
		fmt.Printf("Reloading %s...\n", path)
		jsonStr := removeCommentsFromJSON(string(jsonData))
		err = json.Unmarshal([]byte(jsonStr), &v)
		if err != nil {
			return err
		}
	} else {
		// Load existing JSON data
		jsonData, err := os.ReadFile(jsonFilePath)
		if err != nil {
			return err
		}
		jsonStr := removeCommentsFromJSON(string(jsonData))
		err = json.Unmarshal([]byte(jsonStr), &v)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	headerWidth = 100
	separator   = ""
)

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

func checkForContent(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		printInfo(color.FgYellow, "%s directory not found\n", directory)
		return fmt.Errorf("%s directory not found", directory)
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() || len(info.Name()) != 8 {
			return nil
		}

		titleID := strings.ToLower(info.Name())
		titleData, ok := titles.Titles[titleID]
		if !ok {
			return nil
		}
		printHeader(titleData.TitleName)

		subDirDLC := filepath.Join(path, "$c")
		subInfoDLC, err := os.Stat(subDirDLC)
		if err == nil && subInfoDLC.IsDir() {
			err = processDLCContent(subDirDLC, titleData, titleID, directory)
			if err != nil {
				return err
			}
		} else {
			printInfo(color.FgYellow, "No DLC Found for %s..\n", titleID)
		}

		subDirUpdates := filepath.Join(path, "$u")
		subInfoUpdates, err := os.Stat(subDirUpdates)
		if err == nil && subInfoUpdates.IsDir() {
			err = processUpdates(subDirUpdates, titleData, titleID, directory)
			if err != nil {
				return err
			}
		} else {
			printInfo(color.FgYellow, "No Title Updates Found in $u for %s..\n", titleID)
		}

		return nil
	})
	return err
}
func processDLCContent(subDirDLC string, titleData TitleData, titleID string, directory string) error {
	subContents, err := os.ReadDir(subDirDLC)
	if err != nil {
		return err
	}

	for _, subContent := range subContents {
		subContentPath := filepath.Join(subDirDLC, subContent.Name())
		if !subContent.IsDir() {
			continue
		}

		subDirContents, err := os.ReadDir(subContentPath)
		if err != nil {
			return err
		}

		hasContentMetaXbx := false
		for _, dlcFiles := range subDirContents {
			if strings.Contains(strings.ToLower(dlcFiles.Name()), "contentmeta.xbx") {
				hasContentMetaXbx = true
				break
			}
		}

		if !hasContentMetaXbx {
			continue
		}

		contentID := strings.ToLower(subContent.Name())
		if !contains(titleData.ContentIDs, contentID) {
			printInfo(color.FgRed, "Unknown content found at: %s\n", subContentPath)
			continue
		}

		archivedName := ""
		for _, archived := range titleData.Archived {
			for archivedID, name := range archived {
				if archivedID == contentID {
					archivedName = name
					break
				}
			}
			if archivedName != "" {
				break
			}
		}

		subContentPath = strings.TrimPrefix(subContentPath, directory+"/")
		if archivedName != "" {
			printInfo(color.FgGreen, "Content is known and archived (%s)\n", archivedName)
		} else {
			printInfo(color.FgYellow, "%s has unarchived content found at: %s\n", titleData.TitleName, subContentPath)
		}
	}

	return nil
}
func processUpdates(subDirUpdates string, titleData TitleData, titleID string, directory string) error {
	// Load the ignore list from ignorelist.json in the data folder
	ignoreList, err := loadIgnoreList("data/ignorelist.json")
	if err != nil {
		printInfo(color.FgYellow, "Warning: error loading data/ignorelist.json: %s. Proceeding without ignore list.\n", err.Error())
		ignoreList = []string{} // Empty ignore list to prevent panics
	}

	files, err := os.ReadDir(subDirUpdates)
	if err != nil {
		return err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".xbe" {
			continue
		}

		// Skip if the file is in the ignore list
		if contains(ignoreList, f.Name()) {
			continue
		}

		filePath := filepath.Join(subDirUpdates, f.Name())
		fileHash, err := getSHA1Hash(filePath)
		if err != nil {
			printInfo(color.FgRed, "Error calculating hash for file: %s, error: %s\n", f.Name(), err.Error())
			continue
		}

		knownUpdateFound := false
		for _, knownUpdate := range titleData.TitleUpdatesKnown {
			for knownHash, name := range knownUpdate {
				if knownHash == fileHash {
					printHeader("File Info")
					printInfo(color.FgGreen, "Title update found for %s (%s) (%s)\n", titleData.TitleName, titleID, name)
					filePath = strings.TrimPrefix(filePath, directory+"/")
					printInfo(color.FgGreen, "Path: %s\n", filePath)
					printInfo(color.FgGreen, "SHA1: %s\n", fileHash)
					fmt.Println(separator)
					knownUpdateFound = true
					break
				}
			}
			if knownUpdateFound {
				break
			}
		}

		if !knownUpdateFound {
			printHeader("File Info")
			printInfo(color.FgRed, "Unknown Title Update found for %s (%s)\n", titleData.TitleName, titleID)
			filePath = strings.TrimPrefix(filePath, directory+"/")
			printInfo(color.FgRed, "Path: %s\n", filePath)
			printInfo(color.FgRed, "SHA1: %s\n", fileHash)
		}
	}

	return nil
}

func loadIgnoreList(filepath string) ([]string, error) {
	var ignoreList []string

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &ignoreList); err != nil {
		return nil, err
	}

	return ignoreList, nil
}

func getSHA1Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		//fmt.Printf("Comparing %q to %q\n", item, val)
		if item == val {
			return true
		}
	}
	return false
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
func promptForDownload(url string) bool {
	var response string
	fmt.Printf("The required JSON data is not found. It can be downloaded from %s\n", url)
	fmt.Print("Do you want to download it now? (yes/no): ")
	fmt.Scanln(&response)

	return strings.ToLower(response) == "yes"
}
func main() {
	flag.Parse() // Parse command line flags

	// Check for help flag
	if *helpFlag {
		fmt.Println("Usage of Pinecone:")
		fmt.Println("  -update: Update the JSON data from the source URL. If not set, uses local copies of data.")
		fmt.Println("  -summarize: Print summary statistics for all titles. If not set, checks for content in the TDATA folder.")
		fmt.Println("  -titleid: Filter statistics by Title ID (-titleID=ABCD1234). If not set, statistics are computed for all titles.")
		fmt.Println("  -fatxplorer: Use FATXPlorer's X drive as the root directory. If not set, runs as normal. (Windows Only)")
		fmt.Println("  -help: Display this help information.")
		return
	}
	jsonFilePath := "data/id_database.json"
	jsonDataFolder := "data"
	jsonURL := "https://api.github.com/repos/Xbox-Preservation-Project/Pinecone/contents/data/id_database.json"

	// Ensure data folder exists
	if _, err := os.Stat(jsonDataFolder); os.IsNotExist(err) {
		fmt.Println("Data folder not found. Creating...")
		if mkDirErr := os.Mkdir(jsonDataFolder, 0755); mkDirErr != nil {
			fmt.Println("Error creating data folder:", mkDirErr)
			return
		}
	}
	// Check if JSON file exists
	if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
		// Prompt for download if JSON file doesn't exist
		if promptForDownload(jsonURL) {
			err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", "data/id_database.json", &titles, true)
			if err != nil {
				fmt.Println("Error downloading data:", err)
				return
			}
		} else {
			fmt.Println("Download aborted by user.")
			return
		}
	} else if *updateFlag {
		// Handle manual update
		err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", "data/id_database.json", &titles, true)
		if err != nil {
			fmt.Println("Error updating data:", err)
			return
		}
	} else {
		// Load existing JSON data
		err := loadJSONData(jsonFilePath, "Xbox-Preservation-Project", "Pinecone", "data/id_database.json", &titles, false)
		if err != nil {
			fmt.Println("Error loading data:", err)
			return
		}
	}
	fmt.Println("Pinecone v0.4.2b")
	fmt.Println("Please share output of this program with the Pinecone team if you find anything interesting!")
	flag.Parse()

	if *titleIDFlag != "" {
		// if the titleID flag is set, print stats for that title
		printStats(*titleIDFlag, false)
	} else if *summarizeFlag {
		// if the summarize flag is set, print stats for all titles
		printStats("", true)
	} else if *fatxplorer {
		if runtime.GOOS == "windows" {
			if _, err := os.Stat(`X:\`); os.IsNotExist(err) {
				fmt.Println(`FatXplorer's X: drive not found`)
			} else {
				fmt.Println("Checking for Content...")
				fmt.Println("====================================================================================================")
				checkForContent("X:\\TDATA")
			}
		} else {
			fmt.Println("FatXplorer mode is only available on Windows.")
		}
	} else {
		// If no flag is set, proceed normally
		// Check if TDATA folder exists
		if _, err := os.Stat("dump/TDATA"); os.IsNotExist(err) {
			fmt.Println("TDATA folder not found. Please place TDATA folder in the dump folder.")
			return
		}
		fmt.Println("Checking for Content...")
		fmt.Println("====================================================================================================")
		err := checkForContent("dump/TDATA")
		if err != nil {
			panic(err)
		}
	}
}
