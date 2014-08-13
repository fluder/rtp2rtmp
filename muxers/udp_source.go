package muxers

import (
	"fmt"
	"net"
)

type UdpSource struct {
	OutputChan chan interface {}
}

func NewUdpSource(port int) *UdpSource {
	source := &UdpSource{
		OutputChan: make(chan interface {}),
	}

	go func() {
		addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", port))
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			fmt.Printf("Error during listening udp socket: %s\n", err.Error())
			return
		}

		for {
			buf := make([]byte, 1500)
			size, _, err := conn.ReadFrom(buf)
			if err != nil {
				fmt.Printf("Error during receiving packet: %s\n", err.Error())
				continue
			}

			source.OutputChan <-buf[:size]
		}
	}()

	return source
}
