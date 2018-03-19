package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
)

const catalogURLPattern = `/v2/_catalog`

type ListReposHandler struct {
	next http.Handler
}

func (lrh ListReposHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	listReposFlag := matchListRepos(req)
	if listReposFlag {
		rec = httptest.NewRecorder()
		lrh.next.ServeHTTP(rec, req)
		if rec.Result().StatusCode != http.StatusOK {
			copyResp(rec, rw)
			return
		}
		var ctlg struct {
			Repositories []string `json:"repositories"`
		}
		decoder := json.NewDecoder(rec.Body)
		if err := decoder.Decode(&ctlg); err != nil {
			log.Errorf("Decode repositories error: %v", err)
			copyResp(rec, rw)
			return
		}
		var entries []string
		for repo := range ctlg.Repositories {
			log.Debugf("the repo in the reponse %s", ctlg.Repositories[repo])
			exist := dao.RepositoryExists(ctlg.Repositories[repo])
			if exist {
				entries = append(entries, ctlg.Repositories[repo])
			}
		}
		type Repos struct {
			Repositories []string `json:"repositories"`
		}
		resp := &Repos{Repositories: entries}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			log.Errorf("Encode repositories error: %v", err)
			copyResp(rec, rw)
			return
		}

		for k, v := range rec.Header() {
			rw.Header()[k] = v
		}
		clen := len(respJSON)
		rw.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(clen))
		rw.Write(respJSON)
		return
	}
	lrh.next.ServeHTTP(rw, req)
}

// matchListRepos checks if the request looks like a request to list repositories.
func matchListRepos(req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}
	re := regexp.MustCompile(catalogURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 1 {
		return true
	}
	return false
}
