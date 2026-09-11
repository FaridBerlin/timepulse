package util

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/urfave/cli/v2"
	"os"
	"strconv"
	"time"
)

var SmallTimerDiff = [11]int{7, 6, 6, 7, 6, 6, 7, 6, 4, 7, 7}
var SmallTimerWithoutHourDiff = [11]int{7, 6, 6, 7, 6, 4, 7, 7}

var SmallTimerWithoutHourAndMillisecondDiff = [11]int{7, 6, 6, 7, 6}
var SmallTimerWithoutMillisecondDiff = [11]int{7, 6, 6, 7, 6, 6, 7, 6}

// timerTotalSeconds resolves the countdown duration in seconds from the timer
// flags. When --time is set it takes priority and is parsed on its own
// (rather than being added on top of --hour/--minute/--second); otherwise the
// individual unit flags are combined. A duration that resolves to zero falls
// back to the default 5-minute countdown.
func timerTotalSeconds(hourStr, minuteStr, secondStr, timeStr string) (int, error) {
	var total int
	if timeStr != "" {
		seconds, err := timeStringToSeconds(timeStr)
		if err != nil {
			return 0, err
		}
		total = seconds
	} else {
		hour, _ := strconv.Atoi(hourStr)
		minute, _ := strconv.Atoi(minuteStr)
		second, _ := strconv.Atoi(secondStr)
		total = hour*3600 + minute*60 + second
	}
	if total == 0 {
		total = 300
	}
	return total, nil
}

func Timer(cCtx *cli.Context) error {
	err := termbox.Init()
	if err != nil {
		fmt.Println("failed to initialize termbox:", err)
		return err
	}
	defer termbox.Close()
	termbox.SetOutputMode(termbox.Output256)
	color := FlagColor(cCtx.String("color"))

	total, err := timerTotalSeconds(cCtx.String("hour"), cCtx.String("minute"), cCtx.String("second"), cCtx.String("time"))
	if err != nil {
		fmt.Println("invalid duration:", err)
		return err
	}
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	colonColor := FlagColor(cCtx.String("colon-color"))

	if cCtx.String("colon-color") == "" && cCtx.String("color") != "" {
		colonColor = FlagColor(cCtx.String("color"))
	}
	go func() {
		duration := time.Duration(total) * time.Second
		end := time.Now().Add(duration)
		for {
			termWidth, termHeight := termbox.Size()
			current := time.Until(end)
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			NowTime := ""

			disableMillisecond := cCtx.String("disable-millisecond") == "true"
			if cCtx.String("disable-hour") == "true" {

				if current <= 0 {
					NowTime = "00:00.000"
				} else {
					NowTime = StopwatchFormatTimeWihtoutHour(current, disableMillisecond)
				}
			} else {
				if current <= 0 {
					NowTime = "00:00:00.000"
				} else {
					NowTime = StopwatchFormatTime(current, disableMillisecond)
				}
			}
			diff := -38
			totalString := 12
			if cCtx.String("disable-hour") == "true" {
				if disableMillisecond {
					totalString = 5
					diff = -16
				} else {
					totalString = 9
					diff = -29
				}
			} else if disableMillisecond {
				totalString = 8
				diff = -25
			}

			for i := 0; i < totalString; i++ {
				if i != 0 {
					if cCtx.String("disable-hour") == "true" {
						diff = diff + SmallTimerWithoutHourDiff[i-1]
					} else {
						diff = diff + SmallTimerDiff[i-1]
					}
				}
				if i == 2 || i == 5 {
					CaseNumber(termWidth, termHeight, colonColor, NowTime[i], diff)
				} else {
					CaseNumber(termWidth, termHeight, color, NowTime[i], diff)
				}
			}
			termbox.Flush()

			time.Sleep(10 * time.Millisecond)
		}
	}()

mainLoop:
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch {
			case ev.Key == termbox.KeyEsc || ev.Ch == 'q' || ev.Ch == 'Q':
				break mainLoop
			case ev.Key == termbox.KeyCtrlC:
				break mainLoop
			}
		case termbox.EventError:
			os.Exit(1)
		}
	}
	return nil
}
