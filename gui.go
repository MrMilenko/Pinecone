package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var output = widget.NewMultiLineEntry()

type GUIOptions struct {
	DataFolder   string
	JSONFilePath string
	JSONUrl      string
}

type Settings struct {
	UserName                   string `json:"username"`
	Discord                    string `json:"discord"`
	Twitter                    string `json:"twitter"`
	Reddit                     string `json:"reddit"`
	EnableDiscordNotifications bool   `json:"enableDiscordNotifications"`
}

func loadImage(name, path string) *fyne.StaticResource {
	imgBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return &fyne.StaticResource{
		StaticName:    name,
		StaticContent: imgBytes,
	}
}

func loadSettings() (*Settings, error) {
	settingsPath := filepath.Join("pineconeSettings.json")
	settingsFile, err := os.Open(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the settings file doesn't exist, return default settings
			return &Settings{}, nil
		}
		return nil, err
	}
	defer settingsFile.Close()

	settings := &Settings{}
	err = json.NewDecoder(settingsFile).Decode(settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func saveSettings(settings *Settings) error {
	settingsPath := filepath.Join("pineconeSettings.json")
	settingsFile, err := os.Create(settingsPath)
	if err != nil {
		return err
	}
	defer settingsFile.Close()

	encoder := json.NewEncoder(settingsFile)
	encoder.SetIndent("", "    ")

	err = encoder.Encode(settings)
	if err != nil {
		return err
	}

	return nil
}

func showSettingsDialog(parent fyne.Window, settings *Settings, app fyne.App) {
	settingsWindow := app.NewWindow("Settings")
	settingsWindow.Resize(fyne.Size{Width: 200, Height: 100})

	userNameEntry := widget.NewEntry()
	userNameEntry.SetPlaceHolder("User Name")
	userNameEntry.SetText(settings.UserName)
	userNameEntry.OnChanged = func(text string) {
		settings.UserName = text
	}

	discordEntry := widget.NewEntry()
	discordEntry.SetPlaceHolder("Discord")
	discordEntry.SetText(settings.Discord)
	discordEntry.OnChanged = func(text string) {
		settings.Discord = text
	}

	twitterEntry := widget.NewEntry()
	twitterEntry.SetPlaceHolder("Twitter")
	twitterEntry.SetText(settings.Twitter)
	twitterEntry.OnChanged = func(text string) {
		settings.Twitter = text
	}

	redditEntry := widget.NewEntry()
	redditEntry.SetPlaceHolder("Reddit")
	redditEntry.SetText(settings.Reddit)
	redditEntry.OnChanged = func(text string) {
		settings.Reddit = text
	}

	enableDiscordCheckbox := widget.NewCheck("Enable Discord Notifications", func(checked bool) {
		settings.EnableDiscordNotifications = checked

		err := saveSettings(settings)
		if err != nil {
			dialog.ShowError(err, settingsWindow)
			return
		}
	})

	enableDiscordCheckbox.SetChecked(settings.EnableDiscordNotifications)

	saveButton := widget.NewButton("Save", func() {
		err := saveSettings(settings)
		if err != nil {
			dialog.ShowError(err, settingsWindow)
			return
		}
		settingsWindow.Close()
	})

	cancelButton := widget.NewButton("Cancel", func() {
		settingsWindow.Close()
	})

	content := container.NewVBox(
		widget.NewLabel("User Info:"),
		userNameEntry,
		discordEntry,
		twitterEntry,
		redditEntry,
		enableDiscordCheckbox,
		container.NewHBox(
			layout.NewSpacer(),
			saveButton,
			cancelButton,
		),
	)

	settingsWindow.SetContent(content)
	settingsWindow.Show()
}

func setDumpFolder(window fyne.Window) {
	dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if list == nil { // user cancelled
			return
		}
		// Set path to the selected folder
		var path = list.String()
		// Convert path to be used in the checkForContent function
		path = strings.Replace(path, "file://", "", -1)
		// set global scanpath variable to the selected folder
		if strings.Contains(path, "TDATA") {
			dumpLocation = path
			// We don't want usernames in the log, so we'll just show the folder name AFTER passing it to checkForContent
			// path = strings.Replace(path, homeDir, "", -1)
			output.SetText(output.Text + "Path set to: " + path + "\n")
		} else {
			output.SetText(output.Text + "Incorrect pathing. Please select a TDATA folder.\n")
		}
	}, window)
}

func saveOutput() {
	// Get current time
	t := time.Now()
	// Format time to be used in filename
	timestamp := t.Format("2006-01-02-15-04-05")
	// Define the path to the output file
	outputPath := filepath.Join("output", "output-"+timestamp+".txt")
	// Create the 'output' directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			panic(err)
		}
	}
	// Write output to file
	err := os.WriteFile(outputPath, []byte(output.Text), 0644)
	if err != nil {
		panic(err)
	}
	// Debug output, show the path we're scanning
	output.SetText(output.Text + "Output saved to: " + outputPath + "\n")

}

func startGUI(options GUIOptions) {
	a := app.New()
	w := a.NewWindow("Pinecone 0.5.0")

	// First Load welcome message
	fakeConsole := fmt.Sprintln("Welcome to Pinecone v0.5.0b")
	output.SetText(output.Text + fakeConsole)

	w.Resize(fyne.Size{Width: 800, Height: 300})

	var (
		tdataButtonIcon      = loadImage("tdataButton", options.DataFolder+"/buttons/xboxS.svg")
		searchButtonIcon     = loadImage("searchButton", options.DataFolder+"/buttons/search.svg")
		updateJSONButtonIcon = loadImage("updateJSONButton", options.DataFolder+"/buttons/cloud-down.svg")
		saveButtonIcon       = loadImage("saveButton", options.DataFolder+"/buttons/floppy-disk.svg")
		settingsButtonIcon   = loadImage("settingsButton", options.DataFolder+"/buttons/gear.svg")
		exitButtonIcon       = loadImage("exitButton", options.DataFolder+"/buttons/exit.svg")
	)

	// set folder to scan, but only if it is a TDATA folder.
	setFolder := widget.NewButtonWithIcon("", tdataButtonIcon, func() {
		setDumpFolder(w)
	})

	scanPath := widget.NewButtonWithIcon("", searchButtonIcon, func() {
		if dumpLocation == "" {
			output.SetText(output.Text + "Please set a path first.\n")
		} else {
			path := dumpLocation
			output.SetText(output.Text + "Checking for Content...\n")
			err := checkForContent(path)
			if nil != err {
				
			}
		}
	})
	// Save output to a file in the homeDir with a timestamp.
	saveOutput := widget.NewButtonWithIcon("", saveButtonIcon, func() {
		saveOutput()
	})

	updateJSON := widget.NewButtonWithIcon("", updateJSONButtonIcon, func() {
		var updateJSON = true
		err := checkDatabaseFile(options.JSONFilePath, options.JSONUrl, updateJSON)
		if err != nil {
			log.Fatalln(err)
		}
	})
	// Create the settings button with the settings icon
	settingsButton := widget.NewButtonWithIcon("", settingsButtonIcon, func() {
		// Open the settings screen
		settings, err := loadSettings()
		if err != nil {
			log.Println(err)
			settings = &Settings{}
		}
		showSettingsDialog(w, settings, a)
	})

	// Exit the application
	exit := widget.NewButtonWithIcon("", exitButtonIcon, func() {
		a.Quit()
	})

	// Create a container with vertical box layout for the hamburger menu
	sideMenu := container.NewVBox()

	// Create a container with vertical box layout for the buttons
	buttons := container.NewVBox(setFolder, scanPath, updateJSON, saveOutput, settingsButton, exit)

	// Add the hamburger button to the hamburgerMenu
	sideMenu.Add(buttons)

	// Create a container with scroll for the output
	outputScroll := container.NewScroll(output)

	// Create a container to hold the main content of the window
	mainContent := container.NewBorder(nil, nil, nil, nil, outputScroll)

	// Create a container that includes the hamburger menu and main content
	fullContent := container.NewBorder(nil, nil, sideMenu, nil, mainContent)

	// Place the buttons to the left and the output to the center
	w.SetContent(fullContent)
	w.ShowAndRun()
}
