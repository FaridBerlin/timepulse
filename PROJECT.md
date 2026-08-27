# Timepulse — Project Documentation

> Complete technical and functional overview of the Timepulse application.

**Version:** 1.2
**Module:** `github.com/FaridBerlin/timepulse`
**Author / License:** MIT — Copyright (c) 2024 Night Cat
**Language:** Go (built and tested with Go 1.26.2, declared `go 1.21.0`)

---

## 1. What is Timepulse?

Timepulse is a Go application that provides **three time tools** — a **clock**, a
**stopwatch**, and a **timer** — in **two different interfaces**:

1. **Terminal / TUI** — a full-screen, block-digit style live display rendered
   directly in the terminal (using `termbox-go`).
2. **Desktop GUI** — a graphical window built with **Fyne v2**, featuring three
   separate tabs (Clock, Stopwatch, Timer) plus configurable appearance settings.

Both interfaces expose the same three tools but are operated differently:

| Tool       | Terminal command         | GUI tab     |
| ---------- | ------------------------ | ----------- |
| Clock      | `timepulse clock` (`c`)  | **Clock**   |
| Stopwatch  | `timepulse stopwatch` (`s`) | **Stopwatch** |
| Timer      | `timepulse timer` (`t`)  | **Timer**   |

The command-line layer is powered by the `urfave/cli/v2` library, which provides
sub-commands, one-letter aliases (`c`, `s`, `t`), `-h/--help`, and `-v/--version`.

---

## 2. Project (directory) structure

```
timepulse/
├── main.go              # CLI entry point: command + flag definitions, wires everything
├── go.mod               # module definition + dependencies
├── go.sum               # dependency checksums
├── README.md            # user guide (install, commands, settings, screenshots)
├── PROJECT.md           # this document (technical overview)
├── LICENSE              # MIT License
├── timepulse.desktop    # Linux .desktop launcher → "timepulse gui"
├── .gitignore           # ignores .vscode/ and docs/
├── img/                 # screenshots for the terminal examples in the README
├── desktop-img/         # screenshots + animated GIF of the desktop GUI
├── docs/                # additional screenshots (committed but git-ignored)
├── util/
│   ├── clock.go         # terminal clock implementation
│   ├── stopwatch.go     # terminal stopwatch implementation
│   ├── timer.go         # terminal timer implementation
│   ├── util.go          # shared helpers (formatting, colors, time parsing)
│   └── SmallNumber.go   # block-digit "font" rendering + dispatch
└── gui/
    ├── app.go           # desktop GUI (Fyne): all tabs + appearance controls
    └── app_test.go      # unit tests for timer input parsing
```

---

## 3. What is installed / dependencies

`go.mod` declares the module and pulls in the following libraries:

### Direct dependencies
| Module | Version | Purpose |
| ------ | ------- | ------- |
| `github.com/urfave/cli/v2` | v2.27.1 | Command-line framework (commands, flags, help) |
| `github.com/nsf/termbox-go` | v1.1.1 | Low-level terminal UI / raw screen rendering |
| `fyne.io/fyne/v2` | v2.6.3 | Cross-platform desktop GUI toolkit |

### Indirect dependencies (automatically resolved helpers)
- `fyne.io/systray` — tray integration support for Fyne.
- Fyne graphics stack: `go-gl/gl`, `go-gl/glfw/v3.3/glfw`, `fyne-io/gl-js`,
  `fyne-io/glfw-js`, `go-text/render`, `go-text/typesetting`, `srwiley/oksvg`,
  `srwiley/rasterx`, `fyne-io/oksvg`, `fyne-io/image`.
- `godbus/dbus/v5`, `rymdport/portal` — Linux/desktop integration.
- `jeandeaual/go-locale`, `nicksnyder/go-i18n/v2`, `yuin/goldmark` — localization
  and markup helpers bundled with Fyne.
- Various other transitive deps (`BurntSushi/toml`, `gobmp`, `nfnt/resize`,
  `go-runewidth`, `rivo/uniseg`, `xrash/smetrics`, `cpuguy83/go-md2man`,
  `russross/blackfriday`, `davecgh/go-spew`, `stretchr/testify`).

### System packages (Linux, required to build the Fyne desktop version)
Building the desktop app on Linux needs these dev packages:

```bash
sudo apt install pkg-config libgl1-mesa-dev libx11-dev \
   libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
```

### Go toolchain
- Declared in `go.mod`: **Go 1.21.0**
- Confirmed working in this environment: **Go 1.26.2**

---

## 4. How the project is built and run

### Install the binary (release / manual)
Download the compiled `timepulse` file and place it in your `$PATH`:

```bash
sudo mv ./timepulse /usr/local/bin
```

or build and install it yourself:

```bash
go build -o timepulse .
sudo install -m 755 timepulse /usr/local/bin/timepulse
```

### Run from source
```bash
go run . <command> [options]
```

### Desktop launcher for the Linux application menu
```bash
go build -o timepulse .
sudo install -m 755 timepulse /usr/local/bin/timepulse
mkdir -p ~/.local/share/applications
cp timepulse.desktop ~/.local/share/applications/timepulse.desktop
```

`timepulse.desktop` launches the GUI:

```ini
[Desktop Entry]
Name=Timepulse
Comment=Clock, stopwatch, and timer
Exec=/usr/local/bin/timepulse gui
Terminal=false
Type=Application
Categories=Utility;Clock;
```

### Verify the build / tests
```bash
go build ./...   # builds main, util, gui
go vet ./...     # static analysis
go test ./...    # runs the unit tests (gui package passes)
```

---

## 5. Architecture overview

```
                       main.go (urfave/cli/v2)
        ┌───────────────┬───────────────┬───────────────┐
        │               │               │               │
     clock            stopwatch       timer        gui.Run()
        │               │               │               │
        └───────────────┴───────────────┴───────────────┘
                     util package (termbox-go)
        util.go  clock.go  stopwatch.go  timer.go  SmallNumber.go
                          │
                 drawString / Cell rendering
                          │
                     terminal screen

gui package → Fyne desktop app:
   newClockTab / newStopwatchTab / newTimerTab
   newHeaderControls / newAppearanceControls / newDisplayText
   appAppearance (persisted preferences)
```

**Entry point (`main.go`):**
- Defines a `cli.App` named `timepulse`, version `"1.2"`, usage *"A terminal ttl
  clock timer and stopwatch build by golang"*.
- A global `--color, -c` flag exists at the top level.
- Registers four commands:
  - `gui` → `gui.Run()`
  - `stopwatch`/`s` → `util.Stopwatch`
  - `timer`/`t` → `util.Timer`
  - `clock`/`c` → `util.Clock`
- `app.Run(os.Args)` is invoked; fatal errors are logged via `log.Fatal`.

---

## 6. The terminal interface

The terminal tools use **termbox-go** for alternative-screen, raw terminal
rendering. Common flow for each tool:

1. `termbox.Init()` initializes raw terminal mode and returns an error if it fails.
2. `termbox.Close()` is deferred to restore the terminal on exit.
3. `termbox.SetOutputMode(termbox.Output256)` enables 256-color output.
4. Colors are resolved via `FlagColor(...)` (color names or 1–256 numeric codes).
5. A **rendering goroutine** loops forever:
   - Reads terminal dimensions via `termbox.Size()`.
   - Clears the screen each frame.
   - Computes the current time.
   - **Renders each character as a block-size digit** (see §7).
   - `termbox.Flush()` pushes the frame to the screen.
   - `time.Sleep(...)` paces the refresh (clock: 1 s, stopwatch/timer: ~ms).
6. The **main goroutine blocks on `termbox.PollEvent()`** to handle input:
   - Pressing `Esc`, `q`, `Q`, or `Ctrl+C` exits the tool (`break mainLoop`).
   - `EventError` → `os.Exit(1)`.

### Why a separate rendering goroutine?
The render loop and the input-polling loop run concurrently: the clock keeps
refreshing while the program waits (in the main goroutine) for an exit key. The
deferred `termbox.Close()` restores the terminal when the loop breaks.
---

## 6.1 Clock — terminal

Defined in `util/clock.go`, registered as `timepulse clock` (alias `c`).

**Behaviour:**
- Shows the current wall-clock time with a live, full-screen block-clock display.
- Default format is `HH:MM:SS` (24-hour).
- Refreshes once per second.

**Options (as defined in `main.go`):**

| Flag | Alias | Default | Description |
| ---- | ----- | ------- | ----------- |
| `--color` | `-c` | green | Digit color. |
| `--second` | `-s, --sec` | `"true"` | Show (`true`) or hide (`false`) seconds. |
| `--date` | `-d` | `"false"` | Show (`true`) or hide (`false`) the date. |
| `--dateformate` | `--df` | `"2006/02/01"` | Go reference-time date layout. |
| `--colon-color` | `--cc` | follows `--color` | Color of the `:` separators. |
| `--hour-format` | `--hf` | 24-hour | Set to `12` for a 12-hour display. |

**How it works:**
- `formatTime()` builds `%02d:%02d:%02d`; in 12-hour mode hours are `Hour() % 12`
  and `0` is replaced with `12`.
- When `--second false`, only HH:MM (5 chars) is drawn; when `--second true`,
  HH:MM:SS (8 chars). The `ClockDiff` / `ClockWithSecDiff` arrays space the
  block columns.
- In each digit cell, `i == 2 || i == 5` are the colon positions and use the colon
  color; every other cell uses the main color.
- If `--date true`, a date string formatted with `--dateformate` is centered below
  the time via `drawString`.

> ⚠️ **Note (possible bug):** `main.go` registers an `--hour-format/--hf` flag but
> `clock.go` checks `cCtx.String("color") == "hour-format"` instead of reading the
> actual flag, so `--hour-format 12` is **not wired up correctly** in the terminal.

---

## 6.2 Stopwatch — terminal

Defined in `util/stopwatch.go`, invoked as `timepulse stopwatch` (alias `s`).

**Behaviour:**
- Starts counting **up** the moment the user runs the binary.
- T = `time.Since(start)` rounded to the millisecond.
- Refreshes every ~1 ms.
- Default format: `HH:MM:SS.mmm`. With `--disable-hour true`: `MM:SS.mmm`.

**Options:**

| Flag | Alias | Default | Description |
| ----- | ------ | ------- | ----------- |
| `--color` | `-c` | green | Digit color. |
| `--colon-color` | `--cc` | follows `--color` | Color of `:` separators. |
| `--disable-hour` | `--dh` | `false` | Hide hours, show only `MM:SS`. |

**How it works:**
- `StopwatchFormatTime()` decomposes the duration into h/m/s/ms and formats it
  `%02d:%02d:%02d.%03d`.
- `StopwatchFormatTimeWithoutHour()` omits the hours.
- Column spacing uses `SmallStopwatchDiff` or `SmallStopwatchWithoutHourDiff`.
- No interactive pause/reset in the terminal; exit via `Esc`/`q`/`Q`/`Ctrl+C`.

---

## 6.3 Timer — terminal

Defined in `util/timer.go`, invoked as `timepulse timer` (alias `t`).

**Behaviour:**
- Counts **down** from a user-provided duration to zero.
- Timeline: `end := time.Now().Add(total)`; each frame `current := time.Until(end)`.
- When `current <= 0`, the display freezes at `00:00:00.000` (or `00:00.000` with
  hour-disabled).
- Default duration if nothing is given: **5 minutes** (`300` seconds).
- Refreshes every ~1 ms.

**Options:**

| Flag | Alias | Default | Description |
| ---- | ------ | ------- | ----------- |
| `--color` | `-c` | green | Digit color. |
| `--hour` | `--hr` | 0 | Hours to count down. |
| `--minute` | `-m, --min` | 0 | Minutes to count down. |
| `--second` | `-s, --sec` | 0 | Seconds to count down. |
| `--time` | `-t` | — | Duration as `HH:MM:SS` (parsed by `timeStringToSeconds`). |
| `--disable-hour` | `--dh` | `false` | Show only `MM:SS`. |
| `--disable-millisecond` | `--dm` | `false` | Hide the `.mmm` part. |
| `--colon-color` | `--cc` | follows `--color` | Color of `:` separators. |

Two equivalent ways to specify a duration:

```bash
# classic — individual units
timepulse timer -hr 1 -m 1 -s 1

# lazy — whole string "HH:MM:SS"
timepulse timer -t 1:20:01
```

**How it works:**
- Parses `hour`, `minute`, `second`, adds seconds from `--time` via
  `timeStringToSeconds()`, then `total` receives `hour*3600 + min*60 + sec*60`.
- If `total == 0`, falls back to `300` (5 minutes).
- Chooses one of four column layouts depending on the flags, then renders the
  leftover time every few ms.

> ⚠️ **Note (possible bug):** in the source the seconds component is computed as
> `sec*60` and combined seconds are added on top of the individual units, so a
> literal `-s 1` counts for ~60 seconds rather than 1. The intended behaviour is
> the standard 1-second-per-second countdown described above.

---

## 6.4 Shared terminal helpers (`util/util.go`)

- `StopwatchFormatTime(d, disableMillisecond)` → `HH:MM:SS[.mmm]`.
- `StopwatchFormatTimeWithoutHour(d, disableMillisecond)` → `MM:SS[.mmm]`.
- `formatTime(t, use12HourFormat)` → clock string (optional 12-hour).
- `drawString(x, y, text, fg, bg)` → paints a string to the terminal.
- `FlagColor(nameOrCode)` → maps color names to `termbox.Attribute`; accepts a
  numeric code 1–256; falls back to green. Recognized names: `black, blue, cyan,
  dark-gray, green, light-green, light-blue, light-cyan, light-gray, light-magenta,
  light-red, light-yellow, magenta, red, white, yellow`.
- `timeStringToSeconds("HH:MM:SS")` → total seconds (or error for invalid length).

---

## 7. The block-digit "font" (`util/SmallNumber.go`)

The terminal does not print plain ASCII time. Instead each digit is drawn as a
**block made of `"  "` (two-space) terminal cells** in the chosen color, so the
characters look like a large seven-segment clock. `drawString` is reused with a
color as both foreground and background to produce filled rectangles.

Rendering functions:
- `SmallZero` … `SmallNine` — draw each digit 0–9.
- `SmallColon` — draws the `:` separator (two short vertical stacks).
- `SmallDot` — draws the `.` used as the millisecond delimiter.
- `CaseNumber(termWidth, termHeight, color, char, diff)` — **dispatcher** that
  switches on the byte (`'0'`–`'9'`, `':'`, `'.'`) and calls the matching font
  function at the given horizontal `diff` offset.

Because every "cell" is a 2-column-wide filled block, the columns are spaced with
the per-command Diff arrays (`ClockDiff`, `ClockWithSecDiff`, `SmallStopwatchDiff`,
`SmallTimerDiff`, …) so the digits sit cleanly side by side and the clock looks
modern in the terminal.

---

## 8. The desktop interface (GUI)

Located in `gui/app.go`, launched by the `gui` command (or `timepulse.desktop`).

### 8.1 Window & overall structure
- Creates a Fyne app with ID `github.com.FaridBerlin.timepulse` and a window titled
  **"Timepulse"**.
- Restores the previous window size from Fyne Preferences (fallback 720×560) and
  saves it again on close (`SetCloseIntercept`).
- Uses `container.NewAppTabs` with three tabs — **Clock**, **Stopwatch**, **Timer**.
- `container.NewBorder` places a header bar (hamburger menu `☰`) on top and the
  tabs as the main body.

### 8.2 Clock tab (`newClockTab`)
- A single display that updates every second.
- A goroutine with a 1-second `time.Ticker` reads the wall clock and renders
  `HH:MM:SS` (24-hour) via `current.Format("15:04:05")`.
- A `stop` channel lets the goroutine exit cleanly.

### 8.3 Stopwatch tab (`newStopwatchTab`)
- `Start/Pause` and `Reset` buttons.
- Tracks `startedAt` and accumulated `elapsed`, guarded by a `sync.Mutex`.
- Tapping **Start** sets `startedAt = time.Now()` and shows `Pause`; tapping
  **Pause** adds `time.Since(startedAt)` to `elapsed` and shows `Start`.
- **Reset** zeroes `elapsed` (and restarts the base if currently running).
- A 50 ms ticker continuously recomputes `display.SetText(formatDuration(...))`.
- Display is `HH:MM:SS` — **no milliseconds** in the GUI stopwatch.

### 8.4 Timer tab (`newTimerTab`)
- A text entry accepts `HH:MM:SS`, placeholder
  `"HH:MM:SS (default 00:05:00)"`.
- **Start/Pause** and **Reset** buttons.
- On start: `remaining = parseTimerInput(input.Text)`, then
  `endAt = time.Now().Add(remaining)`.
- A 100 ms ticker refreshes the remaining time; when it hits `<= 0`, the timer
  auto-stops and the button returns to "Start".
- **Reset** restores the 5-minute default.

### 8.5 Appearance & preferences (`appAppearance`)
All settings persist via Fyne `Preferences` and apply through the shared
`displayText` abstraction, so a change affects every tab at once.

| Setting | Preferences key | Options | Default |
| -------- | ---------------- | ------- | ------- |
| Theme | `appearance.theme` | Dark, Light | Dark |
| Digit color | `appearance.color` | Green, Red, Yellow, Blue, White | Green |
| Colon color | `appearance.colonColor` | same list | White |
| Font size | `appearance.size` | Small, Medium, Large, X-Large, XX-Large | Large |
| Window width | `window.width` | int | 720 |
| Window height | `window.height` | int | 560 |

Changes are broadcast to all registered displays via `RegisterDisplay`, persisted,
and can be reset with the **Reset** button in the appearance popup.

### 8.6 The appearance popup
The hamburger button `☰` opens a `widget.NewPopUp` containing **Theme**, **Color**,
**Colon**, **Size** selects plus a **Reset** button, spaced with separators. It
anchors just below the button and dismisses when clicking outside.

### 8.7 The numeric display widget (`newDisplayText`)
- Horizontal layout of `hour` + `colonLeft` + `minute` + `colonRight` + `second`
  bold, monospace `canvas.Text` segments, centered on a background rectangle.
- `SetText("HH:MM:SS")` splits on `:` and updates the three numeric segments.
- `SetColor` / `SetColonColor` / `SetFontSize` / `SetTheme` update appearance and
  call `Refresh()`.
- Font sizes (points): Small=44, Medium=64, Large=90, X-Large=120, XX-Large=200.

### 8.8 Helper formatting in the GUI
- `formatDuration()` → `HH:MM:SS` (clamps negatives to 0).
- `parseTimerInput()` → parses `HH:MM:SS` into a `time.Duration`, falling back to 5
  minutes on invalid input, minutes > 59, or negative numbers. Covered by
  `gui/app_test.go` (`TestParseTimerInput`).
---

## 9. Known quirks / possible bugs

1. **Timer terminal seconds scaling** (`util/timer.go`): the combined total is
   computed with `sec*60` and seconds added on top of the individual units, so a
   countdown intended for a few `--second` values lasts much longer than expected.
2. **12-hour clock flag** (`util/clock.go`): the code checks
   `cCtx.String("color") == "hour-format"` instead of the registered
   `--hour-format` flag, so 12-hour mode is likely not activated by `--hour-format 12`.
3. `FlagColor` silently falls back to green on a misspelled color name or an
   out-of-range numeric code (no user-facing error).
4. The terminal and GUI back-ends keep separate formatting/parsing logic and are
   not fully mirrored (e.g., the GUI does not display milliseconds, the terminal
   block font and GUI canvas-text sizes differ). This can drift over time.

---

## 10. Running / using — quick cheat-sheet

```bash
# Global help / version
timepulse --help
timepulse --version

# Desktop GUI
timepulse gui

# Clock: seconds + red + date
timepulse clock -s true -c red -d true

# Stopwatch: hide hours, custom colon color
timepulse stopwatch -dh true -cc yellow

# Timer: 1h 1m 1s (classic) or 1:20:01 (lazy)
timepulse timer -hr 1 -m 1 -s 1
timepulse timer -t 1:20:01

# One-letter aliases
timepulse c  # clock
timepulse s  # stopwatch
timepulse t  # timer
```

Shorten to `t c`, `t s`, `t t` in Bash/Zsh:

```bash
alias t=timepulse
echo 'alias t=timepulse' >> ~/.bashrc
source ~/.bashrc
```

Exit any terminal tool by pressing **Esc**, **q**, **Q**, or **Ctrl+C**.