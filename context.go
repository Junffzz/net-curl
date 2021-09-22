/*
@Time : 2021/3/3 15:26
@Author : ZhaoJunfeng
@File : context
*/
package curl

import (
    "context"
    "fmt"
    "git.100tal.com/wangxiao_xesbiz_operation/gently-utils/net/curl/httputil"
    "io"
    "math"
    "net/http"
    "sync"
    "time"
)

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*Context)

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

const abortIndex int8 = math.MaxInt8 / 2

// curl上下文
type Context struct {
    Request
    Response

    Index  int8    // 索引
    engine *Engine // 引擎
    Option Options // 来源curl的config拷贝

    handleIndex int8 // 处理链索引
    handlers    HandlersChain

    // This mutex protect Keys map
    mu sync.RWMutex

    // Keys is a key/value pair exclusively for the context of each request.
    Keys map[string]interface{}

    client *http.Client
}

type Request struct {
    Header     map[string]string
    Method     string
    Url        string                 // 包含参数
    Values     io.Reader              // body参数
}

// httpClient请求
func HttpClientDo(ctx context.Context, rtx *Context) {
    var err error
    var resp httputil.HttpResp
    // retry
    if rtx.Option.retryTimes == 0 {
        rtx.Option.retryTimes = 1
    }
    req := rtx.Request
    for i := 0; i < rtx.Option.retryTimes; i++ {
        resp, err = httputil.Do(ctx, req.Method, req.Url, req.Values, req.Header, rtx.client)
        if err != nil {
            debugPrintError(fmt.Errorf("req.Url:%v params:%v err:%w", req.Url, req.Values, err))
            rtx.Error = &OpError{Op: "httputil.DoRaw", Source: req.Url, Err: err}
            rtx.HttpCode = resp.HttpCode
            continue
        }
        rtx.RespHeaders = resp.Headers
        rtx.HttpCode = resp.HttpCode
        rtx.RespData = resp.Body
        if resp.HttpCode >= 200 && resp.HttpCode < 300 {
            break
        }
    }

    return
}

type Response struct {
    RespHeaders map[string][]string // 响应头
    HttpCode    int
    RespData    []byte
    Error       error
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/

func (c *Context) reset() {
    c.Index = -1
    c.engine = nil
    c.handlers = nil
    c.handleIndex = -1
    c.Keys = nil

    c.Request.Header = nil
    c.Request.Url = ""
    c.Request.Values = nil

    c.Response.RespHeaders = nil
    c.Response.HttpCode = 0
    c.Response.RespData = nil
    c.Response.Error = nil
}

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
// See example in GitHub.
func (c *Context) Next() {
    c.handleIndex++
    for c.handleIndex < int8(len(c.handlers)) {
        c.handlers[c.handleIndex](c)
        c.handleIndex++
    }
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
    return c.handleIndex >= abortIndex
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (c *Context) Abort() {
    c.handleIndex = abortIndex
}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
    c.mu.Lock()
    if c.Keys == nil {
        c.Keys = make(map[string]interface{})
    }

    c.Keys[key] = value
    c.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Context) Get(key string) (value interface{}, exists bool) {
    c.mu.RLock()
    value, exists = c.Keys[key]
    c.mu.RUnlock()
    return
}

func (c *Context) TransferToGoCtx() context.Context {
    ctx := context.Background()
    for k, v := range c.Keys {
        ctx = context.WithValue(ctx, k, v)
    }
    return ctx
}

/************************************/
/***** GOLANG.ORG/X/NET/CONTEXT *****/
/************************************/

// Deadline returns the time when work done on behalf of this context
// should be canceled. Deadline returns ok==false when no deadline is
// set. Successive calls to Deadline return the same results.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
    return
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. Done may return nil if this context can
// never be canceled. Successive calls to Done return the same value.
func (c *Context) Done() <-chan struct{} {
    return nil
}

// Err returns a non-nil error value after Done is closed,
// successive calls to Err return the same error.
// If Done is not yet closed, Err returns nil.
// If Done is closed, Err returns a non-nil error explaining why:
// Canceled if the context was canceled
// or DeadlineExceeded if the context's deadline passed.
func (c *Context) Err() error {
    return nil
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *Context) Value(key interface{}) interface{} {
    if key == 0 {
        return c.Request
    }
    if keyAsString, ok := key.(string); ok {
        val, _ := c.Get(keyAsString)
        return val
    }
    return nil
}
