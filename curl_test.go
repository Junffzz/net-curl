/*
@Time : 2020/12/10 16:44
@Author : ZhaoJunfeng
@File : curl_test
*/
package curl_test

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    logger "git.100tal.com/wangxiao_go_lib/xesLogger"
    "git.100tal.com/wangxiao_go_lib/xesLogger/builders"
    "git.100tal.com/wangxiao_xesbiz_operation/gently-utils/net/curl"
    "git.100tal.com/wangxiao_xesbiz_operation/gently-utils/net/curl/httputil"
    "net/url"
    "os"
    "strconv"
    "testing"
    "time"
)

var headers = map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
var ctx = context.WithValue(context.TODO(), "x_trace_id", "testcurl123456789")

func TestMain(m *testing.M) {
    // logger
    config := logger.NewLogConfig()
    config.LogPath = "/var/logs/xeslog/curl.log"
    config.Level = "DEBUG" // 更新其他配置
    config.Console = false
    config.Rotate = false
    logger.InitLogWithConfig(config)
    builder := new(builders.TraceBuilder)
    builder.SetTraceDepartment("curl")
    builder.SetTraceVersion("0.1")
    logger.SetBuilder(builder)

    // test proxy
    _ = httputil.SetTransportProxyUrl("") // 测试环境代理

    appid := ""
    appkey := ""
    GetGatewayAuthHeader(headers, appid, appkey)

    code := m.Run()

    logger.Close()
    os.Exit(code)
}

func TestCurlRequests(t *testing.T) {
    // case2:支持链式批量调用，内部协程并行请求
    testCurl := curl.Default().
        Add("POST", "http://www.xueersi.com/Teacher/Teachers/infos", url.Values{"tIds[]": []string{"601", "607"}}, headers).
        Add("POST", "http://www.xueersi.com/Course/CourseInfo/getCourseInfos", url.Values{"course_id[]": []string{"54207", "60865"}}, headers).
        Add("POST", "http://www.xueersi.com", url.Values{"user_ids[]": []string{"2434061", "57788", "57789", "57849", "2386897"}}, headers).
        Exec(ctx)
    defer testCurl.Close()
    // 遍历所有Add请求的error
    if errs := testCurl.GetErrors(); len(errs) > 0 {
        for i := range errs {
            t.Logf("case2 curl.Add.Exec err:%v\n", errs[i])
        }
    }

    if err := testCurl.GetError(); err != nil {
        t.Errorf("GetError :%v\n", err)
    }

    responses := testCurl.GetResponses()
    for i := range responses {
        fmt.Printf("----testing case2 %+v\n", responses[i])
    }
}

func TestCurlMiddleware(t *testing.T) {
    tCurl := curl.Default()
    defer tCurl.Close()
    // 全局中间件
    tCurl.Use(func(c *curl.Context) {
        fmt.Printf("middleware 1 before.\n")
        c.Next()
        fmt.Printf("middleware 1 after.\n")
    })
    // request 1
    tCurl.Add("POST", "http://www.xueersi.com/Teacher/Teachers/infos", url.Values{"tIds[]": []string{"601", "607"}}, headers).
        Use(func(c *curl.Context) {
            fmt.Printf("teacherapi middleware 2 before.\n")
            c.Next()
            fmt.Printf("teacherapi middleware 2 after.\n")
        })
    // request 2
    tCurl.Add("POST", "http://www.xueersi.com/Teacher/Teachers/infos", url.Values{"tIds[]": []string{"601", "607"}}, headers)
    tCurl.Exec(ctx)

    responses := tCurl.GetResponses()
    for i := range responses {
        t.Logf("----testing curl.Middleware %+v\n", responses[i])
    }
}

/**
curl基准测试：
机器：8核16G
场景1：关闭日志，Add一个请求
goos: darwin
goarch: amd64
pkg: git.100tal.com/wangxiao_xesbiz_operation/gently-utils/net/curl
Benchmark_CurlAdd-8   	      60	  20060757 ns/op
PASS

场景2：关闭日志，Add六个请求
goos: darwin
goarch: amd64
pkg: git.100tal.com/wangxiao_xesbiz_operation/gently-utils/net/curl
Benchmark_CurlAdd-8   	      33	  35232967 ns/op
PASS

一次Add多个可以充分利用cpu的核数。
*/
func Benchmark_CurlAdd(b *testing.B) {
    ctx := context.TODO()
    // Add("POST", "http://www.xueersi.com/Course/CourseInfo/getCourseInfos", url.Values{"course_id[]": []string{"54207", "60865"}}, headers)
    for i := 0; i < b.N; i++ {
        testCurl := curl.New()
        for i := 0; i < 6; i++ {
            testCurl.Add("POST", "http://www.xueersi.com/Teacher/Teachers/infos", url.Values{"tIds[]": []string{"601", "607"}}, headers)
        }
        err := testCurl.Exec(ctx).GetError()
        if err != nil {
            b.Errorf("curl err:%v\n", err)
        }
        // 获取响应结果
        // testCurl.GetResponses()
        testCurl.Close()
        // b.Logf("%v\n",testCurl.GetResponses())
    }
}

func GetGatewayAuthHeader(header map[string]string, appId, appKey string) {
    header["X-Auth-Appid"] = appId
    now := strconv.Itoa(int(time.Now().Unix()))
    md5Ctx := md5.New()
    header["X-Auth-TimeStamp"] = now
    md5Ctx.Write([]byte(appId + "&" + now + appKey))
    cipherStr := md5Ctx.Sum(nil)
    signstr := hex.EncodeToString(cipherStr)
    header["X-Auth-Sign"] = signstr
}
