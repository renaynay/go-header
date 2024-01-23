package p2p

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("header/p2p")

const (
	failedKey           = "failed"
	headerReceivedKey   = "num_headers_received"
	headTypeKey         = "request_type"
	headTypeTrusted     = "trusted_request"
	headTypeUntrusted   = "untrusted_request"
	headStatusKey       = "status"
	headStatusOk        = "ok"
	headStatusTimeout   = "timeout"
	headStatusCanceled  = "canceled"
	headStatusNoHeaders = "no_headers"
)

type exchangeMetrics struct {
	headRequestTimeInst metric.Float64Histogram
	responseSizeInst    metric.Int64Histogram
	responseTimeInst    metric.Float64Histogram
}

func newExchangeMetrics() (m *exchangeMetrics, err error) {
	m = new(exchangeMetrics)
	m.headRequestTimeInst, err = meter.Float64Histogram(
		"hdr_p2p_exch_clnt_head_time_hist",
		metric.WithDescription("exchange client head request time in seconds"),
	)
	if err != nil {
		return nil, err
	}
	m.responseSizeInst, err = meter.Int64Histogram(
		"hdr_p2p_exch_clnt_resp_size_hist",
		metric.WithDescription("exchange client header response size in bytes"),
	)
	if err != nil {
		return nil, err
	}
	m.responseTimeInst, err = meter.Float64Histogram(
		"hdr_p2p_exch_clnt_resp_time_hist",
		metric.WithDescription("exchange client response time in seconds"),
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *exchangeMetrics) head(ctx context.Context, duration time.Duration, headersReceived int, tp, status string) {
	m.observe(ctx, func(ctx context.Context) {
		m.headRequestTimeInst.Record(ctx,
			duration.Seconds(),
			metric.WithAttributes(
				attribute.Int(headerReceivedKey, headersReceived),
				attribute.String(headTypeKey, tp),
				attribute.String(headStatusKey, status),
			),
		)
	})
}

func (m *exchangeMetrics) response(ctx context.Context, size uint64, duration time.Duration, err error) {
	m.observe(ctx, func(ctx context.Context) {
		m.responseSizeInst.Record(ctx,
			int64(size),
			metric.WithAttributes(attribute.Bool(failedKey, err != nil)),
		)
		m.responseTimeInst.Record(ctx,
			duration.Seconds(),
			metric.WithAttributes(attribute.Bool(failedKey, err != nil)),
		)
	})
}

func (m *exchangeMetrics) observe(ctx context.Context, observeFn func(context.Context)) {
	if m == nil {
		return
	}
	if ctx.Err() != nil {
		ctx = context.Background()
	}

	observeFn(ctx)
}
