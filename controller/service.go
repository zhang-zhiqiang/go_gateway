package controller

import (
	"fmt"
	"github.com/e421083458/go_gateway/public"
	"github.com/e421083458/go_gateway_demo/dao"
	"github.com/e421083458/go_gateway_demo/dto"
	"github.com/e421083458/go_gateway_demo/middleware"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

type ServiceController struct {
}

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.GET("/service_delete", service.ServiceDelete)
	group.POST("/service_add_http", service.ServiceAddHttp)
}

// ServiceList godoc
// @Summary 服务列表
// @Description 服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (s *ServiceController) ServiceList(c *gin.Context) {
	params := &dto.ServiceListInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3002, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.PageList(c, tx, params)
	if err != nil {
		middleware.ResponseError(c, 3003, err)
		return
	}

	outList := []dto.ServiceListItemOutput{}
	for _, listItem := range list {
		serviceDetail, err := listItem.ServiceDetail(c, tx, &listItem)
		if err != nil {
			middleware.ResponseError(c, 3004, err)
			return
		}

		serviceAddr := "unknow"
		clusterIp := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIp, clusterPort, serviceDetail.HTTPRule.Rule)
		}

		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}

		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.TCPRule.Port)
		}

		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIp, serviceDetail.GRPCRule.Port)
		}

		ipList := serviceDetail.LoadBalance.GetIPListByModel()
		outItem := dto.ServiceListItemOutput{
			ID:          listItem.ID,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         0,
			Qpd:         0,
			TotalNode:   len(ipList),
		}
		outList = append(outList, outItem)
	}

	out := &dto.ServiceListOutput{
		Total: total,
		List:  outList,
	}
	middleware.ResponseSuccess(c, out)
}

// ServiceDelete godoc
// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [get]
func (s *ServiceController) ServiceDelete(c *gin.Context) {
	params := &dto.ServiceDeleteInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ID: params.Id}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		middleware.ResponseError(c, 3002, err)
		return
	}

	serviceInfo.IsDelete = 1
	if err := serviceInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 3003, err)
		return
	}

	middleware.ResponseSuccess(c, "")
}

// ServiceAddHttp godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (s *ServiceController) ServiceAddHttp(c *gin.Context) {
	params := &dto.ServiceUpdateHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4000, err)
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 4001, err)
		return
	}

	tx = tx.Begin()

	serviceInfo := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
	}

	if _, err = serviceInfo.Find(c, tx, serviceInfo); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 4002, errors.New("服务已存在"))
		return
	}

	httpUrl := &dao.HttpRule{RuleType: 0, Rule: params.Rule}
	if _, err := httpUrl.Find(c, tx, httpUrl); err == nil {
		tx.Rollback()
		middleware.ResponseError(c, 4003, errors.New("接入前缀或域名已存在"))
		return
	}
	if len(strings.Split(params.IpList, "\n")) != len(strings.Split(params.WeightList, "\n")) {
		tx.Rollback()
		middleware.ResponseError(c, 4003, errors.New("IP列表于权重列表数量不一致"))
		return
	}

	serviceInfoModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := serviceInfoModel.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 4004, err)
		return
	}

	httpModel := dao.HttpRule{
		ServiceID:      serviceInfoModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedWebsocket:  params.NeedWebsocket,
		NeedStripUri:   params.NeedStripUri,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpModel.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 4005, err)
		return
	}

	accessControlModel := dao.AccessControl{
		ServiceID:         serviceInfoModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControlModel.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 4006, err)
		return
	}

	loadBalanceModel := dao.LoadBalance{
		ServiceID:              serviceInfoModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadBalanceModel.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 4007, err)
		return
	}

	tx.Commit()
	middleware.ResponseSuccess(c, "")
}
