package entity

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Author struct {
	ID   string
	Name string
}

var (
	ErrAuthorNotFound = status.Error(codes.NotFound, "author not found")
)
