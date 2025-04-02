package library

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CheckError(t *testing.T, err error, code codes.Code) {
	t.Helper()
	s, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, code, s.Code())
}
