package checker

import (
	"context"
	"net"
	"time"
)

// TCPChecker revisa si un puerto específico está abierto, por ejemplo un NAS
// por su puerto de SMB (445) o FTP (21), o cualquier servicio de red que
// escuche en un puerto TCP conocido.
type TCPChecker struct {
	ServiceName string
	Address     string // ej. "192.168.1.10:445"
	Timeout     time.Duration
}

func NewTCPChecker(name, address string, timeout time.Duration) *TCPChecker {
	return &TCPChecker{ServiceName: name, Address: address, Timeout: timeout}
}

func (c *TCPChecker) Name() string { return c.ServiceName }

func (c *TCPChecker) Check(ctx context.Context) Result {
	start := time.Now()
	dialer := net.Dialer{Timeout: c.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.Address)
	latency := time.Since(start)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return Result{ServiceName: c.ServiceName, Status: StatusUnknown, Message: "timeout", Latency: latency, CheckedAt: time.Now()}
		}
		// Ej: "connection refused" -> el host respondió pero el puerto no está abierto.
		return Result{ServiceName: c.ServiceName, Status: StatusError, Message: err.Error(), Latency: latency, CheckedAt: time.Now()}
	}
	_ = conn.Close()
	return Result{ServiceName: c.ServiceName, Status: StatusOK, Message: "puerto abierto", Latency: latency, CheckedAt: time.Now()}
}
