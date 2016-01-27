package api

import (
	"github.com/Cepave/grafana/pkg/middleware"
	"github.com/Cepave/grafana/pkg/setting"
)

/**
 * @function name:   func Cpu(c *middleware.Context)
 * @description:     This function redirects URL "/boss/cpu/:host" to cpu chart.
 * @related issues:  OWL-301
 * @param:           c *middleware.Context
 * @return:          void
 * @author:          Don Hsieh
 * @since:           01/27/2015
 * @last modified:   01/27/2015
 * @called by:       r.Get("/boss/cpu/:host", reqSignedIn, Cpu)
 *                    in grafana/pkg/api/api.go
 */
func Cpu(c *middleware.Context) {
	host := c.Params(":host")
	url := getUrl("cpu", host)
	c.Redirect(setting.AppSubUrl + url)
}

/**
 * @function name:   func Net(c *middleware.Context)
 * @description:     This function redirects URL "/boss/net/:host" to net chart.
 * @related issues:  OWL-301
 * @param:           c *middleware.Context
 * @return:          void
 * @author:          Don Hsieh
 * @since:           01/27/2015
 * @last modified:   01/27/2015
 * @called by:       r.Get("/boss/net/:host", reqSignedIn, Net)
 *                    in grafana/pkg/api/api.go
 */
func Net(c *middleware.Context) {
	host := c.Params(":host")
	url := getUrl("net", host)
	c.Redirect(setting.AppSubUrl + url)
}

/**
 * @function name:   func Overview(c *middleware.Context)
 * @description:     This function redirects URL "/boss/overview/:host"
 *                    to "Overview" dashboard.
 * @related issues:  OWL-301
 * @param:           c *middleware.Context
 * @return:          void
 * @author:          Don Hsieh
 * @since:           01/27/2015
 * @last modified:   01/27/2015
 * @called by:       r.Get("/boss/overview/:host", reqSignedIn, Overview)
 *                    in grafana/pkg/api/api.go
 */
func Overview(c *middleware.Context) {
	host := c.Params(":host")
	url := "/dashboard/db/overview?host=" + host
	c.Redirect(setting.AppSubUrl + url)
}

/**
 * @function name:   func getUrl(metric string, host string) string
 * @description:     This function returns destination URL.
 * @related issues:  OWL-301
 * @param:           metric string
 * @param:           host string
 * @return:          url string
 * @author:          Don Hsieh
 * @since:           01/27/2015
 * @last modified:   01/27/2015
 * @called by:       func Cpu(c *middleware.Context)
 *                   func Net(c *middleware.Context)
 *                    in grafana/pkg/api/chart.go
 */
func getUrl(metric string, host string) string {
	panelId := ""
	if metric == "cpu" {
		panelId = "1"
	} else if metric == "net" {
		panelId = "2"
	}
	dashboardName := "overview"
	url := "/dashboard-solo/db/" + dashboardName + "?panelId=" + panelId
	url += "&fullscreen&from=now-3d&to=now&var-host=" + host
	return url
}
