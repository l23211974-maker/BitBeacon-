package checker

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// DockerChecker revisa el estado de un contenedor usando el CLI de Docker,
// equivalente a correr:
//
//	docker inspect --format '{{.State.Status}}' <contenedor>
//
// Se eligió el CLI en vez del SDK/Docker Engine API para mantener el
// proyecto simple (sin dependencias extra ni manejo de sockets Unix).
// Si más adelante hace falta algo más fino (logs, stats, eventos en vivo),
// se puede reemplazar este archivo por una versión que use la Docker API
// sin tocar el resto del programa, porque ambos implementan la misma
// interfaz Checker.
type DockerChecker struct {
	ServiceName   string
	ContainerName string
	Timeout       time.Duration
}

func NewDockerChecker(name, containerName string, timeout time.Duration) *DockerChecker {
	return &DockerChecker{ServiceName: name, ContainerName: containerName, Timeout: timeout}
}

func (c *DockerChecker) Name() string { return c.ServiceName }

func (c *DockerChecker) Check(ctx context.Context) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Status}}", c.ContainerName)
	out, err := cmd.Output()
	latency := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{ServiceName: c.ServiceName, Status: StatusUnknown, Message: "timeout consultando docker", Latency: latency, CheckedAt: time.Now()}
	}
	if err != nil {
		// El contenedor no existe, el daemon de Docker no está corriendo,
		// o el usuario no tiene permisos para hablar con el socket de Docker.
		return Result{ServiceName: c.ServiceName, Status: StatusError, Message: "docker inspect falló: " + err.Error(), Latency: latency, CheckedAt: time.Now()}
	}

	state := strings.TrimSpace(string(out))
	if state == "running" {
		return Result{ServiceName: c.ServiceName, Status: StatusOK, Message: state, Latency: latency, CheckedAt: time.Now()}
	}
	// Otros valores posibles: "exited", "paused", "restarting", "dead", "created".
	return Result{ServiceName: c.ServiceName, Status: StatusError, Message: state, Latency: latency, CheckedAt: time.Now()}
}
