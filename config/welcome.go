package config

var WelcomeMessageStr = `
   ::::::::::::::::::::
      :+:    :+:    :+          Coyotes v1.0
     +:+    +:+    +:+                                    █
    +#+    +#++:++#:           Powered by mylxsw          █
   +#+    +#+    +#+     github.com/mylxsw/coyotes        █
  #+#    #+#    #+#                                       █
 ###    ###    ###     ████████████████████████████████████

`

// WelcomeMessage function print welcome message
func WelcomeMessage() string {

	if !runtime.Config.ColorfulTTY {
		return "Coyotes Started."
	}

	return WelcomeMessageStr
}
