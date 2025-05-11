package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/vgbhj/vk-techno-test/proto"

	"github.com/vgbhj/vk-techno-test/subpub"

	"github.com/vgbhj/vk-techno-test/internal/config"
	"github.com/vgbhj/vk-techno-test/internal/log"
	"github.com/vgbhj/vk-techno-test/internal/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(err)
	}

	logger, err := log.New(cfg.Log.Level)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	broker := subpub.NewSubPub()

	grpcSrv := grpc.NewServer()
	svc := server.New(broker, logger)
	pb.RegisterPubSubServer(grpcSrv, svc)

	addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	logger.Info("gRPC server starting", zap.String("addr", addr))

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Fatal("serve error", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown signal received")

	grpcSrv.GracefulStop()
	_ = broker.Close(context.Background())
	logger.Info("server exited")
}
