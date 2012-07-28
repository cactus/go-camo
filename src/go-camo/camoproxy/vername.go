package camoproxy

import "fmt"

const (
	ServerName    = "go-camo"
	ServerVersion = "0.0.3"
)

var ServerNameVer = fmt.Sprintf("%s %s", ServerName, ServerVersion)
