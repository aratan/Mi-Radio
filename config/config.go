package config // <-- OJO: El paquete ahora es "config"

import (
	"encoding/json"
	"os"
)

// Config almacena toda la configuración de la aplicación.
type Config struct {
	Hostname   string `json:"provider_hostname"`
	Port       int    `json:"provider_port"`
	Password   string `json:"source_password"`
	MountPoint string `json:"mount_point"`
	StreamID   string `json:"stream_id"`
	Username   string `json:"username"`
	MusicDir   string `json:"music_directory"`
}

// Load carga la configuración desde un fichero JSON.
func Load(path string) (Config, error) {
	var config Config
	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}