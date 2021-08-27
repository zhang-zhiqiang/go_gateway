package controller

import (
	"encoding/json"
	"github.com/e421083458/go_gateway_demo/dto"
	"github.com/e421083458/go_gateway_demo/middleware"
	"github.com/e421083458/go_gateway_demo/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
}

func AdminRegister(group *gin.RouterGroup) {
	adminLogin := &AdminController{}
	group.GET("/admin_info", adminLogin.AdminInfo)
}

// Admin godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (adminlogin *AdminController) AdminInfo(c *gin.Context) {
	sess := sessions.Default(c)
	sessionInfo := sess.Get(public.AdminSessionInfoKey)
	sessionInfoStr := sessionInfo.(string)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(sessionInfoStr), adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	out := &dto.AdminInfoOutput{
		Id:           adminSessionInfo.Id,
		Name:         adminSessionInfo.UserName,
		LoginTime:    adminSessionInfo.LoginTime,
		Avatar:       "https://img3.zhiupimg.cn/group1/M00/03/8C/rBAUDFwkiAGAKXzTAAAK_eDTpQ8201.jpg",
		Introduction: "",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(c, out)
}
