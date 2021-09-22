/*
@Time : 2020/12/9 17:50
@Author : ZhaoJunfeng
@File : curlBase
*/
package curl

import (
    "context"
    "git.100tal.com/wangxiao_xesbiz_operation/gently-utils/logger"
    "git.100tal.com/wangxiao_xesbiz_operation/gently-utils/net/curl/httputil"
    "io/ioutil"
    "net/url"
    "strings"
    "sync"
)

const Tag = "curl"

type ICurl interface {
    // 增加一个请求，内部注入到请求池。param参数支持url.Values、String类型
    Add(method, target string, param interface{}, headers map[string]string, client ...interface{}) ICurl
    // 中间件注入
    Use(...HandlerFunc) ICurl
    // 执行，通过请求池发起http请求
    Exec(ctx context.Context) ICurl
    // 响应结果，数组下标对应requestList的下标
    GetResponses() []Response
    // 返回error。可以将所有请求错误归到一个error。方便调用者进行逻辑判断
    GetError() error
    // 返回所有错误，可以指定某个请求的错误，数组下标对应requestList的index索引，即使error为nil也会返回
    GetErrors() []error
    // 停止执行
    Close()
}

var _ ICurl = (*Curl)(nil)

type Curl struct {
    mu             sync.RWMutex
    Options        *Options
    engine         *Engine
    Handlers       HandlersChain
    RequestListCtx []*Context // 请求池上下文
    root           bool
}

var contextPool sync.Pool

func init() {
    contextPool.New = func() interface{} {
        return &Context{
            Index: -1,
        }
    }
}

func (c *Curl) Add(method, target string, params interface{}, headers map[string]string, client ...interface{}) ICurl {
    method = strings.ToUpper(method)

    reqIndex := int8(len(c.RequestListCtx))

    rCtx := contextPool.Get().(*Context)
    rCtx.reset()
    rCtx.engine = c.engine
    rCtx.Index = reqIndex
    rCtx.Option = *c.Options
    rCtx.Header = headers
    rCtx.Method = strings.ToUpper(method)
    rCtx.Url = target
    rCtx.Values = httputil.SetQueryValues(params)

    rCtx.client = httputil.ClientBuilder(client...)
    // 注入单个请求至请求池
    c.RequestListCtx = append(c.RequestListCtx, rCtx)
    c.root = false
    if method == "GET" && rCtx.Values != nil {
        Url, err := url.Parse(target)
        if err != nil {
            rCtx.Error = &OpError{Op: "write", Source: rCtx.Url, Err: err}
            return c
        }
        valuesByte, _ := ioutil.ReadAll(rCtx.Values)
        values := string(valuesByte)
        if Url.RawQuery != "" {
            Url.RawQuery += "&" + values
        } else {
            Url.RawQuery = values
        }

        c.RequestListCtx[reqIndex].Url = Url.String()
    }

    return c
}

func (c *Curl) SetOption(conf Options) ICurl {
    reqMax := len(c.RequestListCtx)
    if reqMax == 0 {
        // 如果requestList为空，则设置的为全局
        c.Options = &conf
        return c
    }
    reqIndex := reqMax - 1
    c.RequestListCtx[reqIndex].Option = conf

    return c
}

func (c *Curl) Use(middleware ...HandlerFunc) ICurl {
    if len(middleware) <= 0 {
        return c.returnObj()
    }

    var cIndex = -1
    if !c.root {
        reqMax := len(c.RequestListCtx)
        cIndex = reqMax - 1
    }

    for i := range middleware {
        // 过滤
        if middleware[i] == nil {
            continue
        }
        if c.root {
            c.Handlers = append(c.Handlers, middleware[i])
        } else {
            c.RequestListCtx[cIndex].handlers = append(c.RequestListCtx[cIndex].handlers, middleware[i])
        }
    }
    return c.returnObj()
}

func (c *Curl) Exec(ctx context.Context) ICurl {
    var err error
    var wg sync.WaitGroup
    var reqNum = len(c.RequestListCtx)
    if reqNum == 0 {
        logger.Wx(ctx, Tag, "RequestListCtx is empty!")
        return c
    }

    // workerPool schedule
    for k := range c.RequestListCtx {
        wg.Add(1)
        err = workerPool.Submit(c.handleContext(ctx, c.RequestListCtx[k], &wg))
        if err != nil {
            c.RequestListCtx[k].Error = err
        }
    }

    wg.Wait()
    return c
}

func (c *Curl) GetResponses() []Response {
    var responses = make([]Response, len(c.RequestListCtx))

    for i := range c.RequestListCtx {
        responses[i] = c.RequestListCtx[i].Response
    }

    return responses
}

func (c *Curl) GetErrors() []error {
    reqLen := len(c.RequestListCtx)
    if reqLen == 0 {
        return nil
    }

    var curlErrors = make([]error, reqLen)
    for i := range c.RequestListCtx {
        curlErrors[i] = c.RequestListCtx[i].Error
    }
    return curlErrors
}

func (c *Curl) GetError() error {
    reqLen := len(c.RequestListCtx)
    if reqLen == 0 {
        return nil
    }

    var errStrings Errors
    errStrings.Errors = make([]error, 0, reqLen)
    for i := range c.RequestListCtx {
        if c.RequestListCtx[i].Error != nil {
            errStrings.Errors = append(errStrings.Errors, c.RequestListCtx[i].Error)
        }
    }

    if errStrings.Errors == nil || len(errStrings.Errors) == 0 {
        return nil
    }

    return &errStrings
}

func (c *Curl) Close() {
    // 回收Context至对象池
    reqLen := len(c.RequestListCtx)
    if reqLen > 0 {
        for i := range c.RequestListCtx {
            c.RequestListCtx[i].reset()
            contextPool.Put(c.RequestListCtx[i])
        }
    }

    // 清空对象
    c.init()
}
