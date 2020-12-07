package service

type Request struct {
	Id string
}

type Response struct {
	Code int
	Msg  string
	Data interface{}
}
