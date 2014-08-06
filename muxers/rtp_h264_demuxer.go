package muxers

import (
	"fmt"
	"os"
)

type RtpH264Demuxer struct {
	InputChan chan interface {}
	OutputChan chan interface {}
}

func NewRtpH264Demuxer() *RtpH264Demuxer {
	demuxer := &RtpH264Demuxer{
		InputChan: make(chan interface {}),
		OutputChan: make(chan interface {}),
	}

	go func() {
		f, _ := os.Create("/tmp/test.h264")

		for {
			packet := (<-demuxer.InputChan).(*RtpPacket)
			fmt.Println(packet)
			f.Write(packet.Payload[1:])
		}
	}()

	return demuxer
}

type RtpH264Packet struct {
	*RtpPacket
	NalNRI uint8
	NalType uint8
}
