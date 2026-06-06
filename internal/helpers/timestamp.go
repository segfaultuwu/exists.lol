package version

import (
	"strconv"
	"time"
)

func DiscordTimestampRFC3339(value string, style string) string {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}

	if style == "" {
		style = "F"
	}

	return "<t:" + strconv.FormatInt(t.Unix(), 10) + ":" + style + ">"
}
