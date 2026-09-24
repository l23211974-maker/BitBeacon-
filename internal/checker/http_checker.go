package checker

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// HTTPChecker revisa un endpoint HTTP/HTTPS (healthcheck) haciendo un GET
// y evaluando el código de respuesta.
type HTTPChecker struct {
	ServiceName   string
	URL           string
	Timeout       time.Duration
	MinStatusCode int // por defecto 200
	MaxStatusCode int // por defecto 399
}

func NewHTTPChecker(name, url string, timeout time.Duration) *HTTPChecker {
	return &HTTPChecker{
		ServiceName:   name,
		URL:           url,
		Timeout:       timeout,
		MinStatusCode: 200,
		MaxStatusCode: 399,
	}
}

func (c *HTTPChecker) Name() string { return c.ServiceName }

func (c *HTTPChecker) Check(ctx context.Context) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return Result{ServiceName: c.ServiceName, Status: StatusError, Message: err.Error(), CheckedAt: time.Now()}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		// Distinguimos "no sabemos qué pasó" (timeout -> UNKNOWN) de
		// "sabemos que falló" (conexión rechazada, DNS, etc. -> ERROR).
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return Result{ServiceName: c.ServiceName, Status: StatusUnknown, Message: "timeout", Latency: latency, CheckedAt: time.Now()}
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return Result{ServiceName: c.ServiceName, Status: StatusUnknown, Message: "timeout", Latency: latency, CheckedAt: time.Now()}
		}
		return Result{ServiceName: c.ServiceName, Status: StatusError, Message: err.Error(), Latency: latency, CheckedAt: time.Now()}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= c.MinStatusCode && resp.StatusCode <= c.MaxStatusCode {
		return Result{ServiceName: c.ServiceName, Status: StatusOK, Message: resp.Status, Latency: latency, CheckedAt: time.Now()}
	}
	return Result{ServiceName: c.ServiceName, Status: StatusError, Message: resp.Status, Latency: latency, CheckedAt: time.Now()}
}
