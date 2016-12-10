package main

import "github.com/mylxsw/task-runner/config"

func welcomeMessage(runtime *config.Runtime) string {

	if !runtime.Config.ColorfulTTY {
		return "TaskRunner Started."
	}

	return `
 _____         _    ____
|_   _|_ _ ___| | _|  _ \ _   _ _ __  _ __   ___ _ __
  | |/ _| / __| |/ / |_) | | | | '_ \| '_ \ / _ \ '__|
  | | (_| \__ \   <|  _ <| |_| | | | | | | |  __/ |
  |_|\__,_|___/_|\_\_| \_\\__,_|_| |_|_| |_|\___|_|

`
}
