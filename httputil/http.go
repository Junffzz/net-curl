/*
@Time : 2020/11/12 20:26
@Author : ZhaoJunfeng
@File : http.go
*/
package httputil

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "reflect"
    "time"
)

type HttpResp struct {
    Body     []byte
    Headers  map[string][]string
    HttpCode int
}

func Do(ctx context.Context, method, target string, postBody io.Reader, headers map[string]string, client ...interface{}) (resp HttpResp, err error) {
    var req *http.Request
    var httpResp *http.Response
    var httpclient *http.Client

    httpclient = ClientBuilder(client...)

    req, err = http.NewRequestWithContext(ctx, method, target, postBody)
    if err != nil {
        return resp, fmt.Errorf("http.NewRequestWithContext fail:method:%v,url:%v,%w", method, target, err)
    }

    if headers != nil && len(headers) > 0 {
        for k, v := range headers {
            req.Header.Set(k, v)
        }
    }

    httpResp, err = httpclient.Do(req)
    if err != nil {
        return resp, fmt.Errorf("httpclient.Do fail:%w", err)
    }

    // header
    resp.Headers = httpResp.Header

    // body
    resp.Body, err = ioutil.ReadAll(httpResp.Body)
    err = httpResp.Body.Close()

    // statusCode
    if httpResp.StatusCode != 200 || err != nil {
        err = fmt.Errorf("http req:%+v ;resp status:%v,%w", req, httpResp.StatusCode, err)
    }
    resp.HttpCode = httpResp.StatusCode
    return resp, err
}

func bodyFormat(bodyParams map[string]interface{}) (result string) {
    params := url.Values{}
    for k, v := range bodyParams {
        switch reflect.TypeOf(v).Kind() {
        case reflect.String:
            params.Add(k, v.(string))
            break
        default:
            vJson, _ := json.Marshal(v)
            params.Add(k, string(vJson))
            break
        }
    }
    return params.Encode()
}

func ClientBuilder(client ...interface{}) *http.Client {
    var httpclient *http.Client
    var isClient = false
    if client != nil && len(client) == 1 {
        if c, ok := client[0].(*http.Client); ok && c != nil {
            httpclient = c
            isClient = true
        }
    }

    if isClient == false {
        InitTransport()
        httpclient = &http.Client{
            Transport:     Transport,
            CheckRedirect: nil,
            Jar:           nil,
            Timeout:       time.Millisecond * time.Duration(timeout),
        }
    }

    return httpclient
}

func SetQueryValues(query interface{}) (bodyReader io.Reader) {
    if query == nil {
        return nil
    }
    switch query.(type) {
    case url.Values:
        bodyReader = bytes.NewReader([]byte(query.(url.Values).Encode()))
    case io.Reader:
        bodyReader = query.(io.Reader)
    case string:
        bodyReader = bytes.NewReader([]byte(query.(string)))
    case []byte:
        bodyReader = bytes.NewReader(query.([]byte))
    default:

    }

    return bodyReader
}
