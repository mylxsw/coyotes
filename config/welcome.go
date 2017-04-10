package config

var WelcomeMessageStr = `
   ::::::::::::::::::::
      :+:    :+:    :+          TaskRunner v1.0
     +:+    +:+    +:+                                    █
    +#+    +#++:++#:           Powered by mylxsw          █
   +#+    +#+    +#+     github.com/mylxsw/task-runner    █
  #+#    #+#    #+#                                       █
 ###    ###    ###     ████████████████████████████████████

`

// WelcomeMessage function print welcome message
func WelcomeMessage() string {

	if !runtime.Config.ColorfulTTY {
		return "TaskRunner Started."
	}

	return WelcomeMessageStr
}
