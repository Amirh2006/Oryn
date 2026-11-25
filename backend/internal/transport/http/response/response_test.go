package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Amirh2006/Oryn/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("RequestID", "test-req-123")

	payload := map[string]string{"title": "Oryn"}
	JSON(c, http.StatusOK, payload)

	assert.Equal(t, http.StatusOK, w.Code)

	var res Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)

	assert.NotNil(t, res.Data)
	assert.Equal(t, "test-req-123", res.Meta.RequestID)
	assert.Nil(t, res.Error)
}

func TestErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("RequestID", "req-err-456")

	Error(c, http.StatusNotFound, domain.ErrCodeNotFound, "movie not found", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)

	assert.Nil(t, res.Data)
	assert.Equal(t, "req-err-456", res.Meta.RequestID)
	assert.Equal(t, domain.ErrCodeNotFound, res.Error.Code)
	assert.Equal(t, "movie not found", res.Error.Message)
}

func TestHandleError_DomainMapping(t *testing.T) {
	tests := []struct {
		err          error
		expectedCode int
	}{
		{domain.ErrNotFound, http.StatusNotFound},
		{domain.ErrUnauthorized, http.StatusUnauthorized},
		{domain.ErrForbidden, http.StatusForbidden},
		{domain.ErrRateLimited, http.StatusTooManyRequests},
		{domain.ErrProviderFailed, http.StatusBadGateway},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		HandleError(c, tc.err)
		assert.Equal(t, tc.expectedCode, w.Code)
	}
}
