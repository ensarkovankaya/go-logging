package core

import (
	"fmt"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota + 1
	LevelInfo
	LevelWarning
	LevelError
	LevelDisabled
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelDisabled:
		return "DISABLED"
	default:
		return ""
	}
}

func ParseLevel(levelStr string) (Level, error) {
	serialized := strings.ToUpper(strings.ReplaceAll(levelStr, " ", ""))
	switch serialized {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARNING", "WARN":
		return LevelWarning, nil
	case "ERROR":
		return LevelError, nil
	case "DISABLED", "DISABLE", "OFF":
		return LevelDisabled, nil
	default:
		return -1, fmt.Errorf("unknown log level: %s", levelStr)
	}
}
