package version

import (
	"fmt"
	"runtime"
)

// GitCommit describes the git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version describes the main version number that is being run at the moment.
const Version = "0.2.0"

// BuildDate describes the datetime when was compiled.
var BuildDate = ""

var GoVersion = runtime.Version()

var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
