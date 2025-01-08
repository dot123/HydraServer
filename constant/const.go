package constant

const (
	RIdKey    = "rid"
	UIdKey    = "uid"
	ServerId  = "serverId"
	ServerKey = "serverKey"
)

// 断开连接提示码
const (
	DISCONNECT_NORMAL       = 1 // 正常断开连接
	DISCONNECT_KICK_SINGLE  = 2 // 管理后台踢人（单独）
	DISCONNECT_KICK_ALL     = 3 // 管理后台踢人（全服）
	DISCONNECT_DUPLICATE    = 4 // 重复帐号登录
	DISCONNECT_AUTH_FAIL    = 5 // 登录认证失败
	DISCONNECT_NORMAL_AFTER = 6 // 正常断开后几秒后销毁
	DISCONNECT_MAX_ONLINE   = 7 // 最大登录人数限制
	DISCONNECT_BANUSER      = 8 // 封号
	DISCONNECT_BANIP        = 9 // 封IP
)
