package tcore

import (
	"fmt"
	"github.com/smallnest/rpcx/share"
	"tframework.com/rpc/tcore/interface"
	"tframework.com/rpc/tcore/internal/define"
	_interface "tframework.com/rpc/tcore/internal/interface"
	"tframework.com/rpc/tcore/internal/server"
	"tframework.com/rpc/tcore/utils"
)

//***************************************************
//author tim.huang
//2022/8/10
//
//
//***************************************************

// CreateDefaultTServer
// @Description: 创建一个新的服务
// @return *ITServer
// @return error
func CreateDefaultTServer(module tframework.ITModule) (tframework.ITServer, error) {
	server := &tserver.TServer[tframework.ITModule]{}
	server.SetModule(module)
	server.SetConfigService(Config.(_interface.IServerConfigService))
	server.InitStruct()
	return server, nil
}

// CreateAndStartTCPServer
// @Description: 创建一个新的TCP服务
// @return *ITServer
// @return error
func CreateAndStartTCPServer() {
	//server := tcp.NewDefaultTCPServer(Config.GetTCPServer(), new(DefaultTCPService))
	//go server.Start()
}

type DefaultTCPService struct {
}

func (this *DefaultTCPService) Login(ct *share.Context, token string) {
	var (
		key, uuid, reqMetaDataKey string
		register                  bool
	)

	key = fmt.Sprintf(define.User_LoginToken_RedisKey, token)
	uuid = Redis.GetString(key)
	if uuid == "" {
		uuid = utils.GenerateSnowflakeId()
		Redis.Set(key, uuid, 0)
		register = true
	}
	ct.SetValue(tframework.ContextKey_UserId, uuid)
	//
	reqMetaDataKey = fmt.Sprintf(define.User_NodeMeta_RedisKey, uuid)
	reqMetaData := Redis.GetMap(reqMetaDataKey)
	reqMetaData[tframework.ContextKey_UserId] = uuid
	ct.SetValue(share.ReqMetaDataKey, reqMetaData)
	Log.InfoS("[TCP] login token %v , uuid %v register %v", token, uuid, register)
}
