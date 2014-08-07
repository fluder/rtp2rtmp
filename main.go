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
	//flvMuxer := muxers.NewFlvMuxer()

	muxers.Bridge(rtpDemuxer.OutputChan, rtpH264Demuxer.InputChan)
	//muxers.Bridge(rtpH264Demuxer.OutputChan, flvMuxer.InputChan)

	go func() {
		f, _ := os.Create("/tmp/raw1.h264")
		f.Write([]byte{0,0,0,1})
		for {
			data := (<-rtpH264Demuxer.OutputChan).(*muxers.RtpPacket)

			fmt.Println(data.Payload[0] & 31)
			f.Write(data.Payload)
			f.Write([]byte{0,0,1})
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
