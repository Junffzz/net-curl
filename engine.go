/*
@Time : 2021/3/3 16:40
@Author : ZhaoJunfeng
@File : curl
*/
package curl

var _ ICurl = (*Engine)(nil)

type Engine struct {
    Curl

    Option *Options
}

/**
 * 新建Curl引擎
 * @date: 2021/3/4
 */
func New(options ...Option) *Engine {
    ops := loadOptions(options...)

    engine := &Engine{
        Curl: Curl{
            Handlers: nil,
            root:     true,
            Options:  ops,
        },
        Option: ops,
    }
    engine.Curl.engine = engine

    // 初始化请求池
    engine.Curl.RequestListCtx = make([]*Context, 0, ops.reqListCap)
    return engine
}

func (e *Engine) Use(middleware ...HandlerFunc) ICurl {
    e.Curl.Use(middleware...)
    return e
}

func (e *Engine) Close() {
    e.Curl.Close()
}
