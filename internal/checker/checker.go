// Package checker define la abstracción para "revisar el estado de un servicio".
// Cualquier tipo de chequeo (HTTP, Docker, TCP, Ping) implementa la misma interfaz,
// así el resto del programa no necesita saber los detalles de cada uno.
package checker

import (
	"context"
	"time"
)

// Status representa el estado de un servicio monitoreado.
type Status string

const (
	StatusOK      Status = "OK"
	StatusError   Status = "ERROR"
	StatusUnknown Status = "UNKNOWN"
)

// Result es lo que devuelve cada chequeo individual.
type Result struct {
	ServiceName string
	Status      Status
	Message     string        // detalle legible (ej. error, código HTTP, etc.)
	Latency     time.Duration // cuánto tardó el chequeo
	CheckedAt   time.Time
}

// Checker es la interfaz que deben cumplir todos los "chequeadores".
// Check recibe un context.Context para poder cancelar/limitar por timeout
// desde afuera (por ejemplo, desde main.go).
type Checker interface {
	Name() string
	Check(ctx context.Context) Result
}

// Overall calcula el estado agregado de varios resultados, con esta prioridad:
//  1. Si algún servicio está en ERROR, el estado general es ERROR.
//  2. Si no hay errores pero algo está en UNKNOWN (timeout), el general es UNKNOWN.
//  3. Solo si todo respondió bien, el general es OK.
//
// Esta prioridad importa para decidir qué ícono mostrar cuando hay varios
// servicios: preferimos "avisar lo peor" antes que mostrar una carita feliz
// si algo, aunque sea un timeout, no está confirmado como sano.
func Overall(results []Result) Status {
	hasUnknown := false
	for _, r := range results {
		if r.Status == StatusError {
			return StatusError
		}
		if r.Status == StatusUnknown {
			hasUnknown = true
		}
	}
	if hasUnknown {
		return StatusUnknown
	}
	return StatusOK
}
