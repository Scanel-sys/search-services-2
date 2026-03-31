package words

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	log    *slog.Logger
	client wordspb.WordsClient
	conn   *grpc.ClientConn
}

func NewClient(address string, log *slog.Logger) (*Client, error) {

	log.Info("server config",
		"address", address,
	)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Error("failed to connect to words service", "error", err)
		return nil, err
	}

	client := wordspb.NewWordsClient(conn)

	return &Client{
		log:    log,
		client: client,
		conn:   conn,
	}, nil
}

func (c Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return err
	}

	return nil
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	request := &wordspb.WordsRequest{Phrase: phrase}
	reply, err := c.client.Norm(ctx, request)

	if err != nil {
		if status.Code(err) == codes.ResourceExhausted {
			c.log.Error("too long message received", "error", err)
			return nil, err
		}

		c.log.Error("error calling Norm function", "error", err)
		return nil, err
	}

	return reply.Words, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	return err
}
