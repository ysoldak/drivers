// +build !arduino

package main

import "machine"

// Configuration for many boards, such as the PyBadge or any nRF52xxx based
// board.
var (
	//pwm = machine.PWM0
	//pin = machine.D12

	pwm = machine.PWM0
	pin = machine.A1
)
