package dto

import (
	"github.com/e421083458/go_gateway_demo/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminInfoOutput struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
	LoginTime    time.Time `json:"login_time"`
}

type ChangePwdInput struct {
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"` // 新密码
}

func (param *ChangePwdInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
