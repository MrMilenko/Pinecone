package main

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
