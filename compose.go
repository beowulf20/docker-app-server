package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	app_registry "github.com/beowulf20/docker-delta-update-server/framework/app-registry"
	"github.com/compose-spec/compose-go/loader"
	compose "github.com/compose-spec/compose-go/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func LoadDockerComposeFile(fileName string, projectName string) (*compose.Project, error) {

	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	config, err := loader.ParseYAML(b)
	if err != nil {
		return nil, err
	}

	wd, err := filepath.Abs(filepath.Dir(fileName))
	if err != nil {
		return nil, err
	}
	var files []compose.ConfigFile
	files = append(files, compose.ConfigFile{Filename: fileName, Config: config})
	return loader.Load(compose.ConfigDetails{
		WorkingDir:  wd,
		ConfigFiles: files,
	}, withProjectName(projectName))
}

func LoadDockerCompose(data []byte, projectName string) (*compose.Project, error) {
	config, err := loader.ParseYAML(data)
	if err != nil {
		return nil, err
	}

	// wd, err := filepath.Abs(os.Getwd())
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var files []compose.ConfigFile
	files = append(files, compose.ConfigFile{Config: config})
	return loader.Load(compose.ConfigDetails{
		WorkingDir:  wd,
		ConfigFiles: files,
	}, withProjectName(projectName))
}

func withProjectName(name string) func(*loader.Options) {
	return func(lOpts *loader.Options) {
		lOpts.Name = name
	}
}

func StartOrCreateContainers(ctx context.Context, app app_registry.App, cli *client.Client) error {
	var args []filters.KeyValuePair

	project, err := LoadDockerCompose([]byte(app.ComposeScript), app.Name)
	if err != nil {
		return err
	}
	for _, service := range project.AllServices() {
		contName := fmt.Sprintf("%s_%s", project.Name, service.Name)
		args = append(args, filters.KeyValuePair{
			Key:   "name",
			Value: contName,
		})
	}

	filterNames := filters.NewArgs(args...)
	/*ALREADY fD CONTAINERS*/
	var runningConts []string
	conts, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filterNames,
		All:     true,
	})
	if err != nil {
		return err
	}
	for _, cont := range conts {
		if cont.State != "running" {
			err = cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
			if err != nil {
				return err
			}
			log.Printf("started %s", cont.Names[0])
		} else {
			log.Printf("'%s' already running", cont.Names[0])
		}
		runningConts = append(runningConts, cont.Names...)
	}
	/*CREATE CONTAINERS*/
	for _, service := range project.AllServices() {
		contName := fmt.Sprintf("/%s_%s", project.Name, service.Name)
		shouldCreate := true
		for _, cont := range runningConts {
			if cont == contName {
				shouldCreate = false
			}
		}
		if shouldCreate {
			_, err = cli.ContainerCreate(context.Background(), &container.Config{
				Image: service.Image,
			}, nil, nil, nil, contName)
			if err != nil {
				return err
			}
			log.Printf("created %s", contName)
		}
	}

	return nil
}

func StopContainers(ctx context.Context, app app_registry.App, cli *client.Client) error {
	var args []filters.KeyValuePair
	project, err := LoadDockerCompose([]byte(app.ComposeScript), app.Name)
	if err != nil {
		return err
	}
	for _, service := range project.AllServices() {
		contName := fmt.Sprintf("%s_%s", project.Name, service.Name)
		args = append(args, filters.KeyValuePair{
			Key:   "name",
			Value: contName,
		})
	}
	filterNames := filters.NewArgs(args...)
	conts, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filterNames,
	})
	if err != nil {
		return err
	}
	for _, cont := range conts {
		err = cli.ContainerStop(ctx, cont.ID, nil)
		if err != nil {
			return err
		}
		log.Printf("stopped %s", cont.Names[0])
	}
	return nil
}
