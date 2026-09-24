// Package serialcomm se encarga exclusivamente de hablar con el micro:bit
// por puerto serial USB. Deliberadamente no sabe nada de "servicios" ni de
// "checkers": solo sabe enviar bytes y reconectar si la conexión se cae.
// Esta separación es la que permite cambiar mañana el medio de transporte
// (por ejemplo, a radio) tocando solo este archivo.
package serialcomm

import (
	"fmt"
	"log"
	"sync"

	"go.bug.st/serial"
)

// Manager mantiene la conexión serial con el micro:bit y expone SendStatus.
// Es seguro usarlo desde una sola goroutine a la vez gracias al mutex interno.
type Manager struct {
	portName string
	baud     int
	mu       sync.Mutex
	port     serial.Port
	logger   *log.Logger
}

func NewManager(portName string, baud int, logger *log.Logger) *Manager {
	if logger == nil {
		logger = log.Default()
	}
	return &Manager{portName: portName, baud: baud, logger: logger}
}

// connect abre el puerto serial. Debe llamarse con m.mu ya tomado.
func (m *Manager) connect() error {
	mode := &serial.Mode{BaudRate: m.baud}
	p, err := serial.Open(m.portName, mode)
	if err != nil {
		return fmt.Errorf("no se pudo abrir %s: %w", m.portName, err)
	}
	m.port = p
	m.logger.Printf("[serial] conectado a %s", m.portName)
	return nil
}

// SendStatus manda un código de un solo carácter + salto de línea, ej: "O\n".
// Protocolo simple: 'O' = OK, 'E' = ERROR, 'U' = UNKNOWN/timeout.
//
// Si no hay conexión abierta (primera vez, o porque se perdió antes),
// intenta reconectar automáticamente antes de escribir. Si la escritura
// falla, se cierra y se descarta el puerto para forzar una reconexión en
// el próximo ciclo de polling — así no hace falta un loop de reintento
// separado, el propio ticker de cmd/monitor/main.go actúa como reintento.
func (m *Manager) SendStatus(code byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.port == nil {
		if err := m.connect(); err != nil {
			return err
		}
	}

	_, err := m.port.Write([]byte{code, '\n'})
	if err != nil {
		m.logger.Printf("[serial] error escribiendo, se cerrará el puerto para reintentar luego: %v", err)
		_ = m.port.Close()
		m.port = nil
		return err
	}
	return nil
}

// Close cierra el puerto serial si está abierto. Se llama al terminar el programa.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.port != nil {
		_ = m.port.Close()
		m.port = nil
	}
}
