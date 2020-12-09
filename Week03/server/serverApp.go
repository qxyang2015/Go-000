package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

//// 创建一个http服务
//type ServerDemo struct {
//	srver *http.Server
//}

type ServerHandler struct {
}

//实现ServeHTTP接口
func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	time.Sleep(100 * time.Second) //若是在请求过程中被停止就会报错
	_, _ = w.Write([]byte("进入请求"))
}

//新建Server服务
func NewServer(handler http.Handler, stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.Handle("/demo", &ServerHandler{})

	s := &http.Server{
		Addr:    ":8090",
		Handler: handler,
	}

	go func() {
		<-stop
		fmt.Println("触发推出信号！")
		s.Shutdown(context.Background())
	}()

	return s.ListenAndServe()
}
