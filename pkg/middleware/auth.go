package middleware

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Cepave/grafana/pkg/bus"
	m "github.com/Cepave/grafana/pkg/models"
	"github.com/Cepave/grafana/pkg/setting"
	"github.com/Cepave/grafana/pkg/util"
	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type AuthOptions struct {
	ReqGrafanaAdmin bool
	ReqSignedIn     bool
}

func getRequestUserId(c *Context) int64 {
	userId := c.Session.Get(SESS_KEY_USERID)

	if userId != nil {
		return userId.(int64)
	}

	return 0
}

func getApiKey(c *Context) string {
	header := c.Req.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && parts[0] == "Bearer" {
		key := parts[1]
		return key
	}

	return ""
}

func accessForbidden(c *Context) {
	if c.IsApiRequest() {
		c.JsonApiErr(403, "Permission denied", nil)
		return
	}

	c.SetCookie("redirect_to", url.QueryEscape(setting.AppSubUrl+c.Req.RequestURI), 0, setting.AppSubUrl+"/")
	c.Redirect(setting.AppSubUrl + "/login")
}

func notAuthorized(c *Context) {
	if c.IsApiRequest() {
		c.JsonApiErr(401, "Unauthorized", nil)
		return
	}

	c.SetCookie("redirect_to", url.QueryEscape(setting.AppSubUrl+c.Req.RequestURI), 0, setting.AppSubUrl+"/")
	c.Redirect(setting.AppSubUrl + "/login")
}

func RoleAuth(roles ...m.RoleType) macaron.Handler {
	return func(c *Context) {
		ok := false
		for _, role := range roles {
			if role == c.OrgRole {
				ok = true
				break
			}
		}
		if !ok {
			accessForbidden(c)
		}
	}
}

/**
 * @function name:   getOpenFalconSessionUsername(sig string) string
 * @description:     This function returns username if "sig" cookie of Open-Falcon is valid.
 * @related issues:  OWL-201, OWL-159, OWL-124, OWL-115, OWL-110
 * @param:           sig string
 * @return:          username string
 * @author:          Don Hsieh
 * @since:           10/07/2015
 * @last modified:   12/09/2015
 * @called by:       func Auth(options *AuthOptions) macaron.Handler
 *                    in pkg/middleware/auth.go
 */
func getOpenFalconSessionUsername(sig string) string {
	if sig == "" {
		return ""
	}

	str := setting.ConfigOpenFalcon.Db.Addr
	db, err := sql.Open("mysql", str)
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(1000)
	defer db.Close()

	if err != nil {
		return ""
	}

	stmtOut, err := db.Prepare("SELECT id, uid, expired FROM uic.session WHERE sig = ?")
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer stmtOut.Close()

	var id int64
	var uid int64
	var expired string
	err = stmtOut.QueryRow(sig).Scan(&id, &uid, &expired)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	expiredTimeInt, err := strconv.ParseInt(expired, 10, 64)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	now := time.Now().Unix()
	isExpired := now > expiredTimeInt
	if isExpired {
		return ""
	}

	stmtOut, err = db.Prepare("SELECT name FROM uic.user WHERE id = ?")
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer stmtOut.Close()

	var name string
	err = stmtOut.QueryRow(uid).Scan(&name)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	return name
}

func getThirdPartyUsername(cookieName string, token string) string {
	if token == "" {
		return ""
	}

	infoURL := setting.ConfigOpenFalcon.LoginInfo
	req, _ := http.NewRequest("GET", infoURL, nil)
	cookie := http.Cookie{Name: cookieName, Value: token}
	req.AddCookie(&cookie)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return ""
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var nodes = make(map[string]interface{})
	if err := json.Unmarshal(body, &nodes); err != nil {
		return ""
	}
	username := ""
	email := ""
	if data, ok := nodes["data"].(map[string]interface{}); ok {
		if _, ok := data["username"]; ok {
			username = data["username"].(string)
		}
		if _, ok := data["email"]; ok {
			email = data["email"].(string)
		}
	}
	if username == "" {
		return ""
	}

	o := orm.NewOrm()
	var rows []orm.Params
	sql := "SELECT * FROM uic.user WHERE name= ? AND email = ? LIMIT 1"
	num, err := o.Raw(sql, username, email).Values(&rows)
	if err != nil {
		return ""
	} else if num > 0 {
		user := rows[0]
		username = user["name"].(string)
	} else {
		return ""
	}
	return username
}

func loginUserWithUser(user *m.User, c *Context) {
	if user == nil {
		log.Println(3, "User login with nil user")
	}

	days := 86400 * setting.LogInRememberDays
	c.SetCookie(setting.CookieUserName, user.Login, days, setting.AppSubUrl+"/")
	c.SetSuperSecureCookie(util.EncodeMd5(user.Rands+user.Password), setting.CookieRememberName, user.Login, days, setting.AppSubUrl+"/")

	c.Session.Set(SESS_KEY_USERID, user.Id)
}

/**
 * @function name:   func loginWithOpenFalconCookie(c *Context, username string)
 * @description:     This function gets user logged in if "sig" cookie of Open-Falcon is valid.
 * @related issues:  OWL-201, OWL-115, OWL-110
 * @param:           c *middleware.Context
 * @param:           username string
 * @return:          void
 * @author:          Don Hsieh
 * @since:           10/06/2015
 * @last modified:   12/09/2015
 * @called by:       func Auth(options *AuthOptions) macaron.Handler
 *                    in pkg/middleware/auth.go
 */
func loginWithOpenFalconCookie(c *Context, username string) {
	userQuery := m.GetUserByLoginQuery{LoginOrEmail: username}
	err := bus.Dispatch(&userQuery)
	if err == nil {
		user := userQuery.Result
		loginUserWithUser(user, c)
	} else {
		username = "admin"
		userQuery = m.GetUserByLoginQuery{LoginOrEmail: username}
		err := bus.Dispatch(&userQuery)
		if err == nil {
			user := userQuery.Result
			loginUserWithUser(user, c)
		} else {
			log.Println("Error =", err.Error())
		}
	}
}

func Auth(options *AuthOptions) macaron.Handler {
	return func(c *Context) {
		cookieName := setting.ConfigOpenFalcon.LoginCookie
		token := c.GetCookie(cookieName)
		if len(token) > 0 {
			username := getThirdPartyUsername(cookieName, token)
			loginWithOpenFalconCookie(c, username)
		} else {
			sig := c.GetCookie("sig")
			if len(sig) == 0  && options.ReqSignedIn && !c.AllowAnonymous {
				c.SetCookie(setting.CookieUserName, "", -1, setting.AppSubUrl+"/")
				c.SetCookie(setting.CookieRememberName, "", -1, setting.AppSubUrl+"/")
				c.Session.Destory(c)
				url := setting.ConfigOpenFalcon.Login + c.Req.RequestURI
				log.Println(url)
				c.Redirect(url)
				return
			}
			if !c.IsSignedIn {
				username := getOpenFalconSessionUsername(sig)
				loginWithOpenFalconCookie(c, username)
			}

			if !c.IsSignedIn && options.ReqSignedIn && !c.AllowAnonymous {
				notAuthorized(c)
				return
			}

			if !c.IsGrafanaAdmin && options.ReqGrafanaAdmin {
				accessForbidden(c)
				return
			}

		}
	}
}
