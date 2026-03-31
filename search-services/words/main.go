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

type server struct {
	wordspb.UnimplementedWordsServer
}

func (s *server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) Norm(ctx context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {

	if len(in.Phrase) > maxPhraseLen {
		return nil, status.Error(codes.ResourceExhausted, "phrase too large")
	}

	return &wordspb.WordsReply{Words: words.Norm(in.Phrase)}, nil
}

type Config struct {
	Address string `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"80"`
}

func parseConfig(configPath string) (Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error("error reading server config:", "error", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		slog.Error("error reading server env:", "error", err)
		return Config{}, err
	}

	return cfg, nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg, err := parseConfig(configPath)

	if err != nil {
		log.Fatalf("error parsing config: %v", err)
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{})
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
