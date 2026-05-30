package integration

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	linkforgev1 "github.com/leonardolimaArt/linkforge/LinkForge.Redirect/gen/linkforge/v1"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/origin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeLinkServer struct {
	linkforgev1.UnimplementedLinkServiceServer
	handle func(code string) (*linkforgev1.ResolveResponse, error)
}

func (s *fakeLinkServer) Resolve(_ context.Context, req *linkforgev1.ResolveRequest) (*linkforgev1.ResolveResponse, error) {
	return s.handle(req.GetShortCode())
}

func startFakeServer(t *testing.T, fake *fakeLinkServer) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	linkforgev1.RegisterLinkServiceServer(srv, fake)

	go func() {
		_ = srv.Serve(lis)
	}()

	t.Cleanup(func() {
		srv.GracefulStop()
	})

	return lis.Addr().String()
}

func TestIntegration_GrpcFallback_ResolveSuccess(t *testing.T) {
	expectedID := uuid.New()
	fake := &fakeLinkServer{
		handle: func(code string) (*linkforgev1.ResolveResponse, error) {
			return &linkforgev1.ResolveResponse{
				Id:          expectedID.String(),
				ShortCode:   code,
				OriginalUrl: "https://google.com/from-origin",
				CreatedAt:   timestamppb.Now(),
			}, nil
		},
	}

	addr := startFakeServer(t, fake)

	client, err := origin.NewGrpcClient(addr, "any-key", 2*time.Second)
	require.NoError(t, err)
	defer client.Close()

	link, err := client.Resolve(context.Background(), "abc12345")

	require.NoError(t, err)
	assert.Equal(t, expectedID, link.ID)
	assert.Equal(t, "abc12345", link.ShortCode)
	assert.Equal(t, "https://google.com/from-origin", link.OriginalURL)
}

func TestIntegration_GrpcFallback_NotFoundMapsToErrOriginNotFound(t *testing.T) {
	fake := &fakeLinkServer{
		handle: func(code string) (*linkforgev1.ResolveResponse, error) {
			return nil, status.Error(codes.NotFound, "short code not found")
		},
	}

	addr := startFakeServer(t, fake)

	client, err := origin.NewGrpcClient(addr, "any-key", 2*time.Second)
	require.NoError(t, err)
	defer client.Close()

	_, err = client.Resolve(context.Background(), "missing")

	require.Error(t, err)
	assert.ErrorIs(t, err, origin.ErrOriginNotFound)
}
