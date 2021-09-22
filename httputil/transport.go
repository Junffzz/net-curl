/*
@Time : 2020/11/12 20:47
@Author : ZhaoJunfeng
@File : transport
*/
package httputil

import (
    "net"
    "net/http"
    "net/url"
    "time"
)

var (
    timeout               int64 = 10000 // client总超时时间 默认10s。它的时间计算包括从连接(Dial)到读完response body
    disableKeepAlives     bool          // 是否禁用长连接，默认开启
    tLSHandshakeTimeout   int64 = 10    // 限制TLS握手使用的时间
    maxIdleConns          int   = 100   // 最大空闲连接数
    maxConnsPerHost       int   = 100   // 每个host的最大连接数
    maxIdleConnsPerHost   int   = 100   // 每个host的最大空闲连接数
    idleConnTimeout       int64 = 90    // 空闲连接在连接池中的保留时间
    expectContinueTimeout int64 = 1
    enableLog             bool  = false
    Transport             http.RoundTripper
    proxy                 func(reqURL *http.Request) (*url.URL, error) = http.ProxyFromEnvironment // 代理
)

func InitTransport() {
    Transport = &http.Transport{
        Proxy: proxy,
        DialContext: (&net.Dialer{
            Timeout: 30 * time.Second,
            // Deadline:  time.Now().Add(2 * time.Second), //超过这个时间后强制关闭连接，在连接无响应的时候回有用
            KeepAlive: 30 * time.Second, // 默认15s
        }).DialContext,
        TLSClientConfig:       nil,
        TLSHandshakeTimeout:   time.Duration(tLSHandshakeTimeout) * time.Second,
        DisableKeepAlives:     disableKeepAlives,
        DisableCompression:    false,
        MaxIdleConns:          maxIdleConns,
        MaxIdleConnsPerHost:   maxIdleConnsPerHost,
        MaxConnsPerHost:       maxConnsPerHost,
        IdleConnTimeout:       time.Duration(idleConnTimeout) * time.Second,
        ExpectContinueTimeout: time.Duration(expectContinueTimeout) * time.Second,
    }

}

func SetTransportProxyUrl(proxyUrl string) error {
    if proxyUrl != "" {
        proxyParseUrl, err := url.Parse(proxyUrl)
        if err != nil {
            return err
        }
        proxy = http.ProxyURL(proxyParseUrl)
    }
    return nil
}
