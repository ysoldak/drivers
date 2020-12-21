// +build arduino

package main

import "machine"

// Configuration for the Arduino Uno.
var (
	pwm = machine.Timer1
	pin = machine.D9
)
