package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TUI gestiona todos los componentes de la interfaz de usuario.
type TUI struct {
	app        *tview.Application
	statusText *tview.TextView
	onConnect  func() // Callback para cuando se pulsa "Conectar"
}

// NewTUI crea una nueva instancia de la TUI.
func NewTUI(onConnectCallback func()) *TUI {
	t := &TUI{
		app:       tview.NewApplication(),
		onConnect: onConnectCallback,
	}

	dashboardPage := t.createDashboardPage()

	t.app.SetRoot(dashboardPage, true).EnableMouse(true)
	return t
}

// Run inicia el bucle principal de la aplicación TUI.
func (t *TUI) Run() error {
	return t.app.Run()
}

// UpdateStatus es un método seguro para actualizar el texto del dashboard desde cualquier goroutine.
func (t *TUI) UpdateStatus(message string) {
	t.app.QueueUpdateDraw(func() {
		fmt.Fprintln(t.statusText, message)
	})
}

func (t *TUI) createDashboardPage() *tview.Frame {
	t.statusText = tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			t.app.Draw()
		})

	// Iniciar la transmisión inmediatamente al arrancar
	// El botón de conectar ya no es necesario en esta versión simplificada
	go t.onConnect()

	dashboard := tview.NewFrame(t.statusText).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("Radio Dashboard (SOLID)", true, tview.AlignCenter, tcell.ColorWhite).
		AddText("Presiona Ctrl+C para salir", false, tview.AlignCenter, tcell.ColorGreen)

	return dashboard
}