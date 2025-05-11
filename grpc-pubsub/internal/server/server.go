package server

import (
	"context"

	pb "github.com/vgbhj/vk-techno-test/proto"

	"github.com/vgbhj/vk-techno-test/subpub"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	pb.UnimplementedPubSubServer
	broker subpub.SubPub
	logger *zap.Logger
}

func New(broker subpub.SubPub, logger *zap.Logger) *Service {
	return &Service{broker: broker, logger: logger}
}

func (s *Service) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	key := req.GetKey()
	s.logger.Info("new subscription", zap.String("key", key))
	sub, err := s.broker.Subscribe(key, func(msg interface{}) {
		ev := &pb.Event{Data: msg.(string)}
		if err := stream.Send(ev); err != nil {
			s.logger.Warn("send to client failed", zap.Error(err))
		}
	})
	if err != nil {
		return status.Errorf(codes.Internal, "subscribe error: %v", err)
	}
	<-stream.Context().Done()
	sub.Unsubscribe()
	s.logger.Info("subscription closed", zap.String("key", key))
	return nil
}

func (s *Service) Publish(ctx context.Context, req *pb.PublishRequest) (*emptypb.Empty, error) {
	key := req.GetKey()
	data := req.GetData()
	s.logger.Info("publish request", zap.String("key", key), zap.String("data", data))
	if err := s.broker.Publish(key, data); err != nil {
		return nil, status.Errorf(codes.Internal, "publish error: %v", err)
	}
	return &emptypb.Empty{}, nil
}
