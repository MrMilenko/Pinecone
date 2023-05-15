package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	titles        TitleList
	updateFlag    = flag.Bool("update", false, "Update the JSON data from the source URL")
	summarizeFlag = flag.Bool("summarize", false, "Print summary statistics for all titles")
	titleIDFlag   = flag.String("titleid", "", "Filter statistics by Title ID")
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

	return ioutil.ReadAll(resp.Body)
}

func loadJSONData(jsonFilePath, owner, repo, path string, v interface{}, updateFlag bool) error {
	if updateFlag {

		//Notify we're checking for updates
		fmt.Printf("Checking for PineCone updates..\n")

		// Download JSON data
		jsonData, err := downloadJSONData(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path))

		// Check if downloaded JSON is different from existing JSON
		if _, err := os.Stat(jsonFilePath); err == nil {
			existingData, err := ioutil.ReadFile(jsonFilePath)
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
		err = ioutil.WriteFile(jsonFilePath, jsonData, 0644)
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
		jsonData, err := ioutil.ReadFile(jsonFilePath)
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
func checkForContent(directory string) error {
	// check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("%s directory not found", directory)
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && len(info.Name()) == 8 {
			titleID := strings.ToLower(info.Name())
			titleData, ok := titles.Titles[titleID]
			if ok {
				fmt.Printf("Found folder for \"%s\".\n", titleData.TitleName)

				subDirDLC := filepath.Join(path, "$c")
				subInfoDLC, err := os.Stat(subDirDLC)
				if err == nil && subInfoDLC.IsDir() {
					subContents, err := ioutil.ReadDir(subDirDLC)
					if err != nil {
						return err
					}

					foundUnarchivedContent := false
					var subContentPath string
					for _, subContent := range subContents {
						subContentPath = subDirDLC + "/" + subContent.Name()
						if subContent.IsDir() {
							contentID := strings.ToLower(subContent.Name())
							if contains(titleData.ContentIDs, contentID) {
								archivedContentID := strings.ToLower(contentID)
								var archivedName string
								for _, archived := range titleData.Archived {
									for archivedID, name := range archived {
										if archivedID == archivedContentID {
											archivedName = name
											break
										}
									}
									if archivedName != "" {
										break
									}
								}

								if archivedName != "" {
									fmt.Printf("%s content found at: %s is archived (%s).\n", titleData.TitleName, subContentPath, archivedName)
								} else {
									fmt.Printf("%s has unarchived content found at: %s\n", titleData.TitleName, subContentPath)
									foundUnarchivedContent = true
								}
							} else {
								fmt.Printf("%s unknown content found at: %s\n", titleData.TitleName, subContentPath)
							}
						}
					}

					if foundUnarchivedContent {
						files, err := ioutil.ReadDir(subContentPath)
						if err != nil {
							return err
						}

						for _, f := range files {
							if filepath.Ext(f.Name()) == ".xbe" || filepath.Ext(f.Name()) == ".xbx" {
								fmt.Println("Found content.. " + f.Name())
							} else {
								fmt.Println("Found unknown file format: " + f.Name())
							}
						}
					}
				} else {
					fmt.Printf("No DLC Found for %s..\n", titleID)
				}

				subDirUpdates := filepath.Join(path, "$u")
				subInfoUpdates, err := os.Stat(subDirUpdates)
				if err == nil && subInfoUpdates.IsDir() {
					//fmt.Printf("\nFolder found for \"%s\".\n", titleData.TitleName)

					files, err := ioutil.ReadDir(subDirUpdates)
					if err != nil {
						fmt.Println(err)
					}
					hashMatchFound := false
					for _, f := range files {
						if filepath.Ext(f.Name()) == ".xbe" || filepath.Ext(f.Name()) == ".xbx" {
							filePath := filepath.Join(subDirUpdates, f.Name())

							fileHash, err := getSHA1Hash(filePath)
							if err != nil {
								fmt.Printf("Error calculating hash for file %s: %s\n", f.Name(), err)
							} else {
								//fmt.Printf("Path: %s\n", filePath)
								//fmt.Printf("SHA1: %s\n", fileHash)
								for _, knownUpdate := range titleData.TitleUpdatesKnown {
									for knownHash, name := range knownUpdate {
										if knownHash == fileHash {
											fmt.Printf("Title update found for %s (%s) (%s)\n", titleData.TitleName, titleID, name)
											fmt.Printf("Path: %s\n", filePath)
											fmt.Printf("SHA1: %s\n", fileHash)
											fmt.Print("====================================================================================================\n")
											hashMatchFound = true
										}
									}
								}
								if !hashMatchFound {
									fmt.Printf("No SHA1 hash matches found for file %s\n", f.Name())
									fmt.Printf("SHA1 for unknown content: %s\n", fileHash)
									fmt.Printf("Path: %s\n", filePath)
									fmt.Print("====================================================================================================\n")
								}
							}
						}
					}
				} else {
					fmt.Printf("No Title Updates Found in $u for %s..\n", titleID)
				}
			} else {
				fmt.Printf("Title ID %s not present in JSON file.\n", titleID)
				fmt.Print("We found a folder with the correct format, but it's not in the JSON file.\n")
				fmt.Print("Please report this to the developer.\n")
				fmt.Print("Path: " + path + "\n")
				fmt.Print("====================================================================================================\n")
			}
		}
		return nil
	})

	return err
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
		for id, data := range titles.Titles {
			fmt.Printf("Statistics for title ID %s:\n", id)
			printTitleStats(&data)
		}
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

// Prints statistics for a TitleData.
func printTitleStats(data *TitleData) {
	fmt.Println("Title:", data.TitleName)
	fmt.Println("Total number of Content IDs:", len(data.ContentIDs))
	fmt.Println("Total number of Title Updates:", len(data.TitleUpdates))
	fmt.Println("Total number of Known Title Updates:", len(data.TitleUpdatesKnown))
	fmt.Println("Total number of Archived items:", len(data.Archived))
	fmt.Println()
}

func main() {
	// Check if TDATA folder exists
	if _, err := os.Stat("dump/TDATA"); os.IsNotExist(err) {
		fmt.Println("TDATA folder not found. Please place TDATA folder in the dump folder.")
		return
	}

	flag.Parse() // Parse command line flags

	// Check for help flag
	if *helpFlag {
		fmt.Println("Usage of Pinecone:")
		fmt.Println("  -update: Update the JSON data from the source URL. If not set, uses local copies of data.")
		fmt.Println("  -summarize: Print summary statistics for all titles. If not set, checks for content in the TDATA folder.")
		fmt.Println("  -titleid: Filter statistics by Title ID (-titleID=ABCD1234). If not set, statistics are computed for all titles.")
		fmt.Println("  -help: Display this help information.")
		return
	}

	// Load JSON data if update flag is set, otherwise use local copies
	err := loadJSONData("data/id_database.json", "OfficialTeamUIX", "Pinecone", "data/id_database.json", &titles, *updateFlag)
	if err != nil {
		panic(err)
	}
	fmt.Println("Pinecone v0.3")
	fmt.Println("Please share output of this program with the Pinecone team if you find anything interesting!")
	fmt.Println("====================================================================================================")

	if *titleIDFlag != "" {
		// if the titleID flag is set, print stats for that title
		printStats(*titleIDFlag, false)
	} else if *summarizeFlag {
		// if the summarize flag is set, print stats for all titles
		printStats("", true)
	} else {
		// if neither flag is set, proceed normally
		fmt.Println("Checking for Content...")
		fmt.Println("====================================================================================================")
		err = checkForContent("dump/TDATA")
		if err != nil {
			panic(err)
		}
	}
}
