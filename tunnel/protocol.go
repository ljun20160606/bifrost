package tunnel

import "github.com/satori/go.uuid"

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

func (n *NodeInfo) Path() string {
	return BuildRealGroup(n.Group, n.Name)
}

const groupDelim = "/"

// 生成真正的group
func BuildRealGroup(group, name string) string {
	return group + groupDelim + name
}

func NewUUID() string {
	uuids, _ := uuid.NewV4()
	return uuids.String()
}
