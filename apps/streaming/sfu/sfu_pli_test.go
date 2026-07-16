package sfu

import (
	"testing"
	"time"

	"github.com/pion/rtcp"
)

func Test_SFU_SubscribeExisting_RequestsPublisherKeyframe(t *testing.T) {
	s := NewSFU()
	_, publisherPC, publisherTrack := newLoopbackClient(t, s, "call1", "publisher")
	defer publisherPC.Close()
	stop := make(chan struct{})
	go writeVP8Loop(publisherTrack, stop)
	defer close(stop)
	deadline := time.Now().Add(5 * time.Second)
	for len(s.room("call1").PublishedExcept("subscriber")) == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if len(s.room("call1").PublishedExcept("subscriber")) == 0 {
		t.Fatal("publisher track was not registered")
	}

	_, subscriberPC, _ := newLoopbackClient(t, s, "call1", "subscriber")
	defer subscriberPC.Close()

	pli := make(chan struct{}, 1)
	go func() {
		for {
			packets, _, err := publisherPC.GetSenders()[0].ReadRTCP()
			if err != nil {
				return
			}
			for _, packet := range packets {
				if _, ok := packet.(*rtcp.PictureLossIndication); ok {
					pli <- struct{}{}
					return
				}
			}
		}
	}()

	s.SubscribeExisting("call1", "subscriber")

	select {
	case <-pli:
	case <-time.After(5 * time.Second):
		t.Fatal("publisher did not receive PLI when video subscription was created")
	}
}
