package compose

import (
	"errors"
	"os"

	"github.com/compose-spec/compose-go/loader"
	compose "github.com/compose-spec/compose-go/types"
)

func LoadDockerCompose(data []byte, projectName string) (*compose.Project, error) {
	if len(projectName) == 0 {
		return LoadDockerComposeNoName(data)
	}

	config, err := loader.ParseYAML(data)
	if err != nil {
		return nil, err
	}

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

func LoadDockerComposeNoName(data []byte) (*compose.Project, error) {
	config, err := loader.ParseYAML(data)
	if err != nil {
		return nil, err
	}

	projectName, _ := config["project_name"].(string)
	if projectName == "" {
		return nil, errors.New("no project name provided")
	}
	delete(config, "project_name")

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var files []compose.ConfigFile
	files = append(files, compose.ConfigFile{Config: config})
	c, e := loader.Load(compose.ConfigDetails{
		WorkingDir:  wd,
		ConfigFiles: files,
	}, withProjectName(projectName))
	return c, e
}

func withProjectName(name string) func(*loader.Options) {
	return func(lOpts *loader.Options) {
		lOpts.Name = name
	}
}
