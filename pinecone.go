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
	titles       TitleList
	updateFlag   = flag.Bool("update", false, "Update the JSON data")
	ugcFlag      = flag.Bool("ugc", false, "Search for User Generated Content")
	homebrewFlag = flag.Bool("homebrew", false, "Search for Homebrew Content")
)

var xboxExtensions = []string{
	// Modify this list to add/remove extensions to search for
	// Xbox Specific Extensions
	".xpr",
	".xbx",
	".xip",
	".xap",
	// Generic extensions
	".zip",
	".rar",
	".py",
	".txt",
	".cfg",
	".ini",
	".nfo",
	".xml",
}

type TitleData struct {
	TitleName    string              `json:"Title Name,"`
	ContentIDs   []string            `json:"Content IDs"`
	TitleUpdates []string            `json:"Title Updates"`
	Archived     []map[string]string `json:"Archived"`
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

func checkForDLC() error {
	return filepath.Walk("dump/TDATA", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// check if folder is in the correct format (TDATA\TitleID)
			if len(info.Name()) == 8 {
				// check for subfolders in the format of TDATA\TitleID\$c\ContentID
				subDir := filepath.Join(path, "$c")
				subInfo, err := os.Stat(subDir)
				if err == nil && subInfo.IsDir() {
					// get title information from JSON
					titleID := strings.ToLower(info.Name())
					titleData, ok := titles.Titles[titleID]
					if ok {
						fmt.Printf("Found folder for \"%s\".\n", titleData.TitleName)
						fmt.Printf("Found DLC for \"%s\".\n", titleData.TitleName)
					} else {
						fmt.Printf("Title ID %s not present in JSON file. May want to investigate!\n", titleID)
					}
					// scan for content folders within the $c directory
					subContents, err := ioutil.ReadDir(subDir)
					if err == nil {
						foundUnarchivedContent := false
						var subContentPath string // declare subContentPath outside the for loop
						for _, subContent := range subContents {
							subContentPath = subDir + "/" + subContent.Name()
							if subContent.IsDir() {
								contentID := strings.ToLower(subContent.Name())
								if contains(titleData.ContentIDs, contentID) {
									// check if content is archived
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
							// Attempting to get SHA1 hash of the content
							// Scan for files in the folder
							files, err := ioutil.ReadDir(subContentPath)
							if err != nil {
								fmt.Println(err)
							}
							for _, f := range files {
								if filepath.Ext(f.Name()) == ".xbe" || filepath.Ext(f.Name()) == ".xbx" {
									fmt.Println("Found content.. " + f.Name())
								} else {
									fmt.Println("Found unknown file format: " + f.Name())
								}
							}

							// Get SHA1 hash of the files
							// fmt.Println(getSHA1Hash(subContentPath))
						}

					}
				}
			}
		}
		return nil
	})
}

func checkForUpdates(directory string) error {
	// check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("%s directory not found", directory)
	}
	// traverse directory structure
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// check if folder is in the correct format (TDATA\TitleID)
			if len(info.Name()) == 8 {
				// check for subfolders in the format of TDATA\TitleID\$u
				subDir := filepath.Join(path, "$u")
				subInfo, err := os.Stat(subDir)
				if err == nil && subInfo.IsDir() {
					// get title information from JSON
					titleID := strings.ToLower(info.Name())
					titleData, ok := titles.Titles[titleID]
					if ok {
						fmt.Printf("Found Possible Title Updates for \"%s\".\n", titleData.TitleName)
						// scan for XBE files within the $u directory
						files, err := ioutil.ReadDir(subDir)
						if err != nil {
							fmt.Println(err)
						}
						for _, f := range files {
							if filepath.Ext(f.Name()) == ".xbe" || filepath.Ext(f.Name()) == ".xbx" {
								filePath := filepath.Join(subDir, f.Name())
								fileHash, err := getSHA1Hash(filePath)
								if err != nil {
									fmt.Printf("Error calculating hash for file %s: %s\n", f.Name(), err)
								} else {
									fmt.Printf("Path: %s\n", filePath)
									fmt.Printf("SHA1: %s\n", fileHash)
								}
							}
						}
					} else {
						fmt.Printf("No Title Updates Found in $u for %s..\n", titleID)
					}
				}
			}
		}
		return nil
	})
	return err
}

func checkForUGC(directory string) error {
	// check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return fmt.Errorf("%s directory not found", directory)
	}
	fmt.Println("\nChecking for user generated content..")
	// traverse directory structure
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// check if folder is in the correct format (TDATA\TitleID)
			if len(info.Name()) == 8 {
				// check for subfolders in the format of TDATA\TitleID\$k and TDATA\TitleID\$k\Subfolder
				ySubDir := filepath.Join(path, "$y")
				ySubInfo, err := os.Stat(ySubDir)
				if err == nil && ySubInfo.IsDir() {
					titleID := strings.ToLower(info.Name())
					fmt.Printf("Found Possible user made content for: \"%s\".\n", titleID)

					// loop through files in the $k subdirectory
					kFiles, err := ioutil.ReadDir(ySubDir)
					if err != nil {
						fmt.Println(err)
					}
					for _, f := range kFiles {
						if f.Mode().IsRegular() {
							filePath := filepath.Join(ySubDir, f.Name())
							fileHash, err := getSHA1Hash(filePath)
							if err != nil {
								fmt.Printf("Error calculating hash for file %s: %s\n", f.Name(), err)
							} else {
								fmt.Printf("Path: %s\n", filePath)
								fmt.Printf("SHA1: %s\n", fileHash)
							}
						}
					}

					// loop through files in subdirectories of the $k subdirectory
					ySubDirs, err := ioutil.ReadDir(ySubDir)
					if err != nil {
						fmt.Println(err)
					}
					for _, subDir := range ySubDirs {
						if subDir.IsDir() {
							subPath := filepath.Join(ySubDir, subDir.Name())
							subFiles, err := ioutil.ReadDir(subPath)
							if err != nil {
								fmt.Println(err)
							}
							for _, f := range subFiles {
								if f.Mode().IsRegular() {
									filePath := filepath.Join(subPath, f.Name())
									fileHash, err := getSHA1Hash(filePath)
									if err != nil {
										fmt.Printf("Error calculating hash for file %s: %s\n", f.Name(), err)
									} else {
										fmt.Printf("Path: %s\n", filePath)
										fmt.Printf("SHA1: %s\n", fileHash)
									}
								}
							}
						}
					}
				}
			}
		}
		return nil
	})
	return err
}
func checkForHomebrew(rootDir string) error {
	fmt.Println("\nSearching for Homebrew content..")
	// traverse directory structure
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// skip TDATA and UDATA directories
			if info.Name() == "TDATA" || info.Name() == "UDATA" {
				return filepath.SkipDir
			}

		} else if info.Mode().IsRegular() {
			// check if file is known Xbox homebrew file type(s)
			if contains(xboxExtensions, filepath.Ext(info.Name())) {
				filePath := path
				fileHash, err := getSHA1Hash(filePath)
				if err != nil {
					fmt.Printf("Error calculating hash for file %s: %s\n", info.Name(), err)
				} else {
					fmt.Printf("Path: %s\n", filePath)
					fmt.Printf("SHA1: %s\n", fileHash)
				}
			}
		}
		return nil
	})
	return err
}

func main() {
	// Check if TDATA folder exists
	if _, err := os.Stat("dump/TDATA"); os.IsNotExist(err) {
		fmt.Println("TDATA folder not found. Please place TDATA folder in the dump folder.")
		return
	}

	flag.Parse() // Parse command line flags

	// Load JSON data if update flag is set, otherwise use local copies
	err := loadJSONData("data/id_database.json", "MrMilenko", "PineCone", "id_database.json", &titles, *updateFlag)
	if err != nil {
		panic(err)
	}

	// Traverse directory structure
	fmt.Println("Checking for DLC...")
	err = checkForDLC()
	if err != nil {
		panic(err)
	}

	fmt.Println("\nChecking for Title Updates...")
	err = checkForUpdates("dump/TDATA")
	if err != nil {
		panic(err)
	}

	// check for user generated content
	if *ugcFlag {
		err := checkForUGC("dump/TDATA")
		if err != nil {
			fmt.Println(err)
		}
	}
	// check homebrew directories
	if *homebrewFlag {
		checkForHomebrew("dump/")
	}
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
