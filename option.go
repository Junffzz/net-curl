/*
@Time : 2021/1/21 15:20
@Author : ZhaoJunfeng
@File : option
*/
package curl

const (
    defaultRetryTimes = 2
    defaultReqListCap = 1
)

type Option func(opts *Options)

func loadOptions(options ...Option) *Options {
    opts := newDefaultOptions()
    for _, option := range options {
        option(&opts)
    }
    return &opts
}

type Options struct {
    retryTimes int // 重试次数
    reqListCap int // 请求池大小
}

func newDefaultOptions() Options {
    p := Options{
        retryTimes: defaultRetryTimes,
        reqListCap: defaultReqListCap,
    }

    return p
}

func WithReqSize(size int) Option {
    return func(opts *Options) {
        opts.reqListCap = size
    }
}
