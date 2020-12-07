package main

import (
	"fmt"
	"github.com/qxyang2015/Go-000/Week02/service"
	"log"
	"net/http"
)

/*
Week02 作业题目：
我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？
*/

/*
解答：
1.在dao层使用Warp返回堆栈错误信息,不打印日志
2.在最上层service层接受错误，打印日志。
3.调用服务返回:异常码、跟错误
只应该处理error一次。可以有效减少噪声日志，并且在查找问题时又可以有效反应错误信息
*/
func main() {
	//调用service
	http.HandleFunc("/get_user_name", service.GetNameService) //设置访问的路由
	err := http.ListenAndServe(":9090", nil)                  //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Println("done!")
}
