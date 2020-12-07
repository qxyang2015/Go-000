package dao

import (
	"database/sql"
	"github.com/pkg/errors"
)

type DaoConfig struct {
	UserDB sql.DB
}

var Dao DaoConfig

//初始化Dao
func Init() {
}

func GetUserName(id string) (string, error) {
	var name string
	err := Dao.UserDB.QueryRow("select name from tableName where id =?", id).Scan(&name)
	if err != nil {
		return name, errors.Wrap(err, "GetUserName")
	}
	return name, nil
}
