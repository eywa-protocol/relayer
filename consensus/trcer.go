package consensus

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/sirupsen/logrus"
)

func NewTracer() pubsub.EventTracer {
	return &tracer{}
}

type tracer struct{}

func (tr *tracer) Trace(evt *pubsub_pb.TraceEvent) {
	switch evt.Type {
	case pubsub_pb.TraceEvent_ADD_PEER.Enum():
		t := evt.GetAddPeer()
		logrus.Tracef("tracer add peer[%s] : %s", t.GetPeerID(), string(evt.GetPeerID()))
	case pubsub_pb.TraceEvent_REMOVE_PEER.Enum():
		t := evt.GetRemovePeer()
		logrus.Tracef("tracer remove peer[%s] : %s", t.GetPeerID(), string(evt.GetPeerID()))
	case pubsub_pb.TraceEvent_JOIN.Enum():
		t := evt.GetJoin()
		logrus.Tracef("tracer peer[%s] join topic: %s", t.GetTopic(), string(evt.GetPeerID()))
	case pubsub_pb.TraceEvent_LEAVE.Enum():
		t := evt.GetLeave()
		logrus.Tracef("tracer peer[%s] leave topic: %s", t.GetTopic(), string(evt.GetPeerID()))
	}
}
