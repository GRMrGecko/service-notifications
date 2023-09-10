package main

import (
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"github.com/kkyr/fig"
)

// Configurations relating to HTTP server.
type HTTPConfig struct {
	BindAddr string `fig:"bind_addr"`
	Port     uint   `fig:"port"`
	Debug    bool   `fig:"debug"`
	APIKey   string `fig:"api_key"`
}

// Configurations relating to database.
type DBConfig struct {
	Type       string `fig:"type"` // Review documentation at http://gorm.io/docs/connecting_to_the_database.html
	Connection string `fig:"connection"`
	Debug      bool   `fig:"debug"`
}

// Configurations relating to Planning Center API/Sync.
type PlanningCenterConfig struct {
	AppID          string   `fig:"app_id"`
	Secret         string   `fig:"secret"`
	ServiceTypeIDs []uint64 `fig:"service_type_ids"` // Filter to service type IDs listed.
}

// Configurations relating to Slack API/channel creation.
type SlackConfig struct {
	CreateChannelsAhead time.Duration `fig:"create_channels_ahead"` // Amount of time of future services to create channels head for. Defaults to 8 days head.
	APIToken            string        `fig:"api_token"`
	AdminID             string        `fig:"admin_id"` // Slack user that administers this app.
}

// Configuration Structure.
type Config struct {
	HTTP           HTTPConfig           `fig:"http"`
	DB             DBConfig             `fig:"database"`
	PlanningCenter PlanningCenterConfig `fig:"planning_center"`
	Slack          SlackConfig          `fig:"slack"`
}

// Load the configuration.
func (a *App) ReadConfig() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Configuration paths.
	localConfig, _ := filepath.Abs("./config.yaml")
	homeDirConfig := usr.HomeDir + "/.config/service-notifications/config.yaml"
	etcConfig := "/etc/service-notifications/config.yaml"

	// Determine which configuration to use.
	var configFile string
	if _, err := os.Stat(app.flags.ConfigPath); err == nil && app.flags.ConfigPath != "" {
		configFile = app.flags.ConfigPath
	} else if _, err := os.Stat(localConfig); err == nil {
		configFile = localConfig
	} else if _, err := os.Stat(homeDirConfig); err == nil {
		configFile = homeDirConfig
	} else if _, err := os.Stat(etcConfig); err == nil {
		configFile = etcConfig
	} else {
		log.Fatal("Unable to find a configuration file.")
	}

	// Load the configuration file.
	config := &Config{
		HTTP: HTTPConfig{
			BindAddr: "",
			Port:     34935,
			Debug:    true,
		},
		DB: DBConfig{
			Type:       "sqlite3",
			Connection: "service-notifications.db",
		},
		Slack: SlackConfig{
			CreateChannelsAhead: time.Hour * 24 * 8,
		},
	}

	// Load configuration.
	filePath, fileName := path.Split(configFile)
	err = fig.Load(config,
		fig.File(fileName),
		fig.Dirs(filePath),
	)
	if err != nil {
		log.Printf("Error parsing configuration: %s\n", err)
		return
	}

	// Override flags.
	if app.flags.HTTPBind != "" {
		config.HTTP.BindAddr = app.flags.HTTPBind
	}
	if app.flags.HTTPPort != 0 {
		config.HTTP.Port = app.flags.HTTPPort
	}

	// Set global config structure.
	app.config = config
}
