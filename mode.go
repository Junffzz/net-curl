/*
@Time : 2021/3/4 18:41
@Author : ZhaoJunfeng
@File : mode
*/
package curl

import (
	"io"
	"os"
)

// EnvCurlMode indicates environment name for curl mode.
const EnvGinMode = "CURL_MODE"

const (
	// DebugMode indicates curl mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates curl mode is release.
	ReleaseMode = "release"
	// TestMode indicates curl mode is test.
	TestMode = "test"
)

const (
	debugCode = iota
	releaseCode
	testCode
)

// DefaultWriter is the default io.Writer used by Curl for debug output and
// middleware output like Logger() or Recovery().
// Note that both Logger and Recovery provides custom ways to configure their
// output io.Writer.
// To support coloring in Windows use:
// 		import "github.com/mattn/go-colorable"
// 		curl.DefaultWriter = colorable.NewColorableStdout()
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Curl to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

var curlMode = debugCode
var modeName = DebugMode

func init() {
	mode := os.Getenv(EnvGinMode)
	SetMode(mode)
}

// SetMode sets curl mode according to input string.
func SetMode(value string) {
	switch value {
	case DebugMode, "":
		curlMode = debugCode
	case ReleaseMode:
		curlMode = releaseCode
	case TestMode:
		curlMode = testCode
	default:
		panic("curl mode unknown: " + value)
	}
	if value == "" {
		value = DebugMode
	}
	modeName = value
}

// Mode returns currently curl mode.
func Mode() string {
	return modeName
}
