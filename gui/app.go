package gui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type displayText struct {
	hour       *canvas.Text
	colonLeft  *canvas.Text
	minute     *canvas.Text
	colonRight *canvas.Text
	second     *canvas.Text
	background *canvas.Rectangle
	root       fyne.CanvasObject
}

type appAppearance struct {
	application      fyne.App
	preferences      fyne.Preferences
	displays         []*displayText
	selectedTheme    string
	selectedColor    string
	selectedColon    string
	selectedFontSize string
}

const (
	prefThemeKey      = "appearance.theme"
	prefColorKey      = "appearance.color"
	prefColonColorKey = "appearance.colonColor"
	prefSizeKey       = "appearance.size"
	prefWindowWidth   = "window.width"
	prefWindowHeight  = "window.height"
	defaultTheme      = "Dark"
	defaultColor      = "Green"
	defaultColonColor = "White"
	defaultFontSize   = "Large"
	defaultWinWidth   = 720
	defaultWinHeight  = 560
)

var displayColors = map[string]color.NRGBA{
	"Green":  {R: 74, G: 222, B: 128, A: 255},
	"Red":    {R: 248, G: 113, B: 113, A: 255},
	"Yellow": {R: 250, G: 204, B: 21, A: 255},
	"Blue":   {R: 96, G: 165, B: 250, A: 255},
	"White":  {R: 241, G: 245, B: 249, A: 255},
}

var displayColorOptions = []string{"Green", "Red", "Yellow", "Blue", "White"}
var colonColorOptions = []string{"Green", "Red", "Yellow", "Blue", "White"}

var displayBackgroundColors = map[string]color.NRGBA{
	"Dark":  {R: 15, G: 23, B: 42, A: 255},
	"Light": {R: 226, G: 232, B: 240, A: 255},
}

var themeOptions = []string{"Dark", "Light"}

var displayFontSizes = map[string]float32{
	"Small":    44,
	"Medium":   64,
	"Large":    90,
	"X-Large":  120,
	"XX-Large": 200,
}

var displaySizeOptions = []string{"Small", "Medium", "Large", "X-Large", "XX-Large"}

func Run() error {
	application := app.NewWithID("github.com.FaridBerlin.timepulse")
	window := application.NewWindow("Timepulse")
	appearance := newAppAppearance(application)

	// Restore previous window size if saved, otherwise use defaults.
	pref := application.Preferences()
	w := pref.IntWithFallback(prefWindowWidth, defaultWinWidth)
	h := pref.IntWithFallback(prefWindowHeight, defaultWinHeight)

	clock := newClockTab(appearance)
	stopwatch := newStopwatchTab(appearance)
	timer := newTimerTab(appearance)

	tabs := container.NewAppTabs(
		container.NewTabItem("Clock", clock),
		container.NewTabItem("Stopwatch", stopwatch),
		container.NewTabItem("Timer", timer),
	)
	window.SetContent(container.NewBorder(newHeaderControls(appearance, window), nil, nil, nil, tabs))
	// fyne.NewSize expects float32 values.
	window.Resize(fyne.NewSize(float32(w), float32(h)))
	window.CenterOnScreen()

	// Persist window size on close.
	window.SetCloseIntercept(func() {
		sz := window.Canvas().Size()
		pref.SetInt(prefWindowWidth, int(sz.Width))
		pref.SetInt(prefWindowHeight, int(sz.Height))
		window.Close()
	})

	window.ShowAndRun()
	return nil
}

func newClockTab(appearance *appAppearance) fyne.CanvasObject {
	display := newDisplayText()
	appearance.RegisterDisplay(display)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case current := <-ticker.C:
				display.SetText(current.Format("15:04:05"))
			case <-stop:
				return
			}
		}
	}()
	return container.NewCenter(display.Object())
}

func newStopwatchTab(appearance *appAppearance) fyne.CanvasObject {
	display := newDisplayText()
	appearance.RegisterDisplay(display)
	startButton := widget.NewButton("Start", nil)
	resetButton := widget.NewButton("Reset", nil)
	var mutex sync.Mutex
	var startedAt time.Time
	var elapsed time.Duration
	var running bool

	startButton.OnTapped = func() {
		mutex.Lock()
		defer mutex.Unlock()
		if running {
			elapsed += time.Since(startedAt)
			running = false
			startButton.SetText("Start")
			return
		}
		startedAt = time.Now()
		running = true
		startButton.SetText("Pause")
	}
	resetButton.OnTapped = func() {
		mutex.Lock()
		defer mutex.Unlock()
		elapsed = 0
		if running {
			startedAt = time.Now()
		}
		display.SetText(formatDuration(0))
	}

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			mutex.Lock()
			current := elapsed
			if running {
				current += time.Since(startedAt)
			}
			mutex.Unlock()
			display.SetText(formatDuration(current))
		}
	}()
	controls := container.NewCenter(container.NewHBox(startButton, resetButton))
	return container.NewBorder(display.Object(), controls, nil, nil)
}

func newTimerTab(appearance *appAppearance) fyne.CanvasObject {
	display := newDisplayText()
	appearance.RegisterDisplay(display)
	input := widget.NewEntry()
	input.SetPlaceHolder("HH:MM:SS (default 00:05:00)")
	startButton := widget.NewButton("Start", nil)
	resetButton := widget.NewButton("Reset", nil)
	var mutex sync.Mutex
	var endAt time.Time
	var remaining time.Duration
	var running bool

	reset := func() {
		mutex.Lock()
		remaining = 5 * time.Minute
		running = false
		mutex.Unlock()
		display.SetText(formatDuration(remaining))
		startButton.SetText("Start")
	}
	reset()

	startButton.OnTapped = func() {
		mutex.Lock()
		defer mutex.Unlock()
		if running {
			remaining = time.Until(endAt)
			if remaining < 0 {
				remaining = 0
			}
			running = false
			startButton.SetText("Start")
			return
		}
		remaining = parseTimerInput(input.Text)
		endAt = time.Now().Add(remaining)
		running = true
		startButton.SetText("Pause")
	}
	resetButton.OnTapped = reset

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			mutex.Lock()
			if running {
				remaining = time.Until(endAt)
				if remaining <= 0 {
					remaining = 0
					running = false
					startButton.SetText("Start")
				}
			}
			current := remaining
			mutex.Unlock()
			display.SetText(formatDuration(current))
		}
	}()

	controls := container.NewVBox(input, container.NewHBox(startButton, resetButton))
	return container.NewBorder(display.Object(), controls, nil, nil)
}

func newAppAppearance(application fyne.App) *appAppearance {
	preferences := application.Preferences()
	appearance := &appAppearance{
		application:      application,
		preferences:      preferences,
		selectedTheme:    pickAllowed(preferences.StringWithFallback(prefThemeKey, defaultTheme), themeOptions, defaultTheme),
		selectedColor:    pickAllowed(preferences.StringWithFallback(prefColorKey, defaultColor), displayColorOptions, defaultColor),
		selectedColon:    pickAllowed(preferences.StringWithFallback(prefColonColorKey, defaultColonColor), colonColorOptions, defaultColonColor),
		selectedFontSize: pickAllowed(preferences.StringWithFallback(prefSizeKey, defaultFontSize), displaySizeOptions, defaultFontSize),
	}
	appearance.SetTheme(appearance.selectedTheme)
	return appearance
}

func newAppearanceControls(appearance *appAppearance) fyne.CanvasObject {
	themePicker := widget.NewSelect(themeOptions, func(name string) {
		appearance.SetTheme(name)
	})
	themePicker.SetSelected(appearance.selectedTheme)

	colorPicker := widget.NewSelect(displayColorOptions, func(name string) {
		appearance.SetColor(name)
	})
	colorPicker.SetSelected(appearance.selectedColor)

	colonPicker := widget.NewSelect(colonColorOptions, func(name string) {
		appearance.SetColonColor(name)
	})
	colonPicker.SetSelected(appearance.selectedColon)

	sizePicker := widget.NewSelect(displaySizeOptions, func(name string) {
		appearance.SetFontSize(name)
	})
	sizePicker.SetSelected(appearance.selectedFontSize)

	resetButton := widget.NewButton("Reset", func() {
		appearance.ResetDefaults()
		themePicker.SetSelected(appearance.selectedTheme)
		colorPicker.SetSelected(appearance.selectedColor)
		colonPicker.SetSelected(appearance.selectedColon)
		sizePicker.SetSelected(appearance.selectedFontSize)
	})

	// Add some spacing between groups for a cleaner layout.
	spacer := widget.NewSeparator()

	controls := container.NewHBox(
		widget.NewLabel("Theme"), themePicker,
		spacer,
		widget.NewLabel("Color"), colorPicker,
		spacer,
		widget.NewLabel("Colon"), colonPicker,
		spacer,
		widget.NewLabel("Size"), sizePicker,
		spacer,
		resetButton,
	)
	return container.NewPadded(controls)
}

// newHeaderControls creates a top‑bar with a hamburger button that opens a pop‑up
// containing the appearance controls (theme, colour, colon colour, size and reset).
// The pop‑up is anchored to the window's canvas and disappears when the user
// clicks outside of it.
func newHeaderControls(appearance *appAppearance, win fyne.Window) fyne.CanvasObject {
	// Hamburger button (Unicode character for menu)
	menuButton := widget.NewButton("☰", nil)

	// When the button is tapped, display the appearance controls in a popup.
	menuButton.OnTapped = func() {
		// Wrap the appearance controls in a container to give it a background.
		content := container.NewVBox(newAppearanceControls(appearance))
		// Create the popup attached to the window's canvas.
		pop := widget.NewPopUp(content, win.Canvas())
		// Center the popup under the button. We use the button's position on the
		// canvas to calculate an offset. If the position cannot be determined,
		// Show places it near the cursor.
		pos := menuButton.Position()
		// Position the popup just below the menu button.
		pop.Move(pos.Add(fyne.NewPos(0, menuButton.Size().Height)))
		pop.Show()
	}

	// Align the button to the left with some padding.
	return container.NewHBox(menuButton)
}

func (a *appAppearance) RegisterDisplay(display *displayText) {
	a.displays = append(a.displays, display)
	display.SetColor(a.selectedColor)
	display.SetColonColor(a.selectedColon)
	display.SetFontSize(a.selectedFontSize)
	display.SetTheme(a.selectedTheme)
}

func (a *appAppearance) SetTheme(name string) {
	if name != "Dark" && name != "Light" {
		name = "Dark"
	}
	a.selectedTheme = name
	a.preferences.SetString(prefThemeKey, name)
	if name == "Dark" {
		a.application.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.application.Settings().SetTheme(theme.LightTheme())
	}
	for _, display := range a.displays {
		display.SetTheme(name)
	}
}

func (a *appAppearance) SetColor(name string) {
	name = pickAllowed(name, displayColorOptions, defaultColor)
	a.selectedColor = name
	a.preferences.SetString(prefColorKey, name)
	for _, display := range a.displays {
		display.SetColor(name)
	}
}

func (a *appAppearance) SetColonColor(name string) {
	name = pickAllowed(name, colonColorOptions, defaultColonColor)
	a.selectedColon = name
	a.preferences.SetString(prefColonColorKey, name)
	for _, display := range a.displays {
		display.SetColonColor(name)
	}
}

func (a *appAppearance) SetFontSize(name string) {
	name = pickAllowed(name, displaySizeOptions, defaultFontSize)
	a.selectedFontSize = name
	a.preferences.SetString(prefSizeKey, name)
	for _, display := range a.displays {
		display.SetFontSize(name)
	}
}

func (a *appAppearance) ResetDefaults() {
	a.SetTheme(defaultTheme)
	a.SetColor(defaultColor)
	a.SetColonColor(defaultColonColor)
	a.SetFontSize(defaultFontSize)
}

func newDisplayText() *displayText {
	hour := newTimeSegment("00", displayColors["Green"], displayFontSizes["Large"])
	colonLeft := newTimeSegment(":", displayColors["White"], displayFontSizes["Large"])
	minute := newTimeSegment("00", displayColors["Green"], displayFontSizes["Large"])
	colonRight := newTimeSegment(":", displayColors["White"], displayFontSizes["Large"])
	second := newTimeSegment("00", displayColors["Green"], displayFontSizes["Large"])

	background := canvas.NewRectangle(displayBackgroundColors["Dark"])
	background.SetMinSize(fyne.NewSize(640, 160))

	segments := container.NewHBox(hour, colonLeft, minute, colonRight, second)
	root := container.NewPadded(container.NewStack(background, container.NewCenter(segments)))
	return &displayText{
		hour:       hour,
		colonLeft:  colonLeft,
		minute:     minute,
		colonRight: colonRight,
		second:     second,
		background: background,
		root:       root,
	}
}

func newTimeSegment(value string, textColor color.NRGBA, size float32) *canvas.Text {
	text := canvas.NewText(value, textColor)
	text.Alignment = fyne.TextAlignCenter
	text.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	text.TextSize = size
	return text
}

func (d *displayText) Object() fyne.CanvasObject {
	return d.root
}

func (d *displayText) SetText(value string) {
	parts := strings.Split(value, ":")
	if len(parts) == 3 {
		d.hour.Text = parts[0]
		d.minute.Text = parts[1]
		d.second.Text = parts[2]
	} else {
		d.hour.Text = value
		d.minute.Text = ""
		d.second.Text = ""
	}
	d.hour.Refresh()
	d.minute.Refresh()
	d.second.Refresh()
}

func (d *displayText) SetColor(name string) {
	selected, ok := displayColors[name]
	if !ok {
		selected = displayColors["Green"]
	}
	d.hour.Color = selected
	d.minute.Color = selected
	d.second.Color = selected
	d.hour.Refresh()
	d.minute.Refresh()
	d.second.Refresh()
}

func (d *displayText) SetColonColor(name string) {
	selected, ok := displayColors[name]
	if !ok {
		selected = displayColors["White"]
	}
	d.colonLeft.Color = selected
	d.colonRight.Color = selected
	d.colonLeft.Refresh()
	d.colonRight.Refresh()
}

func (d *displayText) SetFontSize(name string) {
	size, ok := displayFontSizes[name]
	if !ok {
		size = displayFontSizes["Large"]
	}
	d.hour.TextSize = size
	d.colonLeft.TextSize = size
	d.minute.TextSize = size
	d.colonRight.TextSize = size
	d.second.TextSize = size
	d.hour.Refresh()
	d.colonLeft.Refresh()
	d.minute.Refresh()
	d.colonRight.Refresh()
	d.second.Refresh()
}

func (d *displayText) SetTheme(name string) {
	backgroundColor, ok := displayBackgroundColors[name]
	if !ok {
		backgroundColor = displayBackgroundColors["Dark"]
	}
	d.background.FillColor = backgroundColor
	d.background.SetMinSize(fyne.NewSize(640, 160))
	d.background.Refresh()
}

func pickAllowed(value string, options []string, fallback string) string {
	for _, option := range options {
		if value == option {
			return value
		}
	}
	return fallback
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	totalSeconds := int(duration / time.Second)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func parseTimerInput(value string) time.Duration {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 3 {
		return 5 * time.Minute
	}
	hours, hourError := strconv.Atoi(parts[0])
	minutes, minuteError := strconv.Atoi(parts[1])
	seconds, secondError := strconv.Atoi(parts[2])
	if hourError != nil || minuteError != nil || secondError != nil || hours < 0 || minutes < 0 || minutes > 59 || seconds < 0 || seconds > 59 {
		return 5 * time.Minute
	}
	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
}
