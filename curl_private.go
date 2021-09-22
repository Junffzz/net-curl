/*
@Time : 2021/1/21 15:10
@Author : ZhaoJunfeng
@File : curl_private
*/
package curl

import (
    "context"
    "errors"
    "fmt"
    "runtime/debug"
    "sync"
)

func (c *Curl) init() {
    c.RequestListCtx = nil
    c.engine = nil
    c.Options = nil
}

func (c *Curl) handleContext(ctx context.Context, rtx *Context, wg *sync.WaitGroup) func() {
    // 合并Context的处理链
    c.combineHandlersToContext(rtx, func(rCtx *Context) {
        if rCtx.IsAborted() {
            return
        }
        HttpClientDo(ctx, rCtx)
    })

    return func() {
        defer func() {
            if err := recover(); err != nil {
                debugPrint("httpClientDoFunc panic: %v; Stack: %v", err, string(debug.Stack()))
                rtx.Error = errors.New(fmt.Sprintf("curl httpClientDoFunc panic. %v", err))
            }
        }()
        defer wg.Done()

        rtx.Next()
    }
}

/**
 * 合并处理链至Context
 * @date: 2021/3/4
 */
func (c *Curl) combineHandlersToContext(rtx *Context, handlers ...HandlerFunc) {
    finalSize := len(c.Handlers) + len(rtx.handlers) + len(handlers)
    if finalSize >= int(abortIndex) {
        panic("too many handlers")
    }
    mergedHandlers := make(HandlersChain, finalSize)
    copy(mergedHandlers, c.Handlers)                         // 全局middleware
    copy(mergedHandlers[len(c.Handlers):], rtx.handlers)     // 请求middleware
    copy(mergedHandlers[finalSize-len(handlers):], handlers) // 追加内核middleware
    rtx.handlers = mergedHandlers
    return
}

func (c *Curl) returnObj() ICurl {
    if c.root {
        return c.engine
    }
    return c
}
