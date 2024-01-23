package p2p

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/metric"
)

type trackerMetrics struct {
	clientReg metric.Registration

	trackedPeersNumInst      metric.Int64ObservableGauge
	disconnectedPeersNumInst metric.Int64ObservableGauge

	blockedPeersNum     atomic.Int64
	blockedPeersNumInst metric.Int64ObservableGauge
}

func (p *peerTracker) withMetrics() error {
	m := &trackerMetrics{}

	var err error
	m.trackedPeersNumInst, err = meter.Int64ObservableGauge(
		"hdr_p2p_peer_trck_peer_num_gauge",
		metric.WithDescription("peer tracker tracked peers number"),
	)
	if err != nil {
		return err
	}
	m.disconnectedPeersNumInst, err = meter.Int64ObservableGauge(
		"hdr_p2p_peer_disconn_peer_num_gauge",
		metric.WithDescription("peer tracker tracked disconnected peers number"),
	)
	if err != nil {
		return err
	}
	m.blockedPeersNumInst, err = meter.Int64ObservableGauge(
		"hdr_p2p_peer_block_peer_num_gauge",
		metric.WithDescription("peer tracker blocked peers number"),
	)
	if err != nil {
		return err
	}

	m.clientReg, err = meter.RegisterCallback(
		p.observeMetrics,
		m.trackedPeersNumInst,
		m.disconnectedPeersNumInst,
		m.blockedPeersNumInst,
	)
	if err != nil {
		return err
	}

	p.metrics = m
	return nil
}

func (tm *trackerMetrics) Close() (err error) {
	if tm == nil {
		return nil
	}
	return tm.clientReg.Unregister()
}

func (tm *trackerMetrics) observe(ctx context.Context, observeFn func(context.Context)) {
	if tm == nil {
		return
	}
	if ctx.Err() != nil {
		ctx = context.Background()
	}

	observeFn(ctx)
}

func (tm *trackerMetrics) peerBlocked() {
	tm.observe(context.Background(), func(ctx context.Context) {
		tm.blockedPeersNum.Add(1)
	})
}

func (p *peerTracker) observeMetrics(_ context.Context, obs metric.Observer) error {
	obs.ObserveInt64(p.metrics.trackedPeersNumInst, p.numTracked())
	obs.ObserveInt64(p.metrics.disconnectedPeersNumInst, p.numDisconnected())
	obs.ObserveInt64(p.metrics.blockedPeersNumInst, p.metrics.blockedPeersNum.Load())
	return nil
}

func (p *peerTracker) numTracked() int64 {
	p.peerLk.RLock()
	defer p.peerLk.RUnlock()
	return int64(len(p.trackedPeers))
}

func (p *peerTracker) numDisconnected() int64 {
	p.peerLk.RLock()
	defer p.peerLk.RUnlock()
	return int64(len(p.disconnectedPeers))
}
