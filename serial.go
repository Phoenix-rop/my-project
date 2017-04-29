//serial
package main

import (
	"fmt"

	"github.com/tarm/serial"
)

func main() {
	Check_Port_Baudrate := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200}
	s, err := serial.OpenPort(Check_Port_Baudrate)

	if err != nil {
		fmt.Println(err)
	}

	for {

		_, err = s.Write([]byte("\x16\x02N0C0 G A\x03\x0d\x0a"))

		if err != nil {
			fmt.Println(err)
		}

		data := make([]byte, 40)
		n, err := s.Read(data)

		if err != nil {
			fmt.Println(err)
		}
		if n > 0 {
			fmt.Println(data)
		}
	}
	s.Close()
}
