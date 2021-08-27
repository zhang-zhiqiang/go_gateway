package controller

import (
	"encoding/json"
	"github.com/e421083458/go_gateway_demo/dao"
	"github.com/e421083458/go_gateway_demo/dto"
	"github.com/e421083458/go_gateway_demo/middleware"
	"github.com/e421083458/go_gateway_demo/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct {
}

func AdminRegister(group *gin.RouterGroup) {
	adminLogin := &AdminController{}
	group.GET("/admin_info", adminLogin.AdminInfo)
	group.POST("/change_pwd", adminLogin.ChangePwd)
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

// Admin godoc
// @Summary 修改密码
// @Description 修改密码
// @Tags 管理员接口
// @ID /admin/change_pwd
// @Accept  json
// @Param body body dto.ChangePwdInput true "body"
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_pwd [post]
func (adminlogin *AdminController) ChangePwd(c *gin.Context) {
	params := &dto.ChangePwdInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	sess := sessions.Default(c)
	sessionInfo := sess.Get(public.AdminSessionInfoKey)
	sessionInfoStr := sessionInfo.(string)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(sessionInfoStr), adminSessionInfo); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	adminInfo := &dao.Admin{}
	adminInfo, err = adminInfo.Find(c, tx, (&dao.Admin{UserName: adminSessionInfo.UserName}))
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	saltPassword := public.GenSaltPassword(adminInfo.Salt, params.Password)

	adminInfo.Password = saltPassword

	if err := adminInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 2004, err)
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
