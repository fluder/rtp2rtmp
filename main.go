package main

import (
	"./muxers"
)

func main() {


	videoUdpSource := muxers.NewUdpSource(5000)
	videoRtpDemuxer := muxers.NewRtpDemuxer()
	videoRtpH264Demuxer := muxers.NewRtpH264Depacketizer()

	flvMuxer := muxers.NewFlvMuxer()
	rtmpSink := muxers.NewRtmpSink("rtmp://s1.transcoding.svoe.tv/gst", "test")

	muxers.Bridge(videoUdpSource.OutputChan, videoRtpDemuxer.InputChan)
	muxers.Bridge(videoRtpDemuxer.OutputChan, videoRtpH264Demuxer.InputChan)
	muxers.Bridge(videoRtpH264Demuxer.OutputChan, flvMuxer.InputVideoChan)

	audioUdpSource := muxers.NewUdpSource(5002)
	audioRtpDemuxer := muxers.NewRtpDemuxer()
	audioRtpMPESDepacketizer := muxers.NewRtpMPESDepacketizer()

	muxers.Bridge(audioUdpSource.OutputChan, audioRtpDemuxer.InputChan)
	muxers.Bridge(audioRtpDemuxer.OutputChan, audioRtpMPESDepacketizer.InputChan)
	muxers.Bridge(audioRtpMPESDepacketizer.OutputChan, flvMuxer.InputAudioChan)

	muxers.Bridge(flvMuxer.OutputChan, rtmpSink.InputChan)

	//go func() {
	//	f, _ := os.Create("/tmp/test.aac")
	//	//f.Write([]byte{0, 0, 0, 1})
	//	for {
	//		data := (<-audioRtpMPESDepacketizer.OutputChan).([]byte)
	//		f.Write(data)
	//		//f.Write([]byte{0, 0, 1})
	//	}
	//}()

	select {}
}
