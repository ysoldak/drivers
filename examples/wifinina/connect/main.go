// +build arduino_nano33 


// This example connects to Access Point and prints some info
package main

import (
	"encoding/binary"
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/wifinina"
)

// access point info
const ssid = ""
const pass = ""

// these are the default pins for the Arduino Nano33 IoT.
// change these to connect to a different UART or pins for the ESP8266/ESP32
var (

	// these are the default pins for the Arduino Nano33 IoT.
	spi = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	adaptor *wifinina.Device
)

func setup() {

	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	adaptor = wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	adaptor.Configure()
}

func main() {

	setup()

	waitSerial()

	connectToAP()

	for {
		println("---------------------------------")
		printIPs()
		printTime()
		printMacAddress()
		time.Sleep(10 * time.Second)
	}

}

func printIPs() {
	ip, subnet, gateway, err := adaptor.GetIP()
	if err != nil {
		println("IP: Unknown (error: ", err.Error(), ")")
		return
	}
	println("IP: ", ip.String())
	println("Subnet: ", subnet.String())
	println("Gateway IP: ", gateway.String())
}

func printTime() {
	print("Time: ")
	t, err := adaptor.GetTime()
	if err != nil {
		println("Unknown (error: ", err.Error(), ")")
	}
	println(time.Unix(int64(t), 0).String())
}

func printMacAddress() {
	print("MAC Address: ")
	b := make([]byte, 8)
	mac, err := adaptor.GetMACAddress()
	if err != nil {
		println("Unknown (", err.Error(), ")")
	}
	binary.LittleEndian.PutUint64(b, uint64(mac))
	macAddress := ""
	for i := 5; i >= 0; i-- {
		macAddress += fmt.Sprintf("%0X", b[i])
		if i != 0 {
			macAddress += ":"
		}
	}
	println(macAddress)
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.UART0.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

// connect to access point
func connectToAP() {
	if len(ssid) == 0 || len(pass) == 0 {
		for {
			println("Connection failed: Either ssid or password not set")
			time.Sleep(10 * time.Second)
		}
	}
	time.Sleep(2 * time.Second)
	message("Connecting to " + ssid)
	adaptor.SetPassphrase(ssid, pass)
	for st, _ := adaptor.GetConnectionStatus(); st != wifinina.StatusConnected; {
		message("Connection status: " + st.String())
		time.Sleep(1 * time.Second)
		st, _ = adaptor.GetConnectionStatus()
	}
	message("Connected.")
}

func message(msg string) {
	println(msg, "\r")
}
