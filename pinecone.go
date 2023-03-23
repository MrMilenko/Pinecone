package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TitleData struct {
	TitleName    string              `json:"Title Name"`
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

func main() {
	// Load JSON data
	fmt.Println("Local JSON file exists.")
	fmt.Println("Loading JSON data...")

	if _, err := os.Stat("id_database.json"); os.IsNotExist(err) {
		// File does not exist locally, so download it
		// Downloading from Milenko's GitHub repo for now.
		fmt.Println("Downloading JSON data...")
		owner := "MrMilenko"
		repo := "PineCone"
		path := "id_database.json"
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Accept", "application/vnd.github.v3.raw")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		out, err := os.Create("id_database.json")
		if err != nil {
			panic(err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			panic(err)
		}
	}

	jsonFile, err := os.Open("id_database.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	//fmt.Println(string(byteValue)) // Print out the JSON string

	jsonStr := removeCommentsFromJSON(string(byteValue))
	var titles TitleList
	err = json.Unmarshal([]byte(jsonStr), &titles)
	if err != nil {
		panic(err)
	}

	// Traverse directory structure
	fmt.Println("Traversing directory structure...")
	fmt.Println("Traversing directory structure...")

	err = filepath.Walk("TDATA", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// check if folder is in the correct format (TDATA\TitleID)
			if len(info.Name()) == 8 {
				// get title information from JSON
				titleID := strings.ToLower(info.Name())
				titleData, ok := titles.Titles[titleID]
				if ok {
					fmt.Printf("Found folder for \"%s\".\n", titleData.TitleName)
				} else {
					fmt.Printf("Title ID %s not present in JSON file. May want to investigate!\n", titleID)
				}
				//Uncomment for sloppy debug.
				//fmt.Printf("Content IDs for %s: %v\n", titleData.TitleName, titleData.ContentIDs)
				//fmt.Printf("Archived Content IDs for %s: %v\n", titleData.TitleName, titleData.Archived)
				// check for subfolders in the format of TDATA\TitleID\$c\ContentID
				subDir := filepath.Join(path, "$c")
				subInfo, err := os.Stat(subDir)
				if err == nil && subInfo.IsDir() {
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
							//Attemptiong to get SHA1 hash of the content
							//Scan for files in the folder
							files, err := ioutil.ReadDir(subContentPath)
							if err != nil {
								fmt.Println(err)
							}
							for _, f := range files {
								if filepath.Ext(f.Name()) == ".xbe" {
									fmt.Println("Found XBE file: " + f.Name())
								} else {
									fmt.Println("Found unknown file: " + f.Name())
								}
							}

							//Get SHA1 hash of the files
							//fmt.Println(getSHA1Hash(subContentPath))
						}

					}

				}
			}

		}
		return nil
	})

	// Traverse directory structure for Title Updates
	fmt.Println("Traversing directory structure for Title Updates...")
	err = filepath.Walk("TDATA", func(path string, info os.FileInfo, err error) error {
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
								fmt.Printf("%s: %s\n", filePath, fileHash)
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
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
