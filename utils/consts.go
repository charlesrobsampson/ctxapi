package utils

import (
	"os"
	"time"
)

var (
	Duration = struct {
		Second time.Duration
		Minute time.Duration
		Hour   time.Duration
		Day    time.Duration
		Week   time.Duration
		Month  time.Duration
		Year   time.Duration
	}{
		Second: time.Second,
		Minute: time.Minute,
		Hour:   time.Hour,
		Day:    time.Duration(24) * time.Hour,
		Week:   time.Duration(24) * time.Hour * 7,
		Month:  time.Duration(24) * time.Hour * 30,
		Year:   time.Duration(24) * time.Hour * 365,
	}

	Unit = map[string]time.Duration{
		"s": Duration.Second,
		"m": Duration.Minute,
		"h": Duration.Hour,
		"d": Duration.Day,
		"w": Duration.Week,
		"M": Duration.Month,
		"y": Duration.Year,
	}

	SkDateFormat = "2006-01-02T15:04:05Z"

	MainTableName  = os.Getenv("mainTableName")
	QueueTableName = os.Getenv("queueTableName")
)
