// Package protocol define los mensajes entre el plugin y el helper.
//
// Cada mensaje es [1 byte tipo][4 bytes longitud big-endian][payload].
// Lo que empieza con mayúscula se exporta (se ve desde otros paquetes);
// lo que empieza con minúscula es privado de este paquete.
package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Version del protocolo. Se sube cuando plugin y helper dejan de entenderse.
const Version byte = 1

// Kind es el tipo de un mensaje. Es un tipo nuevo basado en byte: el
// compilador no deja pasar un byte cualquiera donde se espera un Kind.
type Kind byte

// Tipos de mensaje. iota vale 0 en la primera línea del bloque y sube de uno
// en uno; las líneas sin valor repiten la expresión anterior (iota + 1).
const (
	Hello  Kind = iota + 1 // helper → plugin: versiones
	Data                   // ambos sentidos: bytes del terminal
	Resize                 // plugin → helper: columnas y filas
	Exit                   // helper → plugin: código de salida del shell
)

// String es un método de Kind. Cualquier tipo con un método String() string
// cumple la interfaz fmt.Stringer, y fmt lo usa al imprimir: los tests
// mostrarán [HELLO DATA EXIT] en lugar de [1 2 4].
func (k Kind) String() string {
	switch k {
	case Hello:
		return "HELLO"
	case Data:
		return "DATA"
	case Resize:
		return "RESIZE"
	case Exit:
		return "EXIT"
	}
	return fmt.Sprintf("Kind(%d)", byte(k))
}

// Cabecera: 1 byte de tipo + 4 bytes de longitud.
const headerSize = 5

// Límite por mensaje, para que un dato corrupto no nos haga reservar 4 GB.
const maxPayload = 1 << 20 // 1 MiB

// WriteFrame envía un mensaje completo en una sola escritura.
func WriteFrame(w io.Writer, kind Kind, payload []byte) error {
	buf := make([]byte, headerSize+len(payload))
	buf[0] = byte(kind)
	binary.BigEndian.PutUint32(buf[1:], uint32(len(payload)))
	copy(buf[headerSize:], payload)
	_, err := w.Write(buf)
	return err
}

// ReadFrame lee un mensaje completo. Una función de Go puede devolver varios
// valores; el error va siempre al final.
//
// Devuelve io.EOF si la entrada se cerró justo entre dos mensajes, e
// io.ErrUnexpectedEOF si se cortó a mitad de uno.
func ReadFrame(r io.Reader) (Kind, []byte, error) {
	var header [headerSize]byte // array de tamaño fijo; header[:] es su slice
	// ReadFull insiste hasta llenar el buffer: un solo Read puede devolver
	// menos bytes de los pedidos.
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return 0, nil, err
	}
	size := binary.BigEndian.Uint32(header[1:])
	if size > maxPayload {
		return 0, nil, fmt.Errorf("mensaje demasiado grande: %d bytes", size)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(r, payload); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	return Kind(header[0]), payload, nil
}

// HelloPayload: [versión del protocolo][versión del helper en texto].
func HelloPayload(helperVersion string) []byte {
	return append([]byte{Version}, helperVersion...)
}

// ExitPayload: código de salida como int32 big-endian. En Windows los códigos
// son de 32 bits (por ejemplo 0xC000013A al cerrar con Ctrl+C).
func ExitPayload(code int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(int32(code)))
	return buf
}

// ParseResize: [columnas uint16][filas uint16]. Los resultados tienen nombre,
// lo que sirve de documentación.
func ParseResize(payload []byte) (cols, rows uint16, err error) {
	if len(payload) != 4 {
		return 0, 0, fmt.Errorf("RESIZE mide %d bytes, se esperaban 4", len(payload))
	}
	cols = binary.BigEndian.Uint16(payload[0:2])
	rows = binary.BigEndian.Uint16(payload[2:4])
	// Un panel oculto mide 0: ese tamaño no se le puede dar a una terminal.
	if cols == 0 || rows == 0 {
		return 0, 0, fmt.Errorf("RESIZE de %dx%d: el tamaño debe ser mayor que 0", cols, rows)
	}
	return cols, rows, nil
}
