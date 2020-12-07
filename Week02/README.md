## 1. error和panic
### 1.1. error

 如果函数返回了 value, error，你不能对这个 value 做任何假设，必须先判定 error。

### 1.2. panic 

1）panic意味着 fatal error(就是挂了),意味着代码不能继续运行。

2）对于真正意外的情况，那些表示不可恢复的程序错误，例如索引越界、不可恢复的环境问题、栈溢出，我们才使用 panic。

### 1.3. 防止出现野生Grountine措施
` 
func Go(f func()){
    if err:=recover();err!=nil{
        //***    
    } 
    go f()   
} 
`
## 2. 错误类型
### 2.1. 预定义错误（sentinel error）
    1)会成为 API 公共部分
    2)会在两个包之间创建了依赖
    3)尽可能避免 sentinel errors
### 2.2. MyError 是一个 type
    1)相比预定义错误，提供更多上下文
    2)避免成为上下文的一部分
### 2.3. 非透明错误处理
    1)代码和调用者之间的耦合最少
    2)缺少上下文
    可以提供判断错误类型的方式
## 3. 获取错误
### 3.1 错误处理方式
    1) 维持格式缩进,主线逻辑尽量不进行缩进
    2) 使用包含err的结构体,最终返回结构体err，减少err的产生
### 3.2 Wrap errors
    使用github.com/pkg/errors包提供的Warp方法返回错误堆栈信息
    1) Dao层使用Wrap打印堆栈错误信息，不打印错误日志
    2) 在最上层打印堆栈信息，打印日志,还可以通过errors.Cause(err)拿到原始错误信息
## 4. Go1.13 errors
    1) Unwrap
    2) Is
    3) As
