package console

import (
	"fmt"

	"github.com/mylxsw/task-runner/config"
)

const (
	TextBlack = iota + 30
	TextRed
	TextGreen
	TextYellow
	TextBlue
	TextMagenta
	TextCyan
	TextWhite
)

func ColorfulText(color int, text string) string {
	if config.GetRuntime().Config.ColorfulTTY {
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", color, text)
	}

	return text
}
