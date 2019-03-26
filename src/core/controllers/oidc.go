package controllers

import (
	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common/dao"
	"net/http"
)

type OIDCController struct {
	beego.Controller
}

func (o *OIDCController) Onboard() {
	username := o.GetString("username")
	if username == "" {
		o.CustomAbort(http.StatusBadRequest, "Username is blank.")
	}

	userSub := o.GetSession("sub")
	if userSub == "" {
		o.CustomAbort(http.StatusInternalServerError, "User sub is blank.")
	}

	err := dao.OnBoardOIDCUser(username, userSub.(string))
	if err != nil {
		o.CustomAbort(http.StatusInternalServerError, "fail to onboard OIDC user")
	}
}
