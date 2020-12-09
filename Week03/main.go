package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/qxyang2015/Go-000/Week03/server"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

/*
1. 基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出
*/

func main() {
	stop := make(chan struct{})
	//使用一个携带上下文信息的errgroup
	g, _ := errgroup.WithContext(context.Background())

	//服务1
	g.Go(func() error {
		muxDemo1 := http.NewServeMux()
		return server.NewServer(muxDemo1, stop)
	})

	//服务2
	g.Go(func() error {
		muxDemo2 := http.NewServeMux()
		return server.NewServer(muxDemo2, stop)
	})

	//监听signal信号，当接收到退出相关信号退出
	g.Go(func() error {
		quit := make(chan os.Signal)
		//监听到指定信号就给quit
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
		for {
			fmt.Println("等待退出信号")
			select {
			case <-quit:
				return errors.New("收到退出信号") //这个人造错误用来退出errgroup
			}
		}
	})
	err := g.Wait()
	if err != nil {
		close(stop)
	}
	fmt.Println("Done!")
}
