package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ssov1 "github.com/Braendie/protos/gen/go/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api ssov1.AuthClient
	log *slog.Logger
}

func New(
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("New grpc client", slog.String("address", addr))

	return &Client{
		api: ssov1.NewAuthClient(cc),
	}, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "grpc.IsAdmin"

	resp, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return resp.IsAdmin, nil
}


// func (c *Client) Register(ctx context.Context, email, password string) (int64, error) {
// 	const op = "grpc.Register"

// 	resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{
// 		Email: email,
// 		Password: password,
// 	})
// 	if err != nil {
// 		return 0, fmt.Errorf("%s: %w", op, err)
// 	}

// 	return resp.UserId, nil
// }

// func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
// 	const op = "grpc.Login"

// 	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{
// 		Email: email,
// 		Password: password,
// 		AppId: 1,
// 	})
// 	if err != nil {
// 		return "", fmt.Errorf("%s: %w", op, err)
// 	}

// 	return resp.Token, nil
// }

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
