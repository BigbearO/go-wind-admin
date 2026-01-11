package logging

import (
	"context"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
)

type WriteOperationLogFunc func(ctx context.Context, data *adminV1.OperationAuditLog) error
type WriteLoginLogFunc func(ctx context.Context, data *adminV1.LoginAuditLog) error

type options struct {
	writeOperationLogFunc WriteOperationLogFunc
	writeLoginLogFunc     WriteLoginLogFunc
}

type Option func(*options)

func WithWriteOperationLogFunc(fnc WriteOperationLogFunc) Option {
	return func(opts *options) {
		opts.writeOperationLogFunc = fnc
	}
}

func WithWriteLoginLogFunc(fnc WriteLoginLogFunc) Option {
	return func(opts *options) {
		opts.writeLoginLogFunc = fnc
	}
}
