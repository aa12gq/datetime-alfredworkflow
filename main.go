package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
)

func main() {
	workflow = aw.New()
	workflow.Run(run)
}

var (
	workflow *aw.Workflow

	icon = &aw.Icon{
		Value: aw.IconClock.Value,
		Type:  aw.IconClock.Type,
	}

	layouts = []string{
		"2006-01-02 15:04:05.999 MST",
		"2006-01-02 15:04:05.999 -0700",
		time.RFC3339,
		time.RFC3339Nano,
		time.UnixDate,
		time.RubyDate,
		time.RFC1123Z,
	}

	moreLayouts = []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999",
	}

	regexpTimestamp = regexp.MustCompile(`^[1-9]{1}\d+$`)
)

func run() {
	var err error

	args := workflow.Args()

	if len(args) == 0 {
		return
	}

	defer func() {
		if err == nil {
			workflow.SendFeedback()
			return
		}
	}()

	input := strings.Join(args, " ")

	if input == "now" {
		processNow()
		return
	}

	if regexpTimestamp.MatchString(input) {
		v, e := strconv.ParseInt(args[0], 10, 32)
		if e == nil {
			processTimestamp(time.Unix(v, 0))
			return
		}
		err = e
		return
	}

	err = processTimeStr(input)
}

func processNow() {
	now := time.Now()
	secs := fmt.Sprintf("%d", now.Unix())
	workflow.NewItem(secs).
		Subtitle("unix timestamp").
		Icon(icon).
		Arg(secs).
		Valid(true)

	processTimestamp(now)
}

func processTimestamp(timestamp time.Time) {
	unixTimestamp := timestamp.Unix()
	workflow.NewItem(strconv.FormatInt(unixTimestamp, 10)).
		Subtitle("Unix timestamp").
		Icon(icon).
		Arg(strconv.FormatInt(unixTimestamp, 10)).
		Valid(true)

	for _, layout := range layouts {
		formattedTime := timestamp.Format(layout)
		workflow.NewItem(formattedTime).
			Subtitle(layout).
			Icon(icon).
			Arg(formattedTime).
			Valid(true)
	}

	workflow.SendFeedback()
}

func processTimeStr(timestr string) error {
	if timestamp, ok := processSpecialDateInput(timestr); ok {
		timeVal := time.Unix(timestamp, 0)
		processTimestamp(timeVal)
		return nil
	}

	timestamp := time.Time{}
	layoutMatch := ""

	layoutMatch, timestamp, ok := matchedLayout(layouts, timestr)
	if !ok {
		layoutMatch, timestamp, ok = matchedLayout(moreLayouts, timestr)
		if !ok {
			return errors.New("no matched time layout found")
		}
	}

	secs := fmt.Sprintf("%d", timestamp.Unix())
	workflow.NewItem(secs).
		Subtitle("unix timestamp").
		Icon(icon).
		Arg(secs).
		Valid(true)

	for _, layout := range layouts {
		if layout == layoutMatch {
			continue
		}
		v := timestamp.Format(layout)
		workflow.NewItem(v).
			Subtitle(layout).
			Icon(icon).
			Arg(v).
			Valid(true)
	}

	return nil
}

func processSpecialDateInput(input string) (int64, bool) {
	now := time.Now()

	chineseMonths := map[string]time.Month{
		"一月": time.January, "二月": time.February, "三月": time.March,
		"四月": time.April, "五月": time.May, "六月": time.June,
		"七月": time.July, "八月": time.August, "九月": time.September,
		"十月": time.October, "十一月": time.November, "十二月": time.December,
	}

	var targetTime time.Time
	switch input {
	case "上个月":
		targetTime = now.AddDate(0, -1, 0)
	case "昨天":
		targetTime = now.AddDate(0, 0, -1)
	case "上周":
		targetTime = now.AddDate(0, 0, -7)
	case "明年":
		targetTime = now.AddDate(1, 0, 0)
	case "去年":
		targetTime = now.AddDate(-1, 0, 0)
	case "下个月":
		targetTime = now.AddDate(0, 1, 0)
	case "下周":
		targetTime = now.AddDate(0, 0, 7)
	case "今天", "now":
		targetTime = now
	default:
		if month, ok := chineseMonths[input]; ok {
			targetTime = time.Date(now.Year(), month, 1, 0, 0, 0, 0, now.Location())
		} else if t, err := time.Parse("January", input); err == nil {
			targetTime = time.Date(now.Year(), t.Month(), 1, 0, 0, 0, 0, now.Location())
		} else {
			return 0, false
		}
	}
	return targetTime.Unix(), true
}

func matchedLayout(layouts []string, timestr string) (matched string, timestamp time.Time, ok bool) {
	for _, layout := range layouts {
		v, err := time.Parse(layout, timestr)
		if err == nil {
			return layout, v, true
		}
	}
	return
}
