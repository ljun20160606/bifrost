package tunnel

const (
	Delim = '\n'
)

const (
	// 注册
	MethodRegister = 1
	// 准备接收任务
	MethodConn = 2
)

type NodeInfo struct {
	// id
	Id string `json:"id"`
	// account
	Group string `json:"group"`
	// password
	Name string `json:"name"`
}
