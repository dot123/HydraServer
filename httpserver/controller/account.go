package controller

import (
	"HydraServer/httpserver/errors"
	"HydraServer/httpserver/ginx"
	account "HydraServer/httpserver/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/topfreegames/pitaya/v2/cluster"
)

var AccountControllerSet = wire.NewSet(wire.Struct(new(AccountController), "*"))

type AccountController struct {
	UserLogic        *account.UserLogic
	ServiceDiscovery cluster.ServiceDiscovery
}

// RegisterRoute 注册路由
func (a *AccountController) RegisterRoute(r *gin.RouterGroup) {
	r.POST("/account/register", a.Insert)
	r.POST("/account/changePwd", a.UpdatePwd)
	r.POST("/account/resetPwd", a.ResetPwd)
	r.GET("/serverList", a.ServerList)
}

func (a *AccountController) Insert(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.UserLogic.Insert(ctx, c.Query("username"), c.Query("password"), c.Query("hardware"))
	if err != nil {
		ginx.ResError(c, err)
		return
	}
	ginx.ResOk(c)
}

func (a *AccountController) UpdatePwd(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.UserLogic.UpdatePwd(ctx, c.Query("username"), c.Query("password"), c.Query("newPassword"))
	if err != nil {
		ginx.ResError(c, err)
		return
	}
	ginx.ResOk(c)
}

func (a *AccountController) ResetPwd(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.UserLogic.ResetPwd(ctx, c.Query("username"), c.Query("password"))
	if err != nil {
		ginx.ResError(c, err)
		return
	}
	ginx.ResOk(c)
}

func (a *AccountController) ServerList(c *gin.Context) {
	serverType := c.Query("serverType")
	if serverType == "" {
		err := errors.NewDefaultResponse("服务器类型不能为空")
		ginx.ResError(c, err)
		return
	}

	servers, err := a.ServiceDiscovery.GetServersByType(serverType)
	if err != nil {
		ginx.ResError(c, err)
		return
	}

	serverList := make([]string, len(servers))

	j := 0
	for k := range servers {
		serverList[j] = servers[k].Metadata["ip"] + ":" + servers[k].Metadata["port"]
		j++
	}

	ginx.ResData(c, serverList)
}
