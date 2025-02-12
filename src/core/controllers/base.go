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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/oidc"
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
	if lib.GetAuthMode(cc.Context()) == common.OIDCAuth {
		tk := cc.GetSession(tokenKey).([]byte)
		tkStr := string(tk)
		log.Info(" ============== ")
		log.Info(tkStr)
		log.Info(" ============== ")
		token := oidc.Token{}
		//var url string

		if err := cc.DestroySession(); err != nil {
			log.Errorf("Error occurred in LogOut: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}

		if err := json.Unmarshal(tk, &token); err != nil {
			log.Errorf("Error occurred in Unmarshal: %v", err)
			cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}

		log.Info(" ============== ")
		log.Info(token.RefreshToken)
		log.Info(" ============== ")

		if token.RefreshToken != "" {
			// logout session for the OIDC
			//ep, err := config.ExtEndpoint()
			//if err != nil {
			//	log.Errorf("Failed to get the external endpoint, error: %v", err)
			//	cc.CustomAbort(http.StatusUnauthorized, "")
			//}
			//url := strings.TrimSuffix(ep, "/") + common.OIDCLoginoutPath
			//log.Debugf("Redirect to logout page of OIDC provider")
			//// Return a json to UI with status code 403, as it cannot handle status 302
			//cc.Ctx.Output.Status = http.StatusForbidden
			//err = cc.Ctx.Output.JSON(struct {
			//	Location string `json:"redirect_location"`
			//}{url}, false, false)
			//if err != nil {
			//	log.Errorf("Failed to write json to response body, error: %v", err)
			//}

			sessionType, err := getSessionType(token.RefreshToken)
			if err != nil {
				cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
			}

			// for the session tpyes
			// 1, for the offline type, it should call the end_session_endpoint with refresh_token.
			// 2, for the refresh type, it should call the end_session_endpoint with id_token.
			log.Info(" 111============== ")
			log.Info(sessionType)
			log.Info(" 1111============== ")
			if sessionType == "Offline" {
				if err := revokeOIDCRefreshToken(token.RefreshToken, "", ""); err != nil {
					log.Error(err)
					cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
				}
			} else if sessionType == "Refresh" {
				logoutURL := fmt.Sprintf(
					"https://10.164.142.200:8443/realms/myrealm/protocol/openid-connect/logout?id_token_hint=%s&post_logout_redirect_uri=%s",
					url.QueryEscape(token.RawIDToken),
					url.QueryEscape("https://10.164.142.200/harbor/projects"),
				)
				_, _ = http.Get(logoutURL)
				return
			}

			//oidcLogoutURL := fmt.Sprintf(
			//	"https://10.164.142.200:8443/realms/myrealm/protocol/openid-connect/logout?id_token_hint=%s&post_logout_redirect_uri=%s",
			//	url.QueryEscape(token.RawIDToken),
			//	url.QueryEscape("https://10.164.142.200/harbor/projects"),
			//)
			//
			//log.Info(" ============== ")
			//log.Info(oidcLogoutURL)
			//log.Info(" ============== ")

			//url := strings.TrimSuffix(ep, "/") + common.OIDCLoginPath
			//cc.Ctx.Output.Status = http.StatusOK
			//err := cc.Ctx.Output.JSON(struct {
			//	Location string `json:"redirect_location"`
			//}{oidcLogoutURL}, false, false)
			//if err != nil {
			//	log.Errorf("Failed to write json to response body, error: %v", err)
			//}

			// Redirect user to OIDC Logout
			//cc.Controller.Redirect(oidcLogoutURL, http.StatusFound)
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

// getSessionType determines if the session is offline by decoding the refresh token or not
func getSessionType(refreshToken string) (string, error) {
	parts := strings.Split(refreshToken, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid refresh token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode refresh token: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to unmarshal refresh token: %w", err)
	}

	typ, ok := claims["typ"].(string)
	if !ok {
		return "", errors.New("missing 'typ' claim in refresh token")
	}

	return typ, nil
}

// revokeOIDCRefreshToken revokes an offline session using the refresh token
func revokeOIDCRefreshToken(refreshToken, clientID, clientSecret string) error {
	logoutURL := "https://10.164.142.200:8443/realms/myrealm/protocol/openid-connect/logout"

	// Prepare form data
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)

	// Create request
	req, err := http.NewRequest("POST", logoutURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed, status: %d", resp.StatusCode)
	}

	fmt.Println("Logout successful")
	return nil
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
