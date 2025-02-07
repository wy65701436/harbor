// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/pkg/oidc"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/i18n"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
)

// CommonController handles request from UI that doesn't expect a page, such as /SwitchLanguage /logout ...
type CommonController struct {
	api.BaseController
	i18n.Locale
}

// Render returns nil.
func (cc *CommonController) Render() error {
	return nil
}

// Prepare overwrites the Prepare func in api.BaseController to ignore unnecessary steps
func (cc *CommonController) Prepare() {}

func redirectForOIDC(ctx context.Context, username string) bool {
	if lib.GetAuthMode(ctx) != common.OIDCAuth {
		return false
	}
	u, err := user.Ctl.GetByName(ctx, username)
	if err != nil {
		log.Warningf("Failed to get user by name: %s, error: %v", username, err)
	}
	if u == nil {
		return true
	}
	us, err := user.Ctl.Get(ctx, u.UserID, &user.Option{WithOIDCInfo: true})
	if err != nil {
		log.Debugf("Failed to get OIDC user info for user, id: %d, error: %v", u.UserID, err)
	}
	if us != nil && us.OIDCUserMeta != nil {
		return true
	}
	return false
}

// Login handles login request from UI.
func (cc *CommonController) Login() {
	principal := cc.GetString("principal")
	password := cc.GetString("password")
	if redirectForOIDC(cc.Ctx.Request.Context(), principal) {
		ep, err := config.ExtEndpoint()
		if err != nil {
			log.Errorf("Failed to get the external endpoint, error: %v", err)
			cc.CustomAbort(http.StatusUnauthorized, "")
		}
		url := strings.TrimSuffix(ep, "/") + common.OIDCLoginPath
		log.Debugf("Redirect user %s to login page of OIDC provider", principal)
		// Return a json to UI with status code 403, as it cannot handle status 302
		cc.Ctx.Output.Status = http.StatusForbidden
		err = cc.Ctx.Output.JSON(struct {
			Location string `json:"redirect_location"`
		}{url}, false, false)
		if err != nil {
			log.Errorf("Failed to write json to response body, error: %v", err)
		}
		return
	}

	user, err := auth.Login(cc.Context(), models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		cc.CustomAbort(http.StatusUnauthorized, "")
	}
	cc.PopulateUserSession(*user)
}

// LogOut Harbor UI
func (cc *CommonController) LogOut() {
	// logout session for the OIDC
	if lib.GetAuthMode(cc.Context()) == common.OIDCAuth {
		idToken := cc.GetSession(tokenKey).([]byte)
		idTokenStr := string(idToken)
		log.Info(" ============== ")
		log.Info(idTokenStr)
		log.Info(" ============== ")
		token := oidc.Token{}
		//var url string

		if err := cc.DestroySession(); err != nil {
			log.Errorf("Error occurred in LogOut: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}

		if err := json.Unmarshal(idToken, &token); err != nil {
			log.Errorf("Error occurred in Unmarshal: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}

		log.Info(" ============== ")
		log.Info(token.RawIDToken)
		log.Info(" ============== ")

		if token.RawIDToken != "" {
			oidcLogoutURL := fmt.Sprintf(
				"https://10.164.142.200:8443/realms/myrealm/protocol/openid-connect/logout?id_token_hint=%s&post_logout_redirect_uri=%s",
				url.QueryEscape(token.RawIDToken),
				url.QueryEscape("/harbor/projects"),
			)

			log.Info(" ============== ")
			log.Info(oidcLogoutURL)
			log.Info(" ============== ")

			// Redirect user to OIDC Logout
			cc.Controller.Redirect(oidcLogoutURL, http.StatusFound)
		}

		//if !token.Valid() {
		//	log.Info("Refreshing token")
		//	token, err := oidc.RefreshToken(cc.Context(), &token)
		//	if err != nil {
		//		log.Errorf("Refreshing token: %v", err)
		//		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		//	}
		//	//tb, err := json.Marshal(token)
		//	//if err != nil {
		//	//	log.Errorf("failed to encode the refreshed token, error: %v", err)
		//	//	cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		//	//}
		//	//key, err := oidc.KeyLoader.EncryptKey()
		//	//encToken, _ := utils.ReversibleEncrypt(string(tb), key)
		//	//oidcUser.Token = encToken
		//	//// only updates the token column of the record
		//	//err = dm.metaDao.Update(cc.Context(), oidcUser, "token")
		//	//if err != nil {
		//	//	log.Errorf("Failed to persist token, user id: %d, error: %v", oidcUser.UserID, err)
		//	//}
		//	url = fmt.Sprintf("https://10.164.142.200:8443/realms/myrealm/protocol/openid-connect/logout?post_logout_redirect_uri=https://10.164.142.200/harbor/projects&id_token_hint=%s", token.RawIDToken)
		//	log.Info("Token refreshed and persisted")
		//} else {
		//	url = fmt.Sprintf("https://10.164.142.200:8443/realms/myrealm/protocol/openid-connect/logout?post_logout_redirect_uri=https://10.164.142.200/harbor/projects&id_token_hint=%s", token.RefreshToken)
		//}

		// https://10.164.142.200:8443/realms/myrealm/protocol/openid-connect/logout?post_logout_redirect_uri=https://10.164.142.200/harbor/projects&id_token_hint=123
		//cc.Controller.Redirect(url, http.StatusFound)
	}
}

// UserExists checks if user exists when user input value in sign in form.
func (cc *CommonController) UserExists() {
	ctx := cc.Context()
	flag, err := config.SelfRegistration(ctx)
	if err != nil {
		log.Errorf("Failed to get the status of self registration flag, error: %v, disabling user existence check", err)
	}
	securityCtx, ok := security.FromContext(ctx)
	isAdmin := ok && securityCtx.IsSysAdmin()
	if !flag && !isAdmin {
		cc.CustomAbort(http.StatusPreconditionFailed, "self registration deactivated, only sysadmin can check user existence")
	}

	target := cc.GetString("target")
	value := cc.GetString("value")

	var query *q.Query
	switch target {
	case "username":
		query = q.New(q.KeyWords{"Username": value})
	case "email":
		query = q.New(q.KeyWords{"Email": value})
	}

	n, err := user.Ctl.Count(ctx, query)
	if err != nil {
		log.Errorf("Error occurred in UserExists: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	cc.Data["json"] = n > 0
	if err := cc.ServeJSON(); err != nil {
		log.Errorf("failed to serve json: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

func init() {
	// conf/app.conf -> os.Getenv("config_path")
	configPath := os.Getenv("CONFIG_PATH")
	if len(configPath) != 0 {
		log.Infof("Config path: %s", configPath)
		if err := web.LoadAppConfig("ini", configPath); err != nil {
			log.Errorf("failed to load app config: %v", err)
		}
	}
}
