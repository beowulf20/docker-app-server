package app_registry

import (
	"encoding/json"
	"errors"

	app_compose "github.com/beowulf20/docker-delta-update-server/framework/compose"
	"gorm.io/gorm"
)

var ErrAppNotValid = errors.New("app is not valid")

type App struct {
	Name          string `gorm:"unique;not null" json:"name"`
	ComposeScript string `json:"script"`
	ComposeHash   string `gorm:"unique;not null" json:"hash"`
	gorm.Model
}

func (app *App) Validate() error {
	if len(app.Name) == 0 {
		return ErrAppNotValid
	}

	if err := validateStringSpecialCharacters(app.Name); err != nil {
		return err
	}

	if len(app.ComposeScript) == 0 {
		return ErrAppNotValid
	}
	if len(app.ComposeHash) == 0 {
		return ErrAppNotValid
	}
	return nil
}

func NewApp(name string, script string) (*App, error) {
	project, err := app_compose.LoadDockerCompose([]byte(script), name)
	if err != nil {
		return nil, err
	}
	scriptPayload, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}

	hash, err := calcHash(string(scriptPayload))
	if err != nil {
		return nil, err
	}
	app := &App{
		Name:          project.Name,
		ComposeScript: script,
		ComposeHash:   hash,
	}
	err = app.Validate()
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (reg *AppRegistry) UpdateApp(id uint, newScript string) error {
	return reg.db.Model(&App{}).Where("ID = ?", id).Update("compose_script", newScript).Error
}
