package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leonardolimaArt/linkforge/LinkForge.Redirect/internal/service"
	"github.com/stretchr/testify/assert"
)

//Mock

type mockResolver struct {
	resolveFn func(ctx context.Context, code string) (string, error)
}

func (m *mockResolver) Resolve(ctx context.Context, code string) (string, error) {
	return m.resolveFn(ctx, code)
}

func newTestRouter(h *RedirectHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/r/:shortCode", h.Handler)
	return r
}

//Test

func TestRedirectHandler_Sucesss(t *testing.T) {
	resolver := &mockResolver{
		resolveFn: func(ctx context.Context, code string) (string, error) {
			return "https://google.com", nil
		},
	}

	h := NewRedirectHandler(resolver)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://google.com", w.Header().Get("Location"))
}

func TestRedirectHandler_NotFound(t *testing.T) {
	resolver := &mockResolver{
		resolveFn: func(ctx context.Context, code string) (string, error) {
			return "", service.ErrNotFound
		},
	}

	h := NewRedirectHandler(resolver)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, w.Header().Get("Location"))
}

func TestRedirectHandler_InternalError(t *testing.T) {
	resolver := &mockResolver{
		resolveFn: func(ctx context.Context, code string) (string, error) {
			return "", errors.New("database down")
		},
	}

	h := NewRedirectHandler(resolver)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/r/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
