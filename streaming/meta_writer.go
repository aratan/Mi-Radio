// Fichero: streaming/meta_writer.go

package streaming

import (
	"fmt"
	"io"
)

// MetaWriter es un 'envoltorio' sobre un io.Writer (nuestra conexión de red)
// que intercepta los datos para inyectar metadatos ICY a intervalos regulares.
type MetaWriter struct {
	dest         io.Writer
	interval     int
	bytesWritten int
	currentTitle string
}

// NewMetaWriter crea una nueva instancia de nuestro escritor inteligente.
func NewMetaWriter(dest io.Writer, interval int) *MetaWriter {
	return &MetaWriter{
		dest:     dest,
		interval: interval,
	}
}

// SetTitle actualiza el título que se enviará en el próximo paquete de metadatos.
func (m *MetaWriter) SetTitle(title string) {
	m.currentTitle = title
}

// Write es el corazón de nuestro copiador. Cumple con la interfaz io.Writer.
func (m *MetaWriter) Write(p []byte) (n int, err error) {
	for len(p) > 0 {
		spaceLeft := m.interval - m.bytesWritten
		if len(p) < spaceLeft {
			written, err := m.dest.Write(p)
			if err != nil {
				return n + written, err
			}
			m.bytesWritten += written
			n += written
			return n, nil
		}

		written, err := m.dest.Write(p[:spaceLeft])
		if err != nil {
			return n + written, err
		}
		m.bytesWritten += written
		n += written
		p = p[spaceLeft:]

		// ¡Hemos alcanzado el intervalo! Es hora de enviar los metadatos.
		// AHORA LLAMA A LA FUNCIÓN PÚBLICA (CON 'S' MAYÚSCULA)
		if err := m.SendMetadata(m.dest, m.currentTitle); err != nil {
			return n, err
		}
		
		m.bytesWritten = 0
	}
	return n, nil
}

// SendMetadata construye y envía un paquete de metadatos.
// ES PÚBLICA (CON 'S' MAYÚSCULA) para que pueda ser llamada desde otros ficheros.
func (m *MetaWriter) SendMetadata(dest io.Writer, title string) error {
	metadataText := fmt.Sprintf("StreamTitle='%s';", title)
	var data []byte
	if title == "" {
		data = []byte{0}
	} else {
		padding := 16 - (len(metadataText) % 16)
		finalLength := len(metadataText) + padding
		lengthByte := byte(finalLength / 16)
		data = make([]byte, 1+finalLength)
		data[0] = lengthByte
		copy(data[1:], metadataText)
	}
	_, err := dest.Write(data)
	return err
}
func (m *MetaWriter) UpdateTitleNow(title string) error {
	m.SetTitle(title)
	// Enviamos un paquete de metadatos AHORA, sin esperar al siguiente intervalo.
	// Esto resetea el contador de bytes a cero en el proceso.
	err := m.SendMetadata(m.dest, title)
	if err == nil {
		m.bytesWritten = 0 // Reseteamos el contador tras el envío manual
	}
	return err
}