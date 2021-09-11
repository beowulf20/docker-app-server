package framework_rest

import (
	app_registry "github.com/beowulf20/docker-delta-update-server/framework/app-registry"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

func NewRestServer(reg *app_registry.AppRegistry, cli *client.Client) error {
	r := gin.Default()
	r.GET("/reg/apps/all", appRegListAll(reg))
	r.GET("/reg/app/:id", appParseApp(reg, cli))
	r.POST("/reg/app/:id/stop", regStopApp(reg, cli))
	r.POST("/reg/app/:id/start", regStartApp(reg, cli))
	r.POST("/reg/app/:id/update", regUpdateApp(reg, cli))
	r.POST("/reg/app/new", regNewApp(reg, cli))
	return r.Run()
}
