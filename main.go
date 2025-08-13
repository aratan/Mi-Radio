package main

import (
	"fmt"
	"log"
	"mi-radio/config"   // <-- IMPORT CORREGIDO
	"mi-radio/playlist" // <-- IMPORT CORREGIDO
	"mi-radio/streaming"
)

func main() {
	cfg, err := config.Load("config.json") // <-- USO CORREGIDO
	if err != nil {
		log.Fatalf("Error fatal cargando config.json: %v", err)
	}

	streamer := &streaming.ShoutcastV1Streamer{Config: cfg}

	var appUI *TUI
	appUI = NewTUI(func() {
		p := &playlist.Playlist{MusicDir: cfg.MusicDir} // <-- USO CORREGIDO
		go func() {
			err := streamer.Stream(p, appUI.UpdateStatus)
			if err != nil {
				appUI.UpdateStatus(fmt.Sprintf("[red]ERROR FATAL DEL STREAMER: %v", err))
			}
		}()
	})

	log.Println("Iniciando aplicación de radio...")
	if err := appUI.Run(); err != nil {
		log.Fatalf("Error al ejecutar la TUI: %v", err)
	}
}