/*
@Time : 2021/3/4 18:38
@Author : ZhaoJunfeng
@File : debug
内置方法，可供全局配置
*/
package curl

import (
    "fmt"
    "strings"
)

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(gin.ReleaseMode) to disable debug mode.
func IsDebugging() bool {
    return curlMode == debugCode
}

func debugPrint(format string, values ...interface{}) {
    if IsDebugging() {
        if !strings.HasSuffix(format, "\n") {
            format += "\n"
        }
        fmt.Fprintf(DefaultWriter, "[CURL-debug] "+format, values...)
    }
}

func debugPrintWARNINGDefault() {
    debugPrint(`[WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.
`)
}

func debugPrintWARNINGNew() {
    debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	curl.SetMode(gin.ReleaseMode)
`)
}

func debugPrintWARNINGSetHTMLTemplate() {
    debugPrint(`[WARNING] Since SetHTMLTemplate() is NOT thread-safe. It should only be called
at initialization. ie. before any route is registered or the router is listening in a socket:
	router := curl.Default()
	router.SetHTMLTemplate(template) // << good place

`)
}

func debugPrintError(err error) {
    if err != nil {
        if IsDebugging() {
            fmt.Fprintf(DefaultErrorWriter, "[GIN-debug] [ERROR] %v\n", err)
        }
    }
}
