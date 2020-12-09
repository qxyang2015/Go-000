package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	pkg_errors "github.com/pkg/errors"
	"github.com/qxyang2015/Go-000/Week02/dao"
	"net/http"
)

//service层
func GetNameService(w http.ResponseWriter, r *http.Request) {
	retRes := Response{
		Code: 0,
		Msg:  "OK",
	}

	defer func() {
		jsonRes, err := json.Marshal(retRes)
		if err != nil {
			fmt.Fprintf(w, "json marshal error:[%v]", err.Error())
			return
		}
		//客户端返回
		fmt.Fprintf(w, "%s", string(jsonRes))
	}()

	//调用Dao函数
	name, err := dao.GetUserName("qxyang")
	//查询数据库为空err处理
	if errors.Is(err, sql.ErrNoRows) {
		//打印Error日志,打印所有堆栈信息
		fmt.Printf("查询数据库为空 %+v\n", err)
		//填充response
		retRes.Code = -2
		retRes.Msg = fmt.Sprintf("查询数据库为空")
		return
	}
	//通用错误处理
	if err != nil {
		//打印Error日志,打印所有堆栈信息
		fmt.Printf("Error:Serverce查询Dao错误 %+v\n", err)
		//填充response
		retRes.Code = -1
		retRes.Msg = fmt.Sprintf("查询用户名出现错误")
		return
	}
	retRes.Data = name
}
