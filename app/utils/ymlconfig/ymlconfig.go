package ymlconfig

import (
	"log"
	"sync"
	"time"

	"gin-fast/app/global/app"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var lastChangeTime time.Time

func init() {
	lastChangeTime = time.Now()
}

func CreateYamlFactory(path string, fileName ...string) app.YmlConfigInterf {
	yamlConfig := viper.New()
	yamlConfig.AddConfigPath(path)
	if len(fileName) == 0 {
		yamlConfig.SetConfigName("config")
	} else {
		yamlConfig.SetConfigName(fileName[0])
	}
	yamlConfig.SetConfigType("yml")

	if err := yamlConfig.ReadInConfig(); err != nil {
		log.Fatal("ReadInConfig err: " + err.Error())
	}

	return &ymlConfig{
		viper: yamlConfig,
		mu:    new(sync.RWMutex),
	}
}

type ymlConfig struct {
	viper *viper.Viper
	mu    *sync.RWMutex
}

func (y *ymlConfig) ConfigFileChangeListen(fns ...func()) {
	y.viper.OnConfigChange(func(changeEvent fsnotify.Event) {
		if time.Since(lastChangeTime).Seconds() >= 1 {
			if changeEvent.Op.String() == "WRITE" {
				y.mu.Lock()
				if err := y.viper.ReadInConfig(); err != nil {
					log.Printf("reload config failed: %v", err)
				} else {
					log.Println("config reloaded successfully")
				}
				y.mu.Unlock()
				for _, f := range fns {
					f()
				}
				lastChangeTime = time.Now()
			}
		}
	})
	y.viper.WatchConfig()
}

func (y *ymlConfig) Get(keyName string) interface{} {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.Get(keyName)
}

func (y *ymlConfig) GetString(keyName string) string {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetString(keyName)
}

func (y *ymlConfig) GetBool(keyName string) bool {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetBool(keyName)
}

func (y *ymlConfig) GetInt(keyName string) int {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetInt(keyName)
}

func (y *ymlConfig) GetInt32(keyName string) int32 {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetInt32(keyName)
}

func (y *ymlConfig) GetInt64(keyName string) int64 {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetInt64(keyName)
}

func (y *ymlConfig) GetFloat64(keyName string) float64 {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetFloat64(keyName)
}

func (y *ymlConfig) GetDuration(keyName string) time.Duration {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetDuration(keyName)
}

func (y *ymlConfig) GetStringSlice(keyName string) []string {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetStringSlice(keyName)
}

func (y *ymlConfig) GetUintSlice(keyName string) []uint {
	y.mu.RLock()
	defer y.mu.RUnlock()

	if value := y.viper.Get(keyName); value != nil {
		if uintSlice, ok := value.([]uint); ok {
			return uintSlice
		}
	}

	intSlice := y.viper.GetIntSlice(keyName)
	if len(intSlice) == 0 {
		return []uint{}
	}

	uintSlice := make([]uint, len(intSlice))
	for i, v := range intSlice {
		if v < 0 {
			uintSlice[i] = 0
		} else {
			uintSlice[i] = uint(v)
		}
	}

	return uintSlice
}

func (y *ymlConfig) GetStringMap(keyName string) map[string]interface{} {
	y.mu.RLock()
	defer y.mu.RUnlock()
	return y.viper.GetStringMap(keyName)
}

func (y *ymlConfig) Set(keyName string, value interface{}) {
	y.mu.Lock()
	defer y.mu.Unlock()
	y.viper.Set(keyName, value)
}

func (y *ymlConfig) SaveConfig() error {
	y.mu.Lock()
	defer y.mu.Unlock()

	currentSettings := y.viper.AllSettings()
	configFile := y.viper.ConfigFileUsed()

	// Write through a fresh viper instance so removed map keys do not
	// survive via merge behavior from the original loaded config.
	fresh := viper.New()
	fresh.SetConfigFile(configFile)
	fresh.SetConfigType("yml")
	for key, value := range currentSettings {
		fresh.Set(key, value)
	}

	if err := fresh.WriteConfigAs(configFile); err != nil {
		return err
	}

	return y.viper.ReadInConfig()
}
