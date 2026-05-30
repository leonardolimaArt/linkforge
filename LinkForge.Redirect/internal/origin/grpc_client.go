package origin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	linkforgev1 "github.com/leonardolimaArt/linkforge/LinkForge.Redirect/gen/linkforge/v1"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultCallTimeout = 2 * time.Second
	apiKeyHeader       = "x-api-key"
)

type GrpcClient struct {
	conn    *grpc.ClientConn
	client  linkforgev1.LinkServiceClient
	timeout time.Duration
}

func NewGrpcClient(addr, apiKey string, timeout time.Duration) (*GrpcClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("origin: addr is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("origin: apiKey is required")
	}

	if timeout <= 0 {
		timeout = defaultCallTimeout
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(apiKeyInterceptor(apiKey)),
	)

	if err != nil {
		return nil, fmt.Errorf("origin: failed to dial %s: %w", addr, err)
	}

	return &GrpcClient{
		conn:    conn,
		client:  linkforgev1.NewLinkServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *GrpcClient) Resolve(ctx context.Context, shortCode string) (*domain.ShortLink, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.Resolve(callCtx, &linkforgev1.ResolveRequest{
		ShortCode: shortCode,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrOriginNotFound
		}
		return nil, fmt.Errorf("origni: resolve failed: %w", err)
	}

	return toDomain(resp), nil
}

func (c *GrpcClient) Close() error {
	return c.conn.Close()
}

func apiKeyInterceptor(apiKey string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, apiKeyHeader, apiKey)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func toDomain(resp *linkforgev1.ResolveResponse) *domain.ShortLink {
	var id uuid.UUID
	if rawID := resp.GetId(); rawID != "" {
		if parsed, err := uuid.Parse(rawID); err == nil {
			id = parsed
		}
	}

	var createdAt time.Time
	if ts := resp.GetCreatedAt(); ts != nil {
		createdAt = ts.AsTime()
	}
	return &domain.ShortLink{
		ID:          id,
		ShortCode:   resp.ShortCode,
		OriginalURL: resp.OriginalUrl,
		CreatedAt:   createdAt,
	}
}
