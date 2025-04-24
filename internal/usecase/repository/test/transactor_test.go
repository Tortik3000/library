package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/project/library/internal/usecase/repository"
)

func TestWithTx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operation      func(ctx context.Context) error
		mockErr        error
		wantErr        bool
		expectRollback bool
	}{
		{
			name: "successful transaction",
			operation: func(ctx context.Context) error {
				return nil
			},
			mockErr:        nil,
			wantErr:        false,
			expectRollback: false,
		},
		{
			name: "operation error",
			operation: func(ctx context.Context) error {
				return fmt.Errorf("operation failed")
			},
			mockErr:        nil,
			wantErr:        true,
			expectRollback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockDB, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockDB.Close()

			logger, _ := zap.NewProduction()
			mockTransactor := repository.NewTransactor(mockDB, logger)
			ctx := t.Context()

			mockDB.ExpectBegin()
			if tt.expectRollback {
				mockDB.ExpectRollback()
			} else {
				mockDB.ExpectCommit()
			}

			err = mockTransactor.WithTx(ctx, tt.operation)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mockDB.ExpectationsWereMet())
		})
	}
}
