package main

import (
	"net"
	"fmt"
	"./muxers"
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
	flvMuxer := muxers.NewFlvMuxer()
	rtmpSink := muxers.NewRtmpSink("rtmp://s1.transcoding.svoe.tv/source1", "test_stream")

	muxers.Bridge(rtpDemuxer.OutputChan, rtpH264Demuxer.InputChan)
	muxers.Bridge(rtpH264Demuxer.OutputChan, flvMuxer.InputChan)
	muxers.Bridge(flvMuxer.OutputChan, rtmpSink.InputChan)

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
