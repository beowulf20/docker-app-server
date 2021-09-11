package framework_rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	app_registry "github.com/beowulf20/docker-delta-update-server/framework/app-registry"
	app_compose "github.com/beowulf20/docker-delta-update-server/framework/compose"
	utils "github.com/beowulf20/docker-delta-update-server/framework/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

func appRegListAll(reg *app_registry.AppRegistry) func(c *gin.Context) {
	return func(c *gin.Context) {
		apps, err := reg.ListApps()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var data []gin.H
		for _, app := range apps {
			data = append(data, gin.H{
				"id":       app.ID,
				"name":     app.Name,
				"createAt": app.CreatedAt,
				"hash":     app.ComposeHash,
			})
		}
		c.JSON(http.StatusOK, data)
	}
}

func appParseApp(reg *app_registry.AppRegistry, cli *client.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		app, err := reg.GetAppByID(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		conts, err := utils.AssociateContainerApp(*app, cli)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var containersMap []map[string]interface{}
		for _, cont := range conts {
			containersMap = append(containersMap, map[string]interface{}{
				"name":   cont.Service.Name,
				"status": cont.Status.ToString(),
				"image":  cont.Service.Image,
				"volumes": func() []string {
					volumes := []string{}
					for _, volume := range cont.Service.Volumes {
						volumes = append(volumes, fmt.Sprintf("%s:%s", volume.Source, volume.Target))
					}
					return volumes
				}(),
			})
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"id":         app.ID,
			"name":       app.Name,
			"containers": containersMap,
		})

	}
}

func regStopApp(reg *app_registry.AppRegistry, cli *client.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		app, err := reg.GetAppByID(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		conts, err := utils.AssociateContainerApp(*app, cli)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		for _, cont := range conts {
			if cont.Status == utils.ContainerRunning {
				err = cli.ContainerStop(c, cont.Container.ID, nil)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": err.Error(),
					})
					return
				}
			}
		}

	}
}

func regStartApp(reg *app_registry.AppRegistry, cli *client.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		app, err := reg.GetAppByID(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		conts, err := utils.AssociateContainerApp(*app, cli)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		for _, cont := range conts {
			if cont.Status == utils.ContainerRunning {
				continue
			}
			if cont.Status == utils.ContainerNotCreated {
				newContBody, err := cli.ContainerCreate(c, &container.Config{
					Image: cont.Service.Image,
				}, nil, nil, nil, app.Name+"_"+cont.Service.Name)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": err.Error(),
					})
					return
				}
				err = cli.ContainerStart(c, newContBody.ID, types.ContainerStartOptions{})
				if err != nil {
					c.JSON(401, gin.H{
						"error": err.Error(),
					})
					return
				}
				continue
			}

			err = cli.ContainerStart(c, cont.Container.ID, types.ContainerStartOptions{})
			if err != nil {
				c.JSON(401, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

	}
}

func regUpdateApp(reg *app_registry.AppRegistry, cli *client.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		oldApp, err := reg.GetAppByID(uint(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		composeData, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		newApp, err := app_registry.NewApp(oldApp.Name, string(composeData))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		oldProject, err := app_compose.LoadDockerCompose([]byte(oldApp.ComposeScript), oldApp.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		newProject, err := app_compose.LoadDockerCompose([]byte(newApp.ComposeScript), oldApp.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = reg.UpdateApp(uint(id), string(composeData))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		payload1, _ := json.Marshal(newProject)
		payload2, _ := json.Marshal(oldProject)
		oldHash, _ := app_registry.CalculateHash(payload1)
		newHash, _ := app_registry.CalculateHash(payload2)

		willUpdate := oldHash != newHash

		if willUpdate {
			err = reg.UpdateApp(uint(id), string(composeData))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}
			oldConts, err := utils.AssociateContainerApp(*oldApp, cli)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}
			newConts, err := utils.AssociateContainerApp(*newApp, cli)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			hashes := make(map[string]map[string]string)
			hashes["old"] = make(map[string]string)
			hashes["new"] = make(map[string]string)
			for _, cont := range oldConts {
				hash, _ := cont.CalculateServiceHash()
				hashes["old"][cont.Service.Name] = hash
			}
			for _, cont := range newConts {
				hash, _ := cont.CalculateServiceHash()
				hashes["new"][cont.Service.Name] = hash
			}

			// CHECK BY HASH WHATS CHANGED
			for newContName, newContHash := range hashes["new"] {
				for oldContName, oldContHash := range hashes["old"] {
					if newContHash == oldContHash {
						if oldContName != newContName {
							fmt.Print("\n[HASH] OLD NAME => NEW NAME\n")
							fmt.Printf("[%s] %s => %s\n", newContHash, oldContName, newContName)
						}
					}
				}
			}
			// CHECK BY NAME WHATS CHANGED
			for newContName, newContHash := range hashes["new"] {
				if oldContHash, ok := hashes["old"][newContName]; ok {
					if oldContHash != newContHash {
						fmt.Printf("\n[NAME] OLD HASH => NEW HASH\n")
						fmt.Printf("[%s] %s => %s\n", newContName, oldContHash, newContHash)
					}
				}
			}

			for _, cont := range oldConts {
				if cont.Status == utils.ContainerNotCreated {
					continue
				}
				if cont.Status == utils.ContainerRunning {
					err = cli.ContainerStop(c, cont.Container.ID, nil)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": err.Error(),
						})
						return
					}
				}
				err = cli.ContainerRemove(c, cont.Container.ID, types.ContainerRemoveOptions{})
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": err.Error(),
					})
					return
				}
			}

			for _, cont := range newConts {
				if cont.Status == utils.ContainerRunning {
					continue
				}
				if cont.Status == utils.ContainerNotCreated {
					newContBody, err := cli.ContainerCreate(c, &container.Config{
						Image: cont.Service.Image,
					}, nil, nil, nil, newApp.Name+"_"+cont.Service.Name)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": err.Error(),
						})
						return
					}
					err = cli.ContainerStart(c, newContBody.ID, types.ContainerStartOptions{})
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": err.Error(),
						})
						return
					}
				}
				if cont.Status == utils.ContainerNotRunning {
					err = cli.ContainerStart(c, cont.Container.ID, types.ContainerStartOptions{})
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": err.Error(),
						})
						return
					}
				}
			}
		}

		c.JSON(200, gin.H{
			"hash": map[string]string{
				"old": oldHash,
				"new": newHash,
			},
			"didUpdate": willUpdate,
		})
	}
}

func regNewApp(reg *app_registry.AppRegistry, cli *client.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		composeData, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		app, err := app_registry.NewApp("", string(composeData))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = reg.AddApp(app)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.String(http.StatusOK, "OK")
	}
}
