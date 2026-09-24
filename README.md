# Monitor de servicios con micro:bit

Programa en Go que revisa el estado de servicios (HTTP, Docker, NAS/TCP, ping)
en intervalos configurables y muestra el resultado en la matriz de LEDs de un
micro:bit, conectado por USB.

## Arquitectura

```
monitor-microbit/
├── .github/
├── README.md
├── go.mod
├── go.sum
├── cmd/
│   └── monitor/
│       └── main.go
├── configs/
│   └── config.json
├── internal/
│   ├── checker/         -> "¿cómo reviso un servicio?"
│   ├── config/           -> "¿cómo leo la configuración?"
│   └── serialcomm/       -> "¿cómo le hablo al micro:bit?"
├── testing/
└── microbit/
    └── main.py           -> firmware MicroPython del micro:bit
```

Los tres paquetes de `internal/` no dependen entre sí: `checker` no sabe que
existe `serialcomm`, y viceversa. `cmd/monitor/main.go` es el único que los conoce a
todos y los conecta. Esto es justamente lo que pide la consigna de
"arquitectura modular": se puede agregar un nuevo tipo de chequeo (por
ejemplo, una base de datos PostgreSQL) creando un archivo nuevo en
`internal/checker/` sin tocar `serialcomm` ni el protocolo con el micro:bit.

### El protocolo serial

Es intencionalmente muy simple: un byte ASCII + salto de línea.

| Código enviado | Significado | Ícono en el micro:bit |
|---|---|---|
| `O\n` | Todos los servicios OK | Carita feliz |
| `E\n` | Al menos uno caído/error | X |
| `U\n` | Sin error confirmado, pero algo no respondió a tiempo (timeout) | Signo de interrogación |

El estado enviado es el **agregado** de todos los servicios (función
`checker.Overall`): si un solo servicio está en ERROR, se manda `E` aunque
los demás estén bien. Eso es una decisión de diseño para un proyecto
escolar con una matriz de 5x5 LEDs, que no da para mostrar el detalle de
varios servicios a la vez. En la exposición pueden mencionar como posible
extensión: ciclar entre íconos de cada servicio cada 2 segundos, o mandar
un byte adicional identificando a qué servicio corresponde el estado.

## Preparar el entorno (Windows 11)

1. Instalar Go: bajar el instalador de https://go.dev/dl/ (el `.msi` para
   Windows), correrlo y aceptar las opciones por defecto. Reabrir
   PowerShell después de instalar para que reconozca el comando `go`.
2. Verificar la instalación:

   ```powershell
   go version
   ```

3. Pararse en la carpeta del proyecto y compilar:

   ```powershell
   cd monitor-microbit
   go mod tidy      # descarga go.bug.st/serial y genera go.sum
   go build ./...   # compila todo para verificar que no haya errores
   ```

   Si `go mod tidy` da error de red al bajar `go.bug.st/serial` (pasa en
   algunas redes universitarias que filtran dominios raros), no hace
   falta tocar nada: el `go.mod` ya trae un `replace` que lo baja desde
   GitHub en su lugar, así que alcanza con reintentar el comando.

## Flashear el micro:bit (Windows)

1. Entrar a https://python.microbit.org desde Chrome o Edge.
2. Pegar o cargar el contenido de `microbit/main.py` en el editor.
3. Conectar el micro:bit a la PC por USB. Windows lo va a reconocer como
   una unidad extraíble llamada `MICROBIT` (aparece en el Explorador de
   archivos, como un pendrive).
4. En el editor, tocar "Download" (o "Guardar") para descargar un archivo
   `.hex`.
5. Arrastrar ese `.hex` descargado (normalmente en `Descargas`) hacia la
   unidad `MICROBIT` en el Explorador de archivos, como si copiaran un
   archivo a un pendrive. El micro:bit va a parpadear una lucecita
   amarilla mientras graba, y después se reinicia solo.
6. Al arrancar, va a mostrar una flecha por medio segundo (confirmación
   de que el programa cargó) y quedar con la pantalla apagada, esperando
   datos por USB.

Este método (arrastrar el `.hex`) funciona siempre, sin depender de
permisos especiales del navegador.

## Encontrar el puerto COM del micro:bit

1. Con el micro:bit conectado, abrir el **Administrador de dispositivos**
   (buscarlo en el menú de Inicio).
2. Desplegar **Puertos (COM y LPT)**.
3. Va a aparecer algo como `mbed Serial Port (COM3)`. Ese `COM3` (el
   número puede variar) es el valor que van a poner en `serial_port`
   dentro de `configs/config.json`.

Si tienen dudas de cuál es, pueden desconectar el micro:bit, ver qué
puertos hay, volver a conectarlo y ver cuál apareció nuevo.

## Configurar `configs/config.json`

- `serial_port`: el `COMx` que encontraron en el Administrador de
  dispositivos (ej. `"COM3"`).
- `poll_interval`: cualquier duración válida de Go, ej. `"5s"`, `"30s"`, `"1m"`.
- `services`: lista de servicios, cada uno con `type` (`http`, `docker`,
  `tcp` o `ping`), `name` (para los logs), `target` y `timeout_seconds`.

## Correr el programa

```powershell
go run ./cmd/monitor configs/config.json
```

Van a ver logs como:

```
[API principal] OK (200 OK) - 45.2ms
[Contenedor DB] ERROR (exited) - 3.1ms
[NAS (SMB)] OK (puerto abierto) - 12ms
[Router] OK (responde) - 8ms
estado general: ERROR -> enviado 'E' al micro:bit
```

## Cómo probar los checkers sin tener todavía servicios reales

Antes de tener el NAS, el contenedor Docker, o hasta el micro:bit a mano,
pueden validar que la lógica de cada checker funciona usando el servidor
de prueba liviano que incluye este proyecto en `testing/testserver.go`.
Es un único archivo de Go, sin dependencias externas (nada de Python ni
instaladores extra): levanta un endpoint HTTP que siempre responde OK y
un puerto TCP abierto, ambos en `127.0.0.1`.

1. Abrir una terminal de PowerShell y correr el servidor de prueba (dejarla abierta):

   ```powershell
   cd monitor-microbit\testing
   go run testserver.go
   ```

   Debería mostrar:
   ```
   HTTP de prueba escuchando en http://127.0.0.1:8090/health
   Puerto TCP de prueba abierto en 127.0.0.1:9090
   ```

2. Abrir OTRA terminal de PowerShell (sin cerrar la anterior) y correr el
   monitor apuntando al config de pruebas que ya viene armado:

   ```powershell
   cd monitor-microbit
   go run ./cmd/monitor testing\config.test.json
   ```

   Ese config ya incluye casos de `OK` (contra el testserver) y de
   `ERROR` (contra puertos donde no hay nada escuchando, el `9999`), para
   los checkers de tipo `http`, `tcp` y `ping`. Si todavía no tienen el
   micro:bit conectado, no pasa nada: van a ver en la consola el estado
   de cada servicio igual, y solo va a fallar la línea de envío por
   serial (con un mensaje claro tipo "no se pudo abrir COM3"), sin que el
   programa se caiga.

3. Cuando ya tengan el micro:bit conectado y flasheado, cambien
   `serial_port` en `testing/config.test.json` (o en el config real) por
   el `COMx` correcto y corran de nuevo: ahora sí deberían ver el ícono
   cambiar en la matriz de LEDs según el estado agregado.

4. Para ver el caso `UNKNOWN` (timeout, ni OK ni ERROR confirmado),
   agreguen temporalmente un servicio de tipo `ping` apuntando a una IP
   que no exista en su red (por ejemplo, algo fuera de su rango de LAN,
   tipo `10.99.99.99` si su red no usa ese rango) con `timeout_seconds`
   bajo, y va a tardar exactamente ese tiempo en marcar `UNKNOWN`.

### (Opcional) Simular el micro:bit sin tenerlo conectado

Si quieren probar el envío por serial completo sin tener el micro:bit a
mano, existe una herramienta llamada **com0com** (gratuita, de código
abierto) que crea un par de puertos COM virtuales conectados entre sí en
Windows, similar a un cable USB simulado. Es un paso opcional y no hace
falta para el proyecto: la forma más simple y confiable de probar el
envío serial sigue siendo con el micro:bit real conectado.

## Guía de pruebas de punta a punta

1. **Probar cada checker por separado.** Antes de conectar el micro:bit,
   corran el programa con un solo servicio en `configs/config.json` (por ejemplo
   solo el `http`) apuntando a algo que sepan que funciona (ej.
   `https://example.com`) y revisen que el log diga `OK`. Después apunten a
   una URL que no exista para confirmar que da `ERROR`, y para forzar un
   `UNKNOWN` por timeout apunten a una IP de su propia red que no esté
   usada por ningún dispositivo (por ejemplo, si su router reparte
   `192.168.1.x`, prueben con `192.168.1.250` si saben que ahí no hay
   nada conectado) con un `timeout_seconds` bajo, como `2`.

2. **Probar el Docker checker.** Necesitan Docker Desktop instalado en
   Windows (con el motor de WSL2 activado, que es la opción por defecto
   al instalarlo). Levanten un contenedor de prueba desde PowerShell:

   ```powershell
   docker run -d --name test-nginx nginx
   ```

   Apúntenlo en el config y debe dar `OK`. Después:

   ```powershell
   docker stop test-nginx
   ```

   y verificar que pase a `ERROR` con mensaje `exited`.

3. **Probar el TCP/ping checker contra el NAS.** Usen la IP real del NAS
   de su red y el puerto correspondiente (445 para SMB, por ejemplo).
   Desconecten el NAS de la red (o apáguenlo) y confirmen que el estado
   cambia a `ERROR` o `UNKNOWN` según cómo responda la red.

4. **Probar la comunicación serial sola**, sin lógica de negocio: pueden
   usar un programa como **PuTTY** (gratuito, común en Windows) en modo
   "Serial", apuntado al mismo `COMx` con baud rate `115200`, y escribir
   a mano `O`, `E` o `U` seguido de Enter para verificar que el micro:bit
   cambia de ícono cada vez. Ojo: mientras PuTTY tenga el puerto COM
   abierto, el programa en Go no va a poder conectarse (un puerto serial
   solo lo puede usar un programa a la vez), así que cierren PuTTY antes
   de volver a correr `go run ./cmd/monitor configs/config.json`.

5. **Prueba end-to-end completa:** correr `go run ./cmd/monitor configs/config.json` con el
   micro:bit conectado y varios servicios reales. Ir apagando/desconectando
   servicios uno por uno mientras el programa corre, y verificar en vivo
   que tanto los logs como el ícono en el micro:bit cambian en el
   siguiente ciclo de polling.

6. **Probar la reconexión.** Con el programa corriendo, desconecten
   físicamente el cable USB del micro:bit. El log debe mostrar el error de
   escritura ("no se pudo enviar estado al micro:bit"). Vuelvan a
   conectarlo: en el siguiente ciclo, `SendStatus` detecta que `m.port` es
   `nil` y reabre el puerto solo, sin reiniciar el programa de Go.

## Ideas de extensión (para la exposición, si sobra tiempo)

- Mandar un byte de "servicio" además del de "estado", y que el micro:bit
  cicle entre los últimos N estados recibidos.
- Reemplazar `DockerChecker` (que usa el CLI) por uno que hable directo
  con la Docker Engine API vía su socket Unix, sin depender de que el
  comando `docker` esté en el PATH.
- Usar el botón A/B del micro:bit para pedir un refresco inmediato,
  mandando un byte desde el micro:bit hacia la PC (comunicación bidireccional).
