package terminal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Terminal implementa ports.SessionIO usando stdin/stdout.
type Terminal struct {
	scanner *bufio.Scanner
}

// New crea un adaptador de terminal listo para leer de os.Stdin.
func New() *Terminal {
	return &Terminal{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (t *Terminal) ReadLine() (string, error) {
	if !t.scanner.Scan() {
		if err := t.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(t.scanner.Text()), nil
}

func (t *Terminal) WriteLine(msg string) {
	fmt.Println(msg)
}

func (t *Terminal) Close() error {
	return nil
}
