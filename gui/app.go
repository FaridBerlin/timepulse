package gui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func Run() error {
	application := app.NewWithID("github.com.FaridBerlin.timepulse")
	window := application.NewWindow("Timepulse")

	clock := newClockTab()
	stopwatch := newStopwatchTab()
	timer := newTimerTab()

	tabs := container.NewAppTabs(
		container.NewTabItem("Clock", clock),
		container.NewTabItem("Stopwatch", stopwatch),
		container.NewTabItem("Timer", timer),
	)
	window.SetContent(tabs)
	window.Resize(fyne.NewSize(420, 250))
	window.CenterOnScreen()
	window.ShowAndRun()
	return nil
}

func newClockTab() fyne.CanvasObject {
	label := newDisplayLabel()
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case current := <-ticker.C:
				label.SetText(current.Format("15:04:05"))
			case <-stop:
				return
			}
		}
	}()
	return container.NewCenter(label)
}

func newStopwatchTab() fyne.CanvasObject {
	label := newDisplayLabel()
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
		label.SetText(formatDuration(0))
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
			label.SetText(formatDuration(current))
		}
	}()
	return container.NewBorder(label, container.NewCenter(container.NewHBox(startButton, resetButton)), nil, nil)
}

func newTimerTab() fyne.CanvasObject {
	label := newDisplayLabel()
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
		label.SetText(formatDuration(remaining))
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
			label.SetText(formatDuration(current))
		}
	}()

	controls := container.NewVBox(input, container.NewHBox(startButton, resetButton))
	return container.NewBorder(label, controls, nil, nil)
}

func newDisplayLabel() *widget.Label {
	label := widget.NewLabel("00:00:00")
	label.Alignment = fyne.TextAlignCenter
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Resize(fyne.NewSize(380, 80))
	return label
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
