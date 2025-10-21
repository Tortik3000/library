package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/project/library/config"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/outbox"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

const (
	MaxIdleConnections    = 100
	MaxConnectionsPerHost = 100
	IdleConnectionTimeout = 90 * time.Second
	TLSHandshakeTimeout   = 15 * time.Second
	ExpectContinueTimeout = 2 * time.Second
	Timeout               = 30 * time.Second
	KeepAlive             = 180 * time.Second
)

func runOutbox(
	ctx context.Context,
	cfg *config.Config,
	logger *zap.Logger,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) {
	dialer := &net.Dialer{
		Timeout:   Timeout,
		KeepAlive: KeepAlive,
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          MaxIdleConnections,
		MaxConnsPerHost:       MaxConnectionsPerHost,
		IdleConnTimeout:       IdleConnectionTimeout,
		TLSHandshakeTimeout:   TLSHandshakeTimeout,
		ExpectContinueTimeout: ExpectContinueTimeout,
	}

	client := &http.Client{Transport: transport}

	globalHandler := globalOutboxHandler(
		client, cfg.Outbox.BookSendURL, cfg.Outbox.AuthorSendURL)
	outboxService := outbox.New(
		logger, outboxRepository, globalHandler, cfg, transactor)

	outboxService.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.WaitTimeMS,
		cfg.Outbox.InProgressTTLMS,
	)
}

func globalOutboxHandler(
	client *http.Client,
	bookURL string,
	authorURL string,
) outbox.GlobalHandler {
	return func(kind repository.OutboxKind) (outbox.KindHandler, error) {
		switch kind {
		case repository.OutboxKindBook:
			return bookOutboxHandler(client, bookURL), nil
		case repository.OutboxKindAuthor:
			return authorOutboxHandler(client, authorURL), nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %d", kind)
		}
	}
}

func bookOutboxHandler(client *http.Client, url string) outbox.KindHandler {
	return func(ctx context.Context, data []byte) error {
		book := entity.Book{}
		err := json.Unmarshal(data, &book)

		if err != nil {
			return fmt.Errorf("can not deserialize data in book outbox handler: %w", err)
		}

		return send(ctx, client, []byte(book.ID), url)
	}
}

func authorOutboxHandler(client *http.Client, url string) outbox.KindHandler {
	return func(ctx context.Context, data []byte) error {
		author := entity.Author{}
		err := json.Unmarshal(data, &author)

		if err != nil {
			return fmt.Errorf("can not deserialize data in book outbox handler: %w", err)
		}

		return send(ctx, client, []byte(author.ID), url)
	}
}
func send(ctx context.Context, client *http.Client, body []byte, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if !checkSuccessResponse(resp) {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("non success response: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func checkSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
}
