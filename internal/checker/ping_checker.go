package checker

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

// PingChecker hace un ping ICMP real usando el comando `ping` del sistema
// operativo (en vez de armar paquetes ICMP a mano en Go). La razón es que
// enviar ICMP crudo requiere privilegios de administrador/root en la mayoría
// de sistemas operativos, y para este proyecto no vale la pena esa
// complejidad ni pedirle al usuario que corra el programa como root.
type PingChecker struct {
	ServiceName string
	Host        string
	Timeout     time.Duration
}

func NewPingChecker(name, host string, timeout time.Duration) *PingChecker {
	return &PingChecker{ServiceName: name, Host: host, Timeout: timeout}
}

func (c *PingChecker) Name() string { return c.ServiceName }

func (c *PingChecker) Check(ctx context.Context) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		ms := strconv.Itoa(int(c.Timeout.Milliseconds()))
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", ms, c.Host)
	} else {
		// Linux y macOS: -W espera segundos (Linux) o milisegundos (macOS usa -W en ms también
		// en versiones BSD); para simplificar, si en macOS falla por formato, ajustar a "-t" con reintentos.
		secs := strconv.Itoa(int(c.Timeout.Seconds()))
		if secs == "0" {
			secs = "1"
		}
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", secs, c.Host)
	}

	err := cmd.Run()
	latency := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{ServiceName: c.ServiceName, Status: StatusUnknown, Message: "timeout", Latency: latency, CheckedAt: time.Now()}
	}
	if err != nil {
		return Result{ServiceName: c.ServiceName, Status: StatusError, Message: "sin respuesta al ping", Latency: latency, CheckedAt: time.Now()}
	}
	return Result{ServiceName: c.ServiceName, Status: StatusOK, Message: "responde", Latency: latency, CheckedAt: time.Now()}
}
