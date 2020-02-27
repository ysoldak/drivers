package espat

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

const (
	TCPMuxSingle   = 0
	TCPMuxMultiple = 1

	TCPTransferModeNormal      = 0
	TCPTransferModeUnvarnished = 1
)

var (
	ErrTooManyConnections = errors.New("espat: too many open connections")
	ErrUnknownNetwork     = errors.New("espat: unknown network")
)

type Conn struct {
	device *Device
}

// GetDNS returns the IP address for a domain name.
func (d *Device) GetDNS(domain string) (string, error) {
	d.Set(TCPDNSLookup, "\""+domain+"\"")
	resp, err := d.Response(1000)
	if err != nil {
		return "", err
	}
	if !strings.Contains(string(resp), ":") {
		return "", errors.New("GetDNS error:" + string(resp))
	}
	r := strings.Split(string(resp), ":")
	if len(r) != 2 {
		return "", errors.New("Invalid domain lookup result")
	}
	res := strings.Split(r[1], "\r\n")
	return res[0], nil
}

// Dial connects to the address on the named network.
func (d *Device) Dial(network, address string) (net.Conn, error) {
	switch network {
	case "tcp":
		addr, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		err = d.connectTCPSocket(addr, port)
		if err != nil {
			return nil, err
		}
		return &Conn{device: d}, nil
	default:
		return nil, ErrUnknownNetwork
	}
}

// Close closes the current network connection, freeing the used resources.
func (c *Conn) Close() error {
	return c.device.disconnectSocket()
}

// Read reads data from the connection.
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.device.ReadSocket(b)
}

// Write writes data to the connection.
func (c *Conn) Write(b []byte) (n int, err error) {
	err = c.device.StartSocketSend(len(b))
	if err != nil {
		return
	}
	n, err = c.device.Write(b)
	if err != nil {
		return n, err
	}
	_, err = c.device.Response(1000)
	if err != nil {
		return n, err
	}
	return n, err
}

// connectTCPSocket creates a new TCP socket connection for the ESP8266/ESP32.
// Currently only supports single connection mode.
func (d *Device) connectTCPSocket(addr, port string) error {
	if d.connected {
		return ErrTooManyConnections
	}
	protocol := "TCP"
	val := "\"" + protocol + "\",\"" + addr + "\"," + port + ",120"
	err := d.Set(TCPConnect, val)
	if err != nil {
		return err
	}
	_, e := d.Response(3000)
	if e != nil {
		return e
	}
	return nil
}

// ConnectUDPSocket creates a new UDP connection for the ESP8266/ESP32.
func (d *Device) ConnectUDPSocket(addr, sendport, listenport string) error {
	if d.connected {
		return ErrTooManyConnections
	}
	protocol := "UDP"
	val := "\"" + protocol + "\",\"" + addr + "\"," + sendport + "," + listenport + ",2"
	err := d.Set(TCPConnect, val)
	if err != nil {
		return err
	}
	_, e := d.Response(3000)
	if e != nil {
		return e
	}
	return nil
}

// ConnectSSLSocket creates a new SSL socket connection for the ESP8266/ESP32.
// Currently only supports single connection mode.
func (d *Device) ConnectSSLSocket(addr, port string) error {
	protocol := "SSL"
	val := "\"" + protocol + "\",\"" + addr + "\"," + port + ",120"
	d.Set(TCPConnect, val)
	// this operation takes longer, so wait up to 6 seconds to complete.
	_, err := d.Response(6000)
	if err != nil {
		return err
	}
	return nil
}

// disconnectSocket disconnects the ESP8266/ESP32 from the current TCP/UDP connection.
func (d *Device) disconnectSocket() error {
	err := d.Execute(TCPClose)
	if err != nil {
		return err
	}
	_, e := d.Response(pause)
	if e != nil {
		return e
	}
	return nil
}

// SetMux sets the ESP8266/ESP32 current client TCP/UDP configuration for concurrent connections
// either single TCPMuxSingle or multiple TCPMuxMultiple (up to 4).
func (d *Device) SetMux(mode int) error {
	val := strconv.Itoa(mode)
	d.Set(TCPMultiple, val)
	_, err := d.Response(pause)
	return err
}

// GetMux returns the ESP8266/ESP32 current client TCP/UDP configuration for concurrent connections.
func (d *Device) GetMux() ([]byte, error) {
	d.Query(TCPMultiple)
	return d.Response(pause)
}

// SetTCPTransferMode sets the ESP8266/ESP32 current client TCP/UDP transfer mode.
// Either TCPTransferModeNormal or TCPTransferModeUnvarnished.
func (d *Device) SetTCPTransferMode(mode int) error {
	val := strconv.Itoa(mode)
	d.Set(TransmissionMode, val)
	_, err := d.Response(pause)
	return err
}

// GetTCPTransferMode returns the ESP8266/ESP32 current client TCP/UDP transfer mode.
func (d *Device) GetTCPTransferMode() ([]byte, error) {
	d.Query(TransmissionMode)
	return d.Response(pause)
}

// StartSocketSend gets the ESP8266/ESP32 ready to receive TCP/UDP socket data.
func (d *Device) StartSocketSend(size int) error {
	val := strconv.Itoa(size)
	d.Set(TCPSend, val)

	// when ">" is received, it indicates
	// ready to receive data
	r, err := d.Response(2000)
	if err != nil {
		return err
	}
	if strings.Contains(string(r), ">") {
		return nil
	}
	return errors.New("StartSocketSend error:" + string(r))
}

// EndSocketSend tell the ESP8266/ESP32 the TCP/UDP socket data sending is complete,
// and to return to command mode. This is only used in "unvarnished" raw mode.
func (d *Device) EndSocketSend() error {
	d.Write([]byte("+++"))

	_, err := d.Response(pause)
	return err
}
