package main

import (
	"context"
	"errors"
	"flag"
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

const maxPhraseLen = 20000

type ServerConfig struct {
	Port string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"80"`
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

func parseServerConfig(configPath string) (string, error) {
	var cfg ServerConfig

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("error reading server config:", "error", err)
	} else {
		return cfg.Port, nil
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		slog.Error("error reading server env:", "error", err)
		return "", err
	}

	return cfg.Port, nil
}

func main() {
	configPath := flag.String("config", "config.yaml", "config path")

	flag.Parse()

	port, err := parseServerConfig(*configPath)

	if err != nil {
		log.Fatalf("Error parsing server config: %v", err)
	}
	slog.Info("server config",
		"address", port,
	)

	listener, err := net.Listen("tcp", port)
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
