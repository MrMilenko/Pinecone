package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var outputContainer = container.New(layout.NewVBoxLayout())

var (
	Red    = color.RGBA{255, 0, 0, 255}
	Yellow = color.RGBA{255, 255, 0, 255}
	Green  = color.RGBA{0, 255, 0, 255}
)

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

const (
	guiHeaderWidth = 50
)

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

func addHeader(title string) {
	title = strings.TrimSpace(title)
	if len(title) > guiHeaderWidth-6 { // -6 to account for spaces and equals signs
		title = title[:guiHeaderWidth-4] + "..."
	}
	formattedTitle := "== " + title + " =="
	padLen := (guiHeaderWidth - len(formattedTitle)) / 2
	addText(theme.ForegroundColor(), strings.Repeat("=", padLen)+formattedTitle+strings.Repeat("=", guiHeaderWidth-padLen-len(formattedTitle)))
}

func addText(textColor color.Color, format string, args ...interface{}) {
	output := canvas.NewText(fmt.Sprintf(format, args...), textColor)
	outputContainer.Add(output)
	outputContainer.Refresh()
	outputContainer.Show()
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
		canvas.NewText("User Info:", theme.ForegroundColor()),
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
		// Set tmpDumpPath to the selected folder
		tmpDumpPath := list.String()
		// Convert path to be used in the checkForContent function
		tmpDumpPath = strings.Replace(tmpDumpPath, "file://", "", -1)
		// set global scanpath variable to the selected folder
		// output := widget.NewLabel("")

		if _, err := os.Stat(path.Join(tmpDumpPath + "TDATA")); os.IsNotExist(err) {
			dumpLocation = tmpDumpPath
			// We don't want usernames in the log, so we'll just show the folder name AFTER passing it to checkForContent
			// path = strings.Replace(path, homeDir, "", -1)
			output := canvas.NewText("Path set to: "+tmpDumpPath, theme.ForegroundColor())
			outputContainer.Add(output)
		} else {
			output := canvas.NewText("Incorrect pathing. Please select a dump with TDATA folder.\n", theme.ForegroundColor())
			outputContainer.Add(output)
		}
		// outputContainer.Add(output)
	}, window)
}

func guiScanContent(options GUIOptions) {
	if dumpLocation == "" {
		output := canvas.NewText("Please set a path first.", theme.ForegroundColor())
		outputContainer.Add(output)
	} else {
		path := dumpLocation
		output := canvas.NewText("Checking for Content...", theme.ForegroundColor())
		outputContainer.Add(output)
		err := checkDatabaseFile(options.JSONFilePath, options.JSONUrl, updateFlag)
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			addText(Red, err.Error())
		}

		err = checkDumpFolder(path)
		if nil != err {
			fmt.Println("ERROR: ", err.Error())
			addText(Red, err.Error())
		}

		err = checkParsingSettings()
		if nil != err {
			fmt.Println("ERROR: ", err.Error())
			addText(Red, err.Error())
		}
	}
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
		err = os.MkdirAll(outputDir, 0o755)
		if err != nil {
			panic(err)
		}
	}
	// Write output to file
	fileText := ""
	for _, obj := range outputContainer.Objects {
		if textObj, ok := obj.(*canvas.Text); ok {
			// Append the text value to the string
			fileText += textObj.Text
		}
	}
	err := os.WriteFile(outputPath, []byte(fileText), 0o644)
	if err != nil {
		panic(err)
	}
	// Debug output, show the path we're scanning
	output := widget.NewLabel("Output saved to: " + outputPath + "\n")
	outputContainer.Add(output)
}

func startGUI(options GUIOptions) {
	a := app.New()
	w := a.NewWindow("Pinecone 0.5.0")
	output := widget.NewLabel("")

	// First Load welcome message
	fakeConsole := fmt.Sprintln("Welcome to Pinecone v0.5.0b")
	output.SetText(output.Text + fakeConsole)

	w.Resize(fyne.Size{Width: 800, Height: 600})

	var (
		tdataButtonIcon = loadImage("tdataButton", options.DataFolder+"/buttons/xboxS.svg")
		exitButtonIcon  = loadImage("exitButton", options.DataFolder+"/buttons/exit.svg")
	)

	// set folder to scan, but only if it is a TDATA folder.
	setFolder := widget.NewButtonWithIcon("Set Dump Folder", tdataButtonIcon, func() {
		setDumpFolder(w)
	})

	scanPath := widget.NewButtonWithIcon("Scan For Content", theme.SearchIcon(), func() {
		guiScanContent(options)
	})
	// Save output to a file in the homeDir with a timestamp.
	saveOutput := widget.NewButtonWithIcon("Save Output", theme.DocumentSaveIcon(), func() {
		saveOutput()
	})

	updateJSON := widget.NewButtonWithIcon("Update Database", theme.DownloadIcon(), func() {
		updateJSON := true
		err := checkDatabaseFile(options.JSONFilePath, options.JSONUrl, updateJSON)
		if err != nil {
			fmt.Println(err)
		}
	})
	// Create the settings button with the settings icon
	settingsButton := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		// Open the settings screen
		settings, err := loadSettings()
		if err != nil {
			fmt.Println(err)
			settings = &Settings{}
		}
		showSettingsDialog(w, settings, a)
	})

	// Exit the application
	exit := widget.NewButtonWithIcon("Exit", exitButtonIcon, func() {
		a.Quit()
	})

	// Create a container with vertical box layout for the hamburger menu
	sideMenu := container.NewVBox()

	// Create a container with vertical box layout for the buttons
	buttons := container.NewVBox(setFolder, scanPath, updateJSON, saveOutput, settingsButton, exit)

	// Add the hamburger button to the hamburgerMenu
	sideMenu.Add(buttons)

	outputContainer.Add(output)
	// Create a container with scroll for the output
	outputScroll := container.NewScroll(outputContainer)

	// Create a container to hold the main content of the window
	mainContent := container.NewBorder(nil, nil, nil, nil, outputScroll)

	// Create a container that includes the hamburger menu and main content
	fullContent := container.NewBorder(nil, nil, sideMenu, nil, mainContent)

	// Place the buttons to the left and the output to the center
	w.SetContent(fullContent)
	w.ShowAndRun()
}
