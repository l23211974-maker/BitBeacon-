# main.py - corre en el micro:bit (MicroPython)
#
# Recibe por USB (UART sobre el mismo cable USB) un código de un solo
# carácter enviado por el programa en Go, y dibuja el ícono correspondiente
# en la matriz de 5x5 LEDs:
#   'O' -> OK        -> carita feliz
#   'E' -> ERROR      -> X
#   'U' -> UNKNOWN    -> signo de interrogación
#
# Nota sobre uart.init(): por defecto, el puerto USB del micro:bit está
# conectado al REPL de MicroPython. Al llamar uart.init(), "tomamos" ese
# mismo canal USB para nuestro propio protocolo en vez de usarlo como
# consola interactiva. Es la forma más simple de mandar datos desde la PC
# sin necesitar hardware adicional (no hace falta un módulo de radio ni
# cables extra: se usa el mismo cable USB que ya alimenta al micro:bit).

from microbit import *

uart.init(baudrate=115200)

# Íconos como Image de 5x5. Usamos strings de MicroPython: cada dígito
# 0-9 es el brillo de un LED (9 = máximo). "HAPPY" ya viene incluido en
# la librería, pero definimos todos explícitamente para que se entienda
# bien el patrón en la exposición.

ICON_OK = Image(
    "00000:"
    "09090:"
    "00000:"
    "90009:"
    "09990"
)

ICON_ERROR = Image(
    "90009:"
    "09090:"
    "00900:"
    "09090:"
    "90009"
)

ICON_UNKNOWN = Image(
    "09990:"
    "09009:"
    "00090:"
    "00900:"
    "00900"
)

# Ícono que se muestra un instante al arrancar, para confirmar visualmente
# que el micro:bit está vivo y esperando datos por USB.
display.show(Image.ARROW_E)
sleep(500)
display.clear()

buffer = b""

while True:
    if uart.any():
        data = uart.read()
        if data:
            buffer += data
            # El protocolo termina cada mensaje con '\n', así que vamos
            # cortando el buffer en líneas completas a medida que llegan.
            while b"\n" in buffer:
                line, buffer = buffer.split(b"\n", 1)
                code = line.strip()
                if code == b"O":
                    display.show(ICON_OK)
                elif code == b"E":
                    display.show(ICON_ERROR)
                elif code == b"U":
                    display.show(ICON_UNKNOWN)
                # Cualquier otro código se ignora silenciosamente,
                # para tolerar ruido o basura en la línea serial.
    sleep(100)
