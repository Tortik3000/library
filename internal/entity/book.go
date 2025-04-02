package entity

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Book struct {
	ID        string
	Name      string
	AuthorIDs []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrBookNotFound = status.Error(codes.NotFound, "book not found")
)
