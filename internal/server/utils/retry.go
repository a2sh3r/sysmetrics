package utils

import (
	"context"
	"database/sql"
	"errors"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"io"
	"net"
	"os"
	"syscall"
	"time"
)

type RetriableFunc func() error

func WithRetries(fn RetriableFunc) error {
	retries := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	var lastErr error
	for _, wait := range retries {
		if err := fn(); err != nil {
			logger.Log.Error("retriable error", zap.Error(err), zap.Duration("duration", wait))
			lastErr = err
			if !IsRetriableError(err) {
				return err
			}
			time.Sleep(wait)
		} else {
			return nil
		}
	}

	return lastErr
}

func IsRetriableError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown:
			return true
		}
	}

	if errors.Is(err, sql.ErrConnDone) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			return true
		}
	}

	if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
		return true
	}

	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	return false
}
