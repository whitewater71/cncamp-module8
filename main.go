package main

import (
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    _ "net/http/pprof"
    "github.com/golang/glog"
    "os"
//    "os/signal"
//    "syscall"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    select {
    case <-ctx.Done():
        err := ctx.Err()
        fmt.Println("server:", err)
        internalError := http.StatusInternalServerError
        http.Error(w, err.Error(), internalError)
    default:
    for k, v := range r.Header {
        for _, value := range v {
            w.Header().Set(k, value)
            fmt.Printf("%s=%s\n", k, value)
        }
    }
    w.Header().Set("Version", os.Getenv("VERSION"))
    fmt.Printf("VERSION=%s\n", os.Getenv("VERSION"))

    forwarded := r.Header.Get ("X-FORWARDED-FOR")
    if forwarded != "" {
        fmt.Printf("Remote IP=%s\n", forwarded)
    } else {
        fmt.Printf("Remote IP=%s\n", r.RemoteAddr)
    }
    }
}

func wrapHandlerWithLogging (wrappedHandler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        log.Printf("--> %s %s", req.Method, req.URL.Path)
        lrw := NewLoggingResponseWriter(w)
        wrappedHandler.ServeHTTP(lrw, req)
        statusCode := lrw.statusCode
        log.Printf("<-- %d %s", statusCode, http.StatusText(statusCode))
    })
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
    return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func healthz(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    io.WriteString(w, "ok\n")
}

func main() {
    flag.Set("v", "4")
    flag.Parse()
    glog.V(2).Info("Starting http server...")
    rootHandler := wrapHandlerWithLogging(http.HandlerFunc(handleRoot))
    http.Handle("/", rootHandler)
    http.HandleFunc("/healthz", healthz)
    //sigs := make(chan os.Signal, 1)
    //signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    err := http.ListenAndServe(":" + os.Getenv("MY_SERVICE_PORT"), nil)
    if err != nil {
        log.Fatal(err)
    }
    //sig := <-sigs
    //fmt.Println(sig)
}
