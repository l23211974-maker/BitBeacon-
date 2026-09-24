// testserver.go - servidor de prueba MUY simple para probar los checkers
// sin necesitar nada externo (ni Docker, ni un NAS real, ni Python).
// No es parte del monitor en sí: es solo una herramienta de apoyo para
// probar mientras no tienen servicios reales a mano. Usa únicamente la
// librería estándar de Go, así que no hace falta instalar nada más.
//
// Uso (en PowerShell o CMD, parados en la carpeta testing/):
//
//	go run testserver.go
//
// Levanta:
//   - Un endpoint HTTP de salud en http://127.0.0.1:8090/health (responde 200 OK)
//   - Un puerto TCP abierto en 127.0.0.1:9090 (para probar el checker "tcp", ej. NAS)
//
// Apunten su config de pruebas (por ejemplo testing/config.test.json) a estos dos servicios para ver los
// checkers de HTTP y TCP dar estado OK. Para ver ERROR, apunten a un
// puerto donde NO tengan nada corriendo (ej. 127.0.0.1:9999). Para ver
// UNKNOWN, apunten el checker de tipo "ping" a una IP no ruteable de su
// red (ej. algo fuera de su rango de LAN) con un timeout corto.
package main

import (
	"fmt"
	"net"
	"net/http"
)

func main() {
	// --- Parte HTTP: healthcheck simple ---
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	go func() {
		fmt.Println("HTTP de prueba escuchando en http://127.0.0.1:8090/health")
		if err := http.ListenAndServe("127.0.0.1:8090", nil); err != nil {
			fmt.Println("error en servidor HTTP:", err)
		}
	}()

	// --- Parte TCP: puerto abierto simple (simula un NAS/servicio de red) ---
	ln, err := net.Listen("tcp", "127.0.0.1:9090")
	if err != nil {
		fmt.Println("error abriendo puerto TCP:", err)
		return
	}
	fmt.Println("Puerto TCP de prueba abierto en 127.0.0.1:9090")
	fmt.Println("Dejen esta ventana abierta mientras prueban. Ctrl+C para detener.")
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		conn.Close()
	}
}
