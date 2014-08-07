package muxers

import (
	"github.com/zhangpeihao/gortmp"
	"github.com/zhangpeihao/log"
	"fmt"
	"time"
)

type RtmpSink struct {
	InputChan chan interface {}
}

func NewRtmpSink(url, name string) *RtmpSink {
	sink := &RtmpSink{
		InputChan: make(chan interface {}),
	}

	go func() {
		var err error

		l := log.NewLogger("logs", "publisher", nil, 60, 3600*24, true)
		rtmp.InitLogger(l)

		handler := &RtmpSinkHandler{}
		handler.createStreamChan = make(chan rtmp.OutboundStream)
		handler.flvChan = sink.InputChan

		handler.obConn, err = rtmp.Dial(url, handler, 100)
		if err != nil {
			fmt.Println("Rtmp dial error", err)
			return
		}
		err = handler.obConn.Connect()
		if err != nil {
			fmt.Printf("Connect error: %s", err.Error())
			return
		}

		for {
			select {
			case stream := <-handler.createStreamChan:
				// Publish
				stream.Attach(handler)
				err = stream.Publish(name, "live")
				if err != nil {
					fmt.Printf("Publish error: %s", err.Error())
					return
				}

			case <-time.After(1 * time.Second):
				fmt.Printf("Audio size: %d bytes; Video size: %d bytes\n", handler.audioDataSize, handler.videoDataSize)
			}
		}
	}()

	return sink
}

type RtmpSinkHandler struct {
	status uint
	obConn rtmp.OutboundConn
	createStreamChan chan rtmp.OutboundStream
	videoDataSize int64
	audioDataSize int64
	flvChan chan interface {}
}

func (handler *RtmpSinkHandler) OnStatus(conn rtmp.OutboundConn) {
	var err error
	handler.status, err = handler.obConn.Status()
	fmt.Printf("@@@@@@@@@@@@@status: %d, err: %v\n", handler.status, err)
}

func (handler *RtmpSinkHandler) OnClosed(conn rtmp.Conn) {
	fmt.Printf("@@@@@@@@@@@@@Closed\n")
}

func (handler *RtmpSinkHandler) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
}

func (handler *RtmpSinkHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	fmt.Printf("ReceviedRtmpCommand: %+v\n", command)
}

func (handler *RtmpSinkHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	fmt.Printf("Stream created: %d\n", stream.ID())
	handler.createStreamChan <- stream
}
func (handler *RtmpSinkHandler) OnPlayStart(stream rtmp.OutboundStream) {

}
func (handler *RtmpSinkHandler) OnPublishStart(stream rtmp.OutboundStream) {
	// Set chunk buffer size
	go func() {
		fmt.Println("Publish")
		for {
			flvTag := (<-handler.flvChan).(*FlvTag)
			fmt.Println("Push data")
			stream.PublishVideoData(flvTag.Data, flvTag.Timestamp)
		}
	}()
}
