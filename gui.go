package main

import (
	"log"
	"github.com/gotk3/gotk3/gtk"
)

func startGUI() {
	// Initialize GTK
	gtk.Init(nil)

	// Create a new window
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Pinecone-GTK")
	win.SetDefaultSize(800, 400)
	win.SetResizable(false) // Set the window to be non-resizable
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Create a main grid for the window
	mainGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create main grid:", err)
	}
	win.Add(mainGrid)

	// Create a grid for the buttons
	buttonGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create button grid:", err)
	}

	// Create a scrolled window for the output
	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scrolled window:", err)
	}
	scrolledWindow.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

	// Create the output TextView
	output, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal("Unable to create textview:", err)
	}
	output.SetEditable(false)
	output.SetWrapMode(gtk.WRAP_WORD)
	outputBuffer, err := output.GetBuffer()
	if err != nil {
		log.Fatal("Unable to get text buffer:", err)
	}
	outputBuffer.InsertAtCursor("Output Window\n")

	// Add the TextView to the scrolled window
	scrolledWindow.Add(output)

	// Attach the scrolled window to the main grid
	mainGrid.Attach(scrolledWindow, 1, 0, 20, 1)

	// Create four buttons with tooltips and icons (currently placeholders, we can use stuff like xbox.png later)
	buttonIcons := []string{"folder-open-symbolic", "edit-find-symbolic", "media-playback-start-symbolic", "process-stop-symbolic"}
	buttonTooltips := []string{"Open Folder", "Find", "Play", "Stop"}

	for i := 0; i < 4; i++ {
		button, err := gtk.ButtonNew()
		if err != nil {
			log.Fatal("Unable to create button:", err)
		}

		// Load icon from theme
		icon, err := gtk.ImageNewFromIconName(buttonIcons[i], gtk.ICON_SIZE_BUTTON)
		if err != nil {
			log.Fatal("Unable to load icon:", err)
		}

		// Set tooltip for the button
		button.SetTooltipText(buttonTooltips[i])

		// Add icon to the button
		button.Add(icon)

		// Connect button click event
		button.Connect("clicked", func() {
			outputBuffer.InsertAtCursor(buttonTooltips[i] + " clicked\n")
		})

		// Attach the buttons to the button grid
		buttonGrid.Attach(button, 0, i, 1, 1)
	}

	// Attach the grids to the main grid
	mainGrid.Attach(buttonGrid, 0, 0, 1, 1)

	// Set expand properties for the output grid
	mainGrid.SetRowHomogeneous(true)
	mainGrid.SetColumnHomogeneous(true)

	// Show all widgets
	win.ShowAll()

	// Start the GTK main event loop
	gtk.Main()
}
