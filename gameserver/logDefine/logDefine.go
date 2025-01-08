package logDefine

import "HydraServer/pkg/log"

var (
	Monitor    = log.Register("monitor", "mem", 1800, true)
	Online_Num = log.Register("online_num", "num,time", 1800, true)
	Login      = log.Register("login", "uid,rid,username,nick_name,ip", 1800, true)
	Logout     = log.Register("logout", "uid,rid,username,nick_name,ip,reason,online_time", 1800, true)
)
