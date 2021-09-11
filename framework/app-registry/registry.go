package app_registry

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AppRegistry struct {
	db *gorm.DB
}

func NewAppRegistry() (*AppRegistry, error) {
	// db, err := gorm.Open(sqlite.Open("app_reg.db"), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&App{})
	if err != nil {
		return nil, err
	}

	return &AppRegistry{
		db: db,
	}, nil
}

func (reg *AppRegistry) ListApps() ([]App, error) {
	apps := []App{}
	result := reg.db.Find(&apps)
	if result.Error != nil {
		return nil, result.Error
	}
	return apps, nil
}

func (reg *AppRegistry) AddApp(app *App) error {
	return reg.db.Create(app).Error
}

func (reg *AppRegistry) RemoveAppByID(id uint) error {
	return reg.db.Where("ID = ?", id).Delete(&App{}).Error
}

func (reg *AppRegistry) GetAppByID(id uint) (*App, error) {
	app := new(App)
	result := reg.db.Where("ID LIKE ?", id).First(app)
	if result.Error != nil {
		return nil, result.Error
	}
	return app, nil
}
