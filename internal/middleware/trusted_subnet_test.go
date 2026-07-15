package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func handlerOK() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestTrustedSubnetMiddleware_EmptySubnet_AllowsAny(t *testing.T) {
	mw, err := TrustedSubnetMiddleware("")
	require.NoError(t, err)

	h := mw(handlerOK())

	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestTrustedSubnetMiddleware_IPInSubnet_Allowed(t *testing.T) {
	mw, err := TrustedSubnetMiddleware("192.168.1.0/24")
	require.NoError(t, err)

	h := mw(handlerOK())

	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.42")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestTrustedSubnetMiddleware_IPOutsideSubnet_Forbidden(t *testing.T) {
	mw, err := TrustedSubnetMiddleware("192.168.1.0/24")
	require.NoError(t, err)

	h := mw(handlerOK())

	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.5")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTrustedSubnetMiddleware_MissingHeader_Forbidden(t *testing.T) {
	mw, err := TrustedSubnetMiddleware("192.168.1.0/24")
	require.NoError(t, err)

	h := mw(handlerOK())

	req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestTrustedSubnetMiddleware_InvalidCIDR_ReturnsError(t *testing.T) {
	_, err := TrustedSubnetMiddleware("not-a-cidr")
	require.Error(t, err)
}
