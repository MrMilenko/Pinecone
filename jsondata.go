package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

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
		if err != nil {
			return err
		}

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
