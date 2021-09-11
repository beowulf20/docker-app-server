package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	app_registry "github.com/beowulf20/docker-delta-update-server/framework/app-registry"
	compose "github.com/beowulf20/docker-delta-update-server/framework/compose"
	ctypes "github.com/compose-spec/compose-go/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type ContainerStatus int

func (s ContainerStatus) ToString() string {
	switch s {
	case ContainerNotRunning:
		return "not running"
	case ContainerRunning:
		return "running"
	case ContainerNotCreated:
		return "not created"
	default:
		return "unknown"
	}
}

const (
	ContainerNotRunning ContainerStatus = iota
	ContainerRunning    ContainerStatus = iota
	ContainerNotCreated ContainerStatus = iota
)

type AppContainerLink struct {
	Service   ctypes.ServiceConfig
	Container *types.Container
	Status    ContainerStatus
}

var hasher = sha256.New()

func (link *AppContainerLink) CalculateServiceHash() (string, error) {
	payload, err := json.Marshal(link.Service)
	if err != nil {
		return "", err
	}
	defer hasher.Reset()
	reader := strings.NewReader(string(payload))
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func AssociateContainerApp(app app_registry.App, cli *client.Client) ([]AppContainerLink, error) {
	project, err := compose.LoadDockerCompose([]byte(app.ComposeScript), app.Name)
	if err != nil {
		return nil, err
	}
	var conts []AppContainerLink
	for _, service := range project.AllServices() {
		cont, err := getContainerForAppService(fmt.Sprintf("/%s_%s", project.Name, service.Name), cli)
		if err != nil {
			return nil, err
		}
		status := ContainerNotCreated
		if cont != nil {
			if cont.State != "running" {
				status = ContainerNotRunning
			} else {
				status = ContainerRunning
			}
		}
		conts = append(conts, AppContainerLink{
			Service:   service,
			Container: cont,
			Status:    status,
		})
	}

	return conts, nil
}

func getContainerForAppService(serviceName string, cli *client.Client) (*types.Container, error) {
	conts, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: serviceName,
		}),
	})
	if err != nil {
		return nil, err
	}
	if len(conts) == 0 {
		return nil, nil
	}
	return &conts[0], nil
}
