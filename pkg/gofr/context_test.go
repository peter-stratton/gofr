package gofr

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/peter-stratton/gofr/pkg/gofr/config"
	"github.com/peter-stratton/gofr/pkg/gofr/container"
	gofrHTTP "github.com/peter-stratton/gofr/pkg/gofr/http"
	"github.com/peter-stratton/gofr/pkg/gofr/logging"
)

func Test_newContextSuccess(t *testing.T) {
	httpRequest, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, "/test", bytes.NewBuffer([]byte(`{"key":"value"}`)))
	httpRequest.Header.Set("content-type", "application/json")

	if err != nil {
		t.Fatalf("unable to create request with context %v", err)
	}

	req := gofrHTTP.NewRequest(httpRequest)

	ctx := newContext(nil, req, container.NewContainer(config.NewEnvFile("",
		logging.NewMockLogger(logging.DEBUG))))

	body := map[string]string{}

	err = ctx.Bind(&body)

	assert.Equal(t, map[string]string{"key": "value"}, body, "TEST Failed \n unable to read body")
	assert.Nil(t, err, "TEST Failed \n unable to read body")
}
