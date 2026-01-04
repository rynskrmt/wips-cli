package env

import (
	"os"
	"os/user"
	"runtime"

	"github.com/rynskrmt/wips-cli/internal/model"
)

// GetInfo returns the current environment information.
func GetInfo() model.EnvInfo {
	hostname, _ := os.Hostname()

	username := "unknown"
	u, err := user.Current()
	if err == nil {
		username = u.Username
	}

	return model.EnvInfo{
		Host: hostname,
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		User: username,
	}
}
