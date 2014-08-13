package muxers

type RtpMPESDepacketizer struct {
	InputChan chan interface {}
	OutputChan chan interface {}
}

func NewRtpMPESDepacketizer() *RtpMPESDepacketizer {
	demuxer := &RtpMPESDepacketizer{
		InputChan: make(chan interface {}),
		OutputChan: make(chan interface {}),
	}

	go func() {
		var fragments []*RtpPacket

		for {
			packet := (<-demuxer.InputChan).(*RtpPacket)
			fragments = append(fragments, packet)

			// De-fragmentation
			if packet.Marker {
				payload := make([]byte, 0)
				for _, packet := range fragments {
					payload = append(payload, packet.Payload...)
				}
				packet.Payload = payload
				fragments = nil
			} else {
				continue
			}

			payload := packet.Payload
			auHeadersSize := uint16(0) | (uint16(payload[0]) << 8) | uint16(payload[1])
			payload = payload[2:]
			auHeadersCount := auHeadersSize / 16

			frameSizes := make([]uint16, 0)
			frameIndexes := make([]uint8, 0)
			for i := 0; i < int(auHeadersCount); i++ {
				auHeader := payload[:2]
				payload = payload[2:]
				frameSizes = append(frameSizes, uint16(0) | (uint16(auHeader[0]) << 5) | (uint16(auHeader[1]) >> 3))
				frameIndexes = append(frameIndexes, auHeader[1] & 0x07)
			}

			for i, frameSize := range frameSizes {
				framePayload := payload[:frameSize]
				payload = payload[frameSize:]

				framePacket := *packet
				framePacket.Payload = framePayload
				framePacket.Timestamp += uint32(i)*90000

				demuxer.OutputChan <- &framePacket
			}
		}
	}()

	return demuxer
}
