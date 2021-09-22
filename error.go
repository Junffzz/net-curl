/*
@Time : 2020/12/23 17:52
@Author : ZhaoJunfeng
@File : error
*/
package curl

import (
    "os"
    "syscall"
)

// 定义立体错误类型体系
type Error interface {
    error
    Timeout() bool   // Is the error a timeout?
    Temporary() bool // Is the error temporary?
}

type Errors struct {
    Errors []error
}

func (e *Errors) Error() string {
    if e == nil || len(e.Errors) == 0 {
        return "<nil>"
    }
    s := ""
    for i := range e.Errors {
        if e.Errors[i] != nil {
            s += e.Errors[i].Error()
        }

    }
    return s
}

var _ Error = (*OpError)(nil)

// OpError is the error type usually returned by functions in the curl
// package. It describes the operation, network type, and address of
// an error.
type OpError struct {
    // Op is the operation which caused the error, such as
    // "read" or "write".
    Op string

    // Source is the Curl Request's url.
    Source string

    // Err is the error that occurred during the operation.
    Err error
}

func (e *OpError) Unwrap() error { return e.Err }

func (e *OpError) Error() string {
    if e == nil {
        return "<nil>"
    }
    s := e.Op
    s += " " + e.Source + ": " + e.Err.Error()
    return s
}

type timeout interface {
    Timeout() bool
}

func (e *OpError) Timeout() bool {
    if ne, ok := e.Err.(*os.SyscallError); ok {
        t, ok := ne.Err.(timeout)
        return ok && t.Timeout()
    }
    t, ok := e.Err.(timeout)
    return ok && t.Timeout()
}

type temporary interface {
    Temporary() bool
}

func (e *OpError) Temporary() bool {
    // Treat ECONNRESET and ECONNABORTED as temporary errors when
    // they come from calling accept. See issue 6163.
    if e.Op == "accept" && isConnError(e.Err) {
        return true
    }

    if ne, ok := e.Err.(*os.SyscallError); ok {
        t, ok := ne.Err.(temporary)
        return ok && t.Temporary()
    }
    t, ok := e.Err.(temporary)
    return ok && t.Temporary()
}

func isConnError(err error) bool {
    if se, ok := err.(syscall.Errno); ok {
        return se == syscall.ECONNRESET || se == syscall.ECONNABORTED
    }
    return false
}
