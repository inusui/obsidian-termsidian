// Este test está en el mismo paquete (protocol), así que también ve lo
// privado, como headerSize y maxPayload.
package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

// Un test es una función TestXxx(t *testing.T). Este es "table-driven":
// una lista de casos y un bucle que los prueba todos.
func TestFrameRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		kind    Kind
		payload []byte
	}{
		{"data", Data, []byte("echo hola\r\n")},
		{"vacío", Data, []byte{}},
		{"utf-8", Data, []byte("ñandú 🐧")},
		{"hello", Hello, HelloPayload("1.2.3")},
		{"exit", Exit, ExitPayload(3)},
	}

	// Primero se escriben todos seguidos, como llegarían por la tubería.
	var buf bytes.Buffer // bytes.Buffer es un io.Reader y un io.Writer a la vez
	for _, c := range cases {
		if err := WriteFrame(&buf, c.kind, c.payload); err != nil {
			t.Fatalf("%s: WriteFrame: %v", c.name, err)
		}
	}

	// Después se leen uno a uno y deben salir idénticos y en orden.
	for _, c := range cases {
		kind, payload, err := ReadFrame(&buf)
		if err != nil {
			t.Fatalf("%s: ReadFrame: %v", c.name, err)
		}
		if kind != c.kind || !bytes.Equal(payload, c.payload) {
			t.Errorf("%s: se leyó (%v, %q), se esperaba (%v, %q)",
				c.name, kind, payload, c.kind, c.payload)
		}
	}

	// Sin más datos, la lectura debe terminar con io.EOF limpio.
	if _, _, err := ReadFrame(&buf); err != io.EOF {
		t.Errorf("al final: err = %v, se esperaba io.EOF", err)
	}
}

func TestReadFrameErrors(t *testing.T) {
	full := new(bytes.Buffer)
	WriteFrame(full, Data, []byte("hola"))
	frame := full.Bytes()

	huge := make([]byte, headerSize)
	huge[0] = byte(Data)
	binary.BigEndian.PutUint32(huge[1:], maxPayload+1)

	cases := []struct {
		name  string
		input []byte
		want  error // nil = cualquier error vale, solo tiene que fallar
	}{
		{"cabecera cortada", frame[:3], io.ErrUnexpectedEOF},
		{"payload cortado", frame[:len(frame)-1], io.ErrUnexpectedEOF},
		{"sin payload", frame[:headerSize], io.ErrUnexpectedEOF},
		{"demasiado grande", huge, nil},
	}
	for _, c := range cases {
		// t.Run crea un subtest con nombre: si falla, se sabe cuál fue.
		t.Run(c.name, func(t *testing.T) {
			_, _, err := ReadFrame(bytes.NewReader(c.input))
			if err == nil {
				t.Fatal("se esperaba un error")
			}
			if c.want != nil && !errors.Is(err, c.want) {
				t.Errorf("err = %v, se esperaba %v", err, c.want)
			}
		})
	}
}

func TestParseResize(t *testing.T) {
	cols, rows, err := ParseResize([]byte{0, 120, 0, 40})
	if err != nil || cols != 120 || rows != 40 {
		t.Errorf("ParseResize = (%d, %d, %v), se esperaba (120, 40, nil)", cols, rows, err)
	}
	if _, _, err := ParseResize([]byte{1, 2, 3}); err == nil {
		t.Error("un RESIZE de 3 bytes debería fallar")
	}
	if _, _, err := ParseResize([]byte{0, 0, 0, 40}); err == nil {
		t.Error("un RESIZE de 0 columnas debería fallar")
	}
}

func TestExitPayloadWindowsCode(t *testing.T) {
	// Go devuelve 0xC000013A (proceso cerrado con Ctrl+C en Windows) como
	// 3221225786. No cabe en un int32 con signo: el plugin lo verá negativo,
	// que es como lo muestran las herramientas de Windows.
	got := int32(binary.BigEndian.Uint32(ExitPayload(0xC000013A)))
	if got != -1073741510 {
		t.Errorf("ExitPayload(0xC000013A) se leyó como %d", got)
	}
}

func TestKindString(t *testing.T) {
	if s := Resize.String(); s != "RESIZE" {
		t.Errorf("Resize.String() = %q", s)
	}
	if s := Kind(99).String(); s != "Kind(99)" {
		t.Errorf("Kind(99).String() = %q", s)
	}
}
