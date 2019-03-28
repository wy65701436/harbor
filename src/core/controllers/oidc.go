package controllers

import (
	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common/dao"
	"net/http"
)

// OIDCController ...
type OIDCController struct {
	beego.Controller
}

// Onboard ...
func (o *OIDCController) Onboard() {
	username := o.GetString("username")
	if username == "" {
		o.CustomAbort(http.StatusBadRequest, "Username is blank.")
	}

	userSub := o.GetSession("sub")
	if userSub == "" {
		o.CustomAbort(http.StatusInternalServerError, "User sub is blank.")
	}

	// TODO get secret with secret manager.
	// TODO get email from ID token.
	secret := ""
	email := ""

	err := dao.OnBoardOIDCUser(username, userSub.(string), secret, email)
	if err != nil {
		o.CustomAbort(http.StatusInternalServerError, "fail to onboard OIDC user")
	}
}
