package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/words/words"
)

const maxPhraseLen = 4096

type ServerConfig struct {
	Port string `yaml:"port" env:"WORDS_GRPC_PORT" env-default:"8080"`
}

type Server struct {
	wordspb.UnimplementedWordsServer
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *Server) Norm(ctx context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {

	if len(in.Phrase) > maxPhraseLen {
		return nil, status.Error(codes.ResourceExhausted, "phrase too large")
	}

	stemmedWords := words.Norm(in.Phrase)

	return &wordspb.WordsReply{Words: stemmedWords}, nil
}

func parseServerConfig(configPath string, addrFlag string) (string, string, error) {
	var cfg ServerConfig

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return "", "", err
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return "", "", err
	}

	port := cfg.Port

	if addrFlag == "" {
		addrFlag = fmt.Sprintf(":%s", port)
	}

	return addrFlag, port, nil
}

func main() {
	configPath := flag.String("config", "config.yaml", "config path")
	addrFlag := flag.String("address", "", "server address")

	flag.Parse()

	address, port, err := parseServerConfig(*configPath, *addrFlag)

	if err != nil {
		log.Fatalf("Error parsing server config: %v", err)
	}
	slog.Info("server config",
		"address", address,
		"port", port,
	)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &Server{})
	reflection.Register(s)

	go func() {
		if err := s.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()

	slog.Info("shutting down")

	s.GracefulStop()
}
