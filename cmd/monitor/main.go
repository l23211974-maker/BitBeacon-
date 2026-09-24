package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"monitor-microbit/internal/checker"
	"monitor-microbit/internal/config"
	"monitor-microbit/internal/serialcomm"
)

func main() {
	cfgPath := "configs/config.json"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("error cargando configuración: %v", err)
	}

	interval, err := cfg.PollIntervalDuration()
	if err != nil {
		log.Fatalf("poll_interval inválido: %v", err)
	}

	checkers := buildCheckers(cfg)
	if len(checkers) == 0 {
		log.Fatal("no hay servicios configurados en el archivo de configuración")
	}

	serialMgr := serialcomm.NewManager(cfg.SerialPort, cfg.BaudRate, log.Default())
	defer serialMgr.Close()

	log.Printf("iniciando monitoreo de %d servicio(s) cada %s (puerto serial: %s)",
		len(checkers), interval, cfg.SerialPort)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Corremos un ciclo inmediatamente al arrancar, y luego uno por cada tick.
	runCycle(checkers, serialMgr)
	for range ticker.C {
		runCycle(checkers, serialMgr)
	}
}

// buildCheckers arma la lista de Checkers a partir de la config: un Checker
// concreto por cada entrada en cfg.Services. Este es el único lugar del
// programa que sabe mapear "type" (string) -> struct concreto; agregar un
// quinto tipo de servicio es agregar un case acá y su archivo en internal/checker.
func buildCheckers(cfg *config.Config) []checker.Checker {
	var checkers []checker.Checker
	for _, s := range cfg.Services {
		timeout := time.Duration(s.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 3 * time.Second
		}
		switch s.Type {
		case "http":
			checkers = append(checkers, checker.NewHTTPChecker(s.Name, s.Target, timeout))
		case "docker":
			checkers = append(checkers, checker.NewDockerChecker(s.Name, s.Target, timeout))
		case "tcp":
			checkers = append(checkers, checker.NewTCPChecker(s.Name, s.Target, timeout))
		case "ping":
			checkers = append(checkers, checker.NewPingChecker(s.Name, s.Target, timeout))
		default:
			log.Printf("tipo de servicio desconocido, se ignora: %q (%s)", s.Type, s.Name)
		}
	}
	return checkers
}

// runCycle ejecuta todos los checkers EN PARALELO (una goroutine por servicio,
// así un servicio lento no atrasa a los demás), loggea cada resultado,
// calcula el estado agregado con checker.Overall y lo envía al micro:bit.
func runCycle(checkers []checker.Checker, serialMgr *serialcomm.Manager) {
	results := make([]checker.Result, len(checkers))
	var wg sync.WaitGroup

	for i, c := range checkers {
		wg.Add(1)
		go func(i int, c checker.Checker) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			results[i] = c.Check(ctx)
		}(i, c)
	}
	wg.Wait()

	for _, r := range results {
		log.Printf("[%s] %s (%s) - %v", r.ServiceName, r.Status, r.Message, r.Latency)
	}

	overall := checker.Overall(results)
	code := statusToCode(overall)

	if err := serialMgr.SendStatus(code); err != nil {
		log.Printf("[serial] no se pudo enviar estado al micro:bit: %v", err)
	} else {
		log.Printf("estado general: %s -> enviado '%c' al micro:bit", overall, code)
	}
}

// statusToCode traduce el Status a un solo byte ASCII: el protocolo que
// entiende el micro:bit del lado de MicroPython.
func statusToCode(s checker.Status) byte {
	switch s {
	case checker.StatusOK:
		return 'O'
	case checker.StatusError:
		return 'E'
	default: // checker.StatusUnknown
		return 'U'
	}
}
