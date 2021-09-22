/*
@Time : 2021/3/4 18:38
@Author : ZhaoJunfeng
@File : defaut
*/
package curl

import (
    "context"
    "errors"
    "fmt"
    "git.100tal.com/wangxiao_go_lib/xesLogger/logtrace"
    "git.100tal.com/wangxiao_xesbiz_operation/gently-utils/logger"
    "github.com/panjf2000/ants/v2"
    "github.com/spf13/cast"
    "io"
    "runtime/debug"
    "strings"
)

// goroutine pool, the default size 5000 = 10M memory.
var (
    workerPoolSize = 5000
    workerPool     *ants.Pool
)

func init() {
    var err error
    workerPool, err = ants.NewPool(workerPoolSize, ants.WithPreAlloc(true))
    if err != nil {
        debugPrint("init create workerPool fail: %v", err)
    }
}

// SetWorkerPoolSize sets curl workerPool's capacity according to input int.
func SetWorkerPoolSize(size int) {
    if size <= 0 {
        debugPrint("size %v don't set workerPoolSize!", size)
        return
    }
    workerPool.Tune(size)
}

// Default returns an Engine instance with the Logger and Recovery middleware already attached.
func Default() *Engine {
    // debugPrintWARNINGDefault()
    engine := New()
    engine.Use(Logger(), Recovery())
    return engine
}

// Logger instances a Logger middleware that will write the logs to gin.DefaultWriter.
// By default gin.DefaultWriter = os.Stdout.
func Logger() HandlerFunc {
    return func(rtx *Context) {
        rtx.Next()

        var statLevel = ""
        if rtx.HttpCode > 0 && (rtx.HttpCode > 299 || rtx.HttpCode < 200) {
            statLevel = ".error"
        }

        ctx := rtx.TransferToGoCtx()
        logtraceMap := logtrace.GenLogTraceMetadata()

        logtraceMap.Set("x_trace_id", "\""+cast.ToString(ctx.Value("x_trace_id"))+"\"")

        ctx = context.WithValue(ctx, logtrace.GetMetadataKey(), logtraceMap)
        tracenode := logtrace.ExtractTraceNodeFromXesContext(ctx)
        tracenode.Set("ori_url", "\""+rtx.Url+"\"")
        // tracenode.Set("params", "\""+rtx.Values+"\"")
        tracenode.Set("method", "\""+rtx.Method+"\"")
        tracenode.Set("http_code", "\""+cast.ToString(rtx.HttpCode)+"\"")
        // tracenode.Set("total_time", "\""+strconv.FormatFloat(stat.totalTime.Seconds(), 'E', -1, 64)+"\"")
        // tracenode.Set("x_cost_ms", "\""+cast.ToString(stat.totalTime.Milliseconds())+"\"")
        tag := "http." + strings.ToLower(rtx.Method) + statLevel
        if err := rtx.Error; err != nil {
            logger.Ix(ctx, tag, "curl log, url: %v; err:%v", rtx.Url, err)
        } else {
            logger.Ix(ctx, tag, "curl log, url: %v", rtx.Url)
        }
    }
}

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() HandlerFunc {
    return RecoveryWithWriter(DefaultErrorWriter)
}

// RecoveryWithWriter returns a middleware for a given writer that recovers from any panics and writes a 500 if there was one.
func RecoveryWithWriter(out io.Writer) HandlerFunc {
    return func(c *Context) {
        defer func() {
            if err := recover(); err != nil {
                debugPrint("recover panic: %v; Stack: %v", err, string(debug.Stack()))
                logger.E("curl", "[recover] panic: %v; Stack: %v", err, string(debug.Stack()))
                c.Error = errors.New(fmt.Sprintf("curl recover panic. %v", err))
                c.Abort()
            }
        }()
        c.Next()
    }
}
