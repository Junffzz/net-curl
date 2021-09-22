/*
@Time : 2020/12/9 18:00
@Author : ZhaoJunfeng
@File : httpClient
*/
package curl

import (
    "bytes"
    "context"
    "errors"
    "mime/multipart"
    "net/url"
)

func Call(ctx context.Context, method, target string, params interface{}, headers map[string]string, client ...interface{}) (ret Response, err error) {
    curlInc := New().Add(method, target, params, headers, client).Exec(ctx)
    defer curlInc.Close()

    if err = curlInc.GetError(); err != nil {
        return
    }

    // responses
    responses := curlInc.GetResponses()
    if len(responses) == 0 {
        return ret, &OpError{Op: "read", Source: target, Err: errors.New("responses is empty! ")}
    }
    ret = responses[0]

    return
}

func PostForm(ctx context.Context, target string, values url.Values, headers map[string]string, client ...interface{}) (ret Response, err error) {
    buffer := &bytes.Buffer{}

    writer := multipart.NewWriter(buffer)
    var keys []string
    for i := range values {
        keys = append(keys, i)
    }
    for _, v := range keys {
        _ = writer.WriteField(v, values.Get(v))
    }

    err = writer.Close()
    if err != nil {
        return
    }
    headers["Content-Type"] = writer.FormDataContentType()
    return PostRaw(ctx, target, buffer.String(), headers, client...)
}

func PostRaw(ctx context.Context, target string, params interface{}, headers map[string]string, client ...interface{}) (ret Response, err error) {
    curlInc := New().Add("POST", target, params, headers, client).Exec(ctx)
    defer curlInc.Close()

    if err = curlInc.GetError(); err != nil {
        return
    }

    responses := curlInc.GetResponses()
    ret = responses[0]
    return
}

func Get(ctx context.Context, target string, headers map[string]string, client ...interface{}) (ret Response, err error) {
    curlInc := New().Add("GET", target, nil, headers, client).Exec(ctx)
    defer curlInc.Close()

    if err = curlInc.GetError(); err != nil {
        return
    }

    responses := curlInc.GetResponses()
    ret = responses[0]
    return ret, nil
}

func Post(ctx context.Context, target string, params interface{}, headers map[string]string, client ...interface{}) (ret Response, err error) {
    curlInc := New().Add("POST", target, params, headers, client).Exec(ctx)
    defer curlInc.Close()

    if err = curlInc.GetError(); err != nil {
        return
    }

    responses := curlInc.GetResponses()
    ret = responses[0]
    return
}
