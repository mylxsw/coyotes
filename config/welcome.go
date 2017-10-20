package config

// VERSION 当前版本
const VERSION = "v1.0"

// WelcomeMessageStr is a message to show
var WelcomeMessageStr = `
  ____                  _
 / ___|___  _   _  ___ | |_ ___  ___
| |   / _ \| | | |/ _ \| __/ _ \/ __|
| |__| (_) | |_| | (_) | ||  __/\__ \
 \____\___/ \__, |\___/ \__\___||___/ ` + VERSION + `
            |___/

        ❉ Powered by mylxsw ❉
      github.com/mylxsw/coyotes

`

// WelcomeMessage function print welcome message
func WelcomeMessage() string {

	if !runtime.Config.ColorfulTTY {
		return "coyotes " + VERSION + " started."
	}

	return WelcomeMessageStr
}
