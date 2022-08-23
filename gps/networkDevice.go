package gps

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/b3nn0/stratux/v2/common"
)

type NetworkDevice struct {
	rxMessageCh          chan<- RXMessage
	discoveredDevicesCh  chan<- DiscoveredDevice
	qh                   *common.QuitHelper
}


func NewNetworkGPSDevice(rxMessageCh chan<- RXMessage, discoveredDevicesCh chan<- DiscoveredDevice) NetworkDevice {
	m := NetworkDevice{
		rxMessageCh:          rxMessageCh,
		discoveredDevicesCh:  discoveredDevicesCh,
		qh:                   common.NewQuitHelper(),
	}
	return m
}

func (b *NetworkDevice) defaultDeviceDiscoveryData(name string, connected bool) DiscoveredDevice {
	return DiscoveredDevice{
		Name:               name,
		Connected:          connected,
		GpsDetectedType:    GPS_TYPE_NETWORK, // TODO: Should we be more specific for example mention that it's an SoftRF device?
		GpsSource:          GPS_SOURCE_NETWORK,
		GpsTimeOffsetPpsMs: 100.0 * time.Millisecond,
	}
}

func (b *NetworkDevice) updateDeviceDiscovery(name string, connected bool) {
	b.discoveredDevicesCh <- b.defaultDeviceDiscoveryData(name, connected)
}

/* Server that can be used to feed NMEA data to, e.g. to connect OGN Tracker wirelessly */
func (n *NetworkDevice) tcpNMEAInListener(port int) {
	n.qh.Add()
	defer n.qh.Done()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		log.Printf(err.Error())
		return
	}

	go func() {
		<- n.qh.C
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if n.qh.IsQuit() {
			return
		}
		if err != nil {
			log.Printf(err.Error())
			continue
		}
		go n.handleNmeaInConnection(conn)
		time.Sleep(250 * time.Millisecond)
	}	
}

func (n *NetworkDevice) handleNmeaInConnection(c net.Conn) {
	n.qh.Add()
	defer n.qh.Done()
	reader := bufio.NewReader(c)
	remoteAddress := c.RemoteAddr().String()
	n.updateDeviceDiscovery(remoteAddress, true)

	go func() {
		<- n.qh.C
		c.Close()
	}()

	for {
		line, err := reader.ReadString('\r')
		if err != nil {
			break
		}
		trimedLine := strings.TrimSpace(line)
		if len(trimedLine) > 0 {
			n.rxMessageCh <- RXMessage{
				Name:     remoteAddress,
				NmeaLine: trimedLine,
			}		
		}
	}
	n.updateDeviceDiscovery(remoteAddress, false)
}

func (n *NetworkDevice) Stop() {
	n.qh.Quit()
}

func (n *NetworkDevice) Run() {
	ports := [...]int{30011} 
	for _, port := range ports {
		go n.tcpNMEAInListener(port)
	}
}

