package muxers

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type FlvMuxer struct {
	InputChan chan interface {}
	OutputChan chan interface {}
}

func NewFlvMuxer() *FlvMuxer {
	muxer := &FlvMuxer{
		InputChan: make(chan interface {}),
		OutputChan: make(chan interface {}),
	}

	go func() {
		var SPS, PPS []byte
		var firstTimestamp uint32
		// Write Flv header
		muxer.OutputChan <-FlvHeader

		for {
			var videoDataPayload []byte
			packet := (<-muxer.InputChan).(*RtpPacket)
			fmt.Println(packet.Timestamp)
			if firstTimestamp == 0 {
				firstTimestamp = packet.Timestamp
			}
			if packet.Payload[0] & 31 == 7 {
				SPS = packet.Payload
			}
			if packet.Payload[0] & 31 == 8 {
				PPS = packet.Payload
			}
			if packet.Payload[0] & 31 == 7 || packet.Payload[0] & 31 == 8 {
				if SPS != nil && PPS != nil {
					record := &AVCDecoderConfigurationRecord{
						ConfigurationVersion: 1,
						AVCProfileIndication: SPS[1],
						ProfileCompatibility: SPS[2],
						AVCLevelIndication: SPS[3],
						SPS: SPS,
						PPS: PPS,
					}

					videoData := &VideoData{
						FrameType: FRAME_TYPE_KEY,
						CodecID: CODEC_AVC,
						AVCPacketType: AVC_SEQ_HEADER,
						CompositionTime: 0,
						Data: marshalAVCDecoderConfigurationRecord(record),
					}
					videoDataPayload = marshalVideoData(videoData)
					fmt.Println("SPS & PPS")
					SPS = nil
					PPS = nil
				} else {
					continue
				}
			} else {
				videoData := &VideoData{
					FrameType: FRAME_TYPE_KEY,
					CodecID: CODEC_AVC,
					AVCPacketType: AVC_NALU,
					CompositionTime: 0,
					Data: packet.Payload,
				}
				videoDataPayload = marshalVideoData(videoData)
			}

			flvTag := &FlvTag{
				TagType: TAG_VIDEO,
				DataSize: uint32(len(videoDataPayload)),
				Timestamp: uint32(packet.Timestamp - firstTimestamp) / 90,
				Data: videoDataPayload,
			}
			flvTagPayload := marshalFlvTag(flvTag)
			fmt.Println(uint32(packet.Timestamp - firstTimestamp) / 90)

			muxer.OutputChan <-flvTagPayload
			tagSize := make([]byte, 4)
			binary.BigEndian.PutUint32(tagSize, uint32(len(flvTagPayload)))
			muxer.OutputChan <-tagSize
		}
	}()

	return muxer
}

var FlvHeader []byte = []byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00}

type FlvTag struct {
	TagType uint8
	DataSize uint32
	Timestamp uint32
	Data []byte
}

const (
	TAG_AUDIO = 8
	TAG_VIDEO = 9
	TAG_SCRIPT = 18
)

type VideoData struct {
	FrameType uint8
	CodecID uint8
	AVCPacketType uint8
	CompositionTime int32
	Data []byte
}

type AVCDecoderConfigurationRecord struct {
	ConfigurationVersion uint8
	AVCProfileIndication uint8
	ProfileCompatibility uint8
	AVCLevelIndication uint8
	SPS []byte
	PPS []byte
}

const (
	FRAME_TYPE_KEY = 1
	FRAME_TYPE_INTER = 2
	FRAME_TYPE_DISP_INTER = 3
	FRAME_TYPE_GEN_INTER = 4
	FRAME_TYPE_INFO = 5
)

const (
	CODEC_JPEG = 1
	CODEC_H263 = 2
	CODEC_SCREEN = 3
	CODEC_VP6 = 4
	CODEC_VP6_ALPHA = 5
	CODEC_SCREEN2 = 6
	CODEC_AVC = 7
)

const (
	AVC_SEQ_HEADER = 0
	AVC_NALU = 1
	AVC_SEQ_END = 2
)

func marshalVideoData(videoData *VideoData) []byte {
	writer := bytes.NewBuffer([]byte{})

	binary.Write(writer, binary.BigEndian, (videoData.FrameType << 4) | videoData.CodecID)
	binary.Write(writer, binary.BigEndian, int32(0) | (int32(videoData.AVCPacketType) << 24) | videoData.CompositionTime)
	if videoData.AVCPacketType == AVC_NALU {
		binary.Write(writer, binary.BigEndian, uint32(len(videoData.Data)))
	}
	writer.Write(videoData.Data)

	return writer.Bytes()
}

func marshalFlvTag(flvTag *FlvTag) []byte {
	writer := bytes.NewBuffer([]byte{})

	binary.Write(writer, binary.BigEndian, uint32(0) | (uint32(flvTag.TagType) << 24) | flvTag.DataSize)
	binary.Write(writer, binary.BigEndian, flvTag.Timestamp << 8)
	writer.Write([]byte{0, 0, 0})
	writer.Write(flvTag.Data)

	return writer.Bytes()
}

func marshalAVCDecoderConfigurationRecord(record *AVCDecoderConfigurationRecord) []byte {
	writer := bytes.NewBuffer([]byte{})

	binary.Write(writer, binary.BigEndian, record.ConfigurationVersion)
	binary.Write(writer, binary.BigEndian, record.AVCProfileIndication)
	binary.Write(writer, binary.BigEndian, record.ProfileCompatibility)
	binary.Write(writer, binary.BigEndian, record.AVCLevelIndication)
	binary.Write(writer, binary.BigEndian, uint8(0xff))
	binary.Write(writer, binary.BigEndian, uint8(0xe1))
	binary.Write(writer, binary.BigEndian, uint16(len(record.SPS)))
	writer.Write(record.SPS)
	binary.Write(writer, binary.BigEndian, uint8(0x01))
	binary.Write(writer, binary.BigEndian, uint16(len(record.PPS)))
	writer.Write(record.PPS)

	return writer.Bytes()
}
