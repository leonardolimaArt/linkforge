package origin

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	linkforgev1 "github.com/leonardolimaArt/linkforge/LinkForge.Redirect/gen/linkforge/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToDomain_CompleteResponse_MapsAllFields(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	resp := &linkforgev1.ResolveResponse{
		Id:          id.String(),
		ShortCode:   "abcd1234",
		OriginalUrl: "https://google.com",
		CreatedAt:   timestamppb.New(createdAt),
	}

	link := toDomain(resp)

	assert.Equal(t, id, link.ID)
	assert.Equal(t, "abcd1234", link.ShortCode)
	assert.Equal(t, "https://google.com", link.OriginalURL)
	assert.True(t, createdAt.Equal(link.CreatedAt))
}

func TestToDomain_EmptyId_ReturnsUUIDNil(t *testing.T) {
	resp := &linkforgev1.ResolveResponse{
		Id:          "",
		ShortCode:   "abcd1234",
		OriginalUrl: "https://google.com",
		CreatedAt:   timestamppb.New(time.Now()),
	}
	link := toDomain(resp)

	assert.Equal(t, uuid.Nil, link.ID)
	assert.Equal(t, "abcd1234", link.ShortCode)
	assert.Equal(t, "https://google.com", link.OriginalURL)
}

func TestToDomain_InvalidId_ReturnsUUIDNil(t *testing.T) {
	resp := &linkforgev1.ResolveResponse{
		Id:          "not-a-uuid-maybe-or-it-is",
		ShortCode:   "abcd1234",
		OriginalUrl: "https://google.com",
		CreatedAt:   timestamppb.New(time.Now()),
	}

	link := toDomain(resp)

	assert.Equal(t, uuid.Nil, link.ID, "invalid UUID should fallback to uuid.Nil instead of panicking")
}

func TestToDomain_NilCreatedAt_ReturnsZeroTime(t *testing.T) {
	resp := &linkforgev1.ResolveResponse{
		Id:          uuid.New().String(),
		ShortCode:   "abcd1234",
		OriginalUrl: "https://google.com",
		CreatedAt:   nil,
	}

	link := toDomain(resp)
	assert.True(t, link.CreatedAt.IsZero(), "nil Timestamp should map to zero time.Time")
}

func TestApiKeyInterceptor_InjectsHeaderIntoOutgoingContext(t *testing.T) {
	const expectedKey = "my-secret-api-key-super-cool"
	interceptor := apiKeyInterceptor(expectedKey)
	var capturedMetadata metadata.MD
	mockInvoker := func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		capturedMetadata = md
		return nil
	}

	err := interceptor(
		context.Background(),
		"/linkforge.v1.LinkService/Resolve",
		nil,
		nil,
		nil,
		mockInvoker,
	)

	assert.NoError(t, err)
	values := capturedMetadata.Get(apiKeyHeader)
	assert.Len(t, values, 1)
	assert.Equal(t, expectedKey, values[0])
}

func TestApiKeyInterceptor_PropagatesInvokerError(t *testing.T) {
	interceptor := apiKeyInterceptor("any-key-hahahahahah")

	expectedErr := assert.AnError
	mockInvoker := func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		return expectedErr
	}

	err := interceptor(
		context.Background(),
		"/linkforge.v1.LinkService/Resolve",
		nil,
		nil,
		nil,
		mockInvoker,
	)
	assert.ErrorIs(t, err, expectedErr)
}

// mim ajuda
