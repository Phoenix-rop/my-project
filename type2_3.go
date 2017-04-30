package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tarm/goserial"
)

var Preamble int = 0
var Type int = 1
var TypeData int = 2
var Length int = 3
var Payload int = 4
var Sequence int = 5
var Checksum int = 6
var State int = 0
var Preamble_length int = 0

type PackageData struct {
	Type_payload byte
	Seq          byte
}

var Payload_length int = 0
var Current_Payload_length int = 0

func readserial() {
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
			//getdata1(data)
			getdata2(data)
		}
	}
	s.Close()
}

func getdata1(data []byte) {
	var length int = 0
	var count int = 0
	var state int = 0
	data_playload := make([]byte, 6)
	for i := 0; i < len(data); i++ {
		switch state {
		case 0, 2, 4, 6:
			if data[i] == 0x80 {
				state++
			} else {
				state = 0
			}
		case 1, 3, 5, 7:
			if data[i] == 0x00 {
				state++
			} else {
				state = 0
			}
		case 8:
			if data[i] == 0x01 {
				state++
			} else {
				state = 0
			}
		case 9:
			if len(data) > 0 {
				length = int(data[i])
				if length == 6 {
					state++
				}
			}
		case 10:
			if data[i] != 0 {
				data_playload[count] = data[i]
				count++
			} else {
				state = 0
			}
			if count == length {
				convert_byte(data_playload)
				length = 0
				count = 0
				state = 0
			}
		}
	}
	HR_BR_MM := make([]int16, 3)
	PrintHex(data_playload)
	HR_BR_MM[0] = int16(uint16(data_playload[0])<<8 | uint16(data_playload[1]))
	HR_BR_MM[1] = int16(uint16(data_playload[2])<<8 | uint16(data_playload[3]))
	HR_BR_MM[2] = int16(uint16(data_playload[4])<<8 | uint16(data_playload[5]))
	if HR_BR_MM[2] != 0 {
		fmt.Printf("Movement (MM) : ")
		fmt.Println(HR_BR_MM[2])
		//convert_byte2(Payload_type2)
		i := int64(HR_BR_MM[2])
		str2 := strconv.FormatInt(int64(i), 10)
		apiBase2 := "https://api.thingspeak.com/update?api_key=DGV740AU28L9VU1K&field3="
		api2 := apiBase2 + str2
		fmt.Println(api2)
		resp, err := http.Get(api2)
		if err != nil {
			// handle error
		}
		defer resp.Body.Close()
		//body, err := ioutil.ReadAll(resp.Body)
	}
}

func convert_byte(convert []byte) {
	HR_BR_MM := make([]int16, 3)
	PrintHex(convert)
	HR_BR_MM[0] = int16(uint16(convert[0])<<8 | uint16(convert[1]))
	HR_BR_MM[1] = int16(uint16(convert[2])<<8 | uint16(convert[3]))
	HR_BR_MM[2] = int16(uint16(convert[4])<<8 | uint16(convert[5]))
	fmt.Printf("HR: %d\n", HR_BR_MM[0])
	fmt.Printf("BR: %d\n", HR_BR_MM[1])
	fmt.Printf("MM: %d\n", HR_BR_MM[2])

	p := fmt.Println
	now := time.Now()
	then := time.Date(2016, 11, 25, 12, 0, 0, 0, time.UTC)
	diff := now.Sub(then)
	var time float64 = 0.0
	time = diff.Seconds()
	p(time)
	p(" : time")

	sum_data := make([]string, 4)
	sum_data[0] = strconv.FormatFloat(time, 'E', -1, 64)
	sum_data[1] = strconv.Itoa(int(HR_BR_MM[0]))
	sum_data[2] = strconv.Itoa(int(HR_BR_MM[1]))
	sum_data[3] = strconv.Itoa(int(HR_BR_MM[2]))
	fmt.Println("2d", sum_data)

	output := make([][]string, 1)
	for _, name := range sum_data {
		output[0] = append(output[0], name)
	}

	fmt.Println("%v", output)

	csvfile, err := os.OpenFile("packagedata.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	checkError("Cannot create file", err)
	defer csvfile.Close()

	writer := csv.NewWriter(csvfile)

	for _, value := range output {
		err := writer.Write(value)
		checkError("Cannot write to file", err)
	}
	writer.Flush()
}

func getdata2(data []byte) {
	Data_type := PackageData{}
	var count2 int = 0
	var count3 int = 0
	Payload_type2 := make([]byte, 1)
	Payload_type3 := make([]byte, 1)
	for i := 0; i < len(data); i++ {
		switch State {
		case Preamble:
			if Preamble_length%2 == 0 {
				if data[i] == 0x80 {
					Preamble_length++
				} else {
					Preamble_length = 0
				}
			} else {
				if data[i] == 0x00 {
					Preamble_length++
				} else {
					Preamble_length = 0
				}
			}
			if Preamble_length > 7 {
				State = TypeData
			}
			break

		case TypeData:
			if data[i] == 0x02 || data[i] == 0x03 {
				State = Length
				Data_type.Type_payload = data[i]
			} else {
				State = Preamble
			}
			break

		case Length:
			if data[i] > 0 {
				Payload_length = int(data[i])
				Current_Payload_length = 0
				State = Payload
			} else {
				State = Preamble
			}
			break
		case Payload:
			if data[i-2] == 0x02 {
				Payload_type2[count2] = data[i]
				Current_Payload_length++
			}
			if data[i-2] == 0x03 {
				Payload_type3[count3] = data[i]
				Current_Payload_length++
			}
			if Current_Payload_length > Payload_length {
				State = Sequence
			} else {
				State = Preamble
			}
			break
		case Sequence:
			Data_type.Seq = data[i]
			State = Checksum
			break
		case Checksum:
			State = Preamble
			break
		}
	}
	if Payload_type2[0] != 0 {
		fmt.Printf("Type 2 (HR) : ")
		fmt.Println(Payload_type2)
		//convert_byte2(Payload_type2)
		str1 := convert(Payload_type2)
		apiBase1 := "https://api.thingspeak.com/update?api_key=DGV740AU28L9VU1K&field1="
		api1 := apiBase1 + str1
		fmt.Println(api1)
		resp, err := http.Get(api1)
		if err != nil {
			// handle error
		}
		defer resp.Body.Close()
		//body, err := ioutil.ReadAll(resp.Body)
	}
	if Payload_type3[0] != 0 {
		fmt.Printf("					Type 3 (BR) : ")
		fmt.Println(Payload_type3)
		//convert_byte3(Payload_type3)
		str := convert(Payload_type3)
		apiBase := "https://api.thingspeak.com/update?api_key=DGV740AU28L9VU1K&field2="
		api := apiBase + str
		fmt.Println(api)
		resp, err := http.Get(api)
		if err != nil {
			// handle error
		}
		defer resp.Body.Close()
		//body, err := ioutil.ReadAll(resp.Body)
	}

}

func convert_byte2(convert []byte) {
	BR := make([]int16, 1)
	PrintHex(convert)
	BR[0] = int16(uint16(convert[0]))
	fmt.Printf("HR: %d\n", BR[0])

	now := time.Now()
	then := time.Date(2016, 11, 25, 12, 0, 0, 0, time.UTC)
	diff := now.Sub(then)
	var time float64 = 0.0
	time = diff.Seconds()
	fmt.Print(time)
	fmt.Println(" : time")

	sum_data := make([]string, 2)
	sum_data[0] = strconv.FormatFloat(time, 'E', -1, 64)
	sum_data[1] = strconv.Itoa(int(BR[0]))

	fmt.Println("2d", sum_data)

	output := make([][]string, 1)
	for _, name := range sum_data {
		output[0] = append(output[0], name)
	}

	fmt.Println("%v", output)

	csvfile, err := os.OpenFile("packagedatatype2.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	checkError("Cannot create file", err)
	defer csvfile.Close()

	writer := csv.NewWriter(csvfile)

	for _, value := range output {
		err := writer.Write(value)
		checkError("Cannot write to file", err)
	}
	writer.Flush()
}

func convert_byte3(convert []byte) {
	BR := make([]int16, 1)
	fmt.Print("					")
	PrintHex(convert)
	BR[0] = int16(uint16(convert[0]))
	fmt.Printf("					BR: %d\n", BR[0])

	now := time.Now()
	then := time.Date(2016, 11, 25, 12, 0, 0, 0, time.UTC)
	diff := now.Sub(then)
	var time float64 = 0.0
	time = diff.Seconds()
	fmt.Print("					")
	fmt.Print(time)
	fmt.Println("	 : time")

	sum_data := make([]string, 2)
	sum_data[0] = strconv.FormatFloat(time, 'E', -1, 64)
	sum_data[1] = strconv.Itoa(int(BR[0]))

	fmt.Println("					2d", sum_data)

	output := make([][]string, 1)
	for _, name := range sum_data {
		output[0] = append(output[0], name)
	}

	fmt.Println("					%v", output)

	csvfile, err := os.OpenFile("packagedatatype3.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	checkError("Cannot create file", err)
	defer csvfile.Close()

	writer := csv.NewWriter(csvfile)

	for _, value := range output {
		err := writer.Write(value)
		checkError("Cannot write to file", err)
	}
	writer.Flush()
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func PrintHex(data_playload []byte) {
	for i := 0; i < len(data_playload); i++ {
		if data_playload[i] < 0x10 {
			fmt.Printf("0x0%X ", data_playload[i])
		} else {
			fmt.Printf("0x%X ", data_playload[i])
		}
	}
	fmt.Println("")

}

func convert(b []byte) string {
	s := make([]string, len(b))
	for i := range b {
		s[i] = strconv.Itoa(int(b[i]))
	}
	return strings.Join(s, ",")
}

func main() {
	readserial()
}
