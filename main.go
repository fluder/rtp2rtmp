package main

import (
	"net"
	"fmt"
	"./muxers"
	"os"
)

func main() {
	addr, _ := net.ResolveUDPAddr("udp4", ":5000")
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Printf("Error during listening udp socket: %s\n", err.Error())
		return
	}

	rtpDemuxer := muxers.NewRtpDemuxer()
	rtpH264Demuxer := muxers.NewRtpH264Depacketizer()

	muxers.Bridge(rtpDemuxer.OutputChan, rtpH264Demuxer.InputChan)

	go func() {
		f, _ := os.Create("/tmp/test.h264")
		f.Write([]byte{0, 0, 0, 1})
		for {
			packet := (<-rtpH264Demuxer.OutputChan).(*muxers.RtpPacket)
			fmt.Println(packet.Payload[0] & 31)
			fmt.Println(packet.Timestamp)
			if packet.Payload[0] & 31 < 23 {
				f.Write(packet.Payload)
				f.Write([]byte{0, 0, 1})
			}
		}
	}()

	for {
		buf := make([]byte, 1500)
		size, _, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("Error during receiving packet: %s\n", err.Error())
			continue
		}

		rtpDemuxer.InputChan <- buf[:size]
	}
}
