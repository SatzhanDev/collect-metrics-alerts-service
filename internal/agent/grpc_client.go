package agent

import (
	"context"
	"fmt"
	"time"

	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	pb "github.com/SatzhanDev/collect-metrics-alerts-service/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCSender struct {
	client pb.MetricsClient
	conn   *grpc.ClientConn
	realIP string
}

func NewGRPCSender(addr string) (*GRPCSender, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", addr, err)
	}

	return &GRPCSender{
		client: pb.NewMetricsClient(conn),
		conn:   conn,
		realIP: outboundIP(), // та же функция, что уже использует HTTPSender
	}, nil
}

func (s *GRPCSender) Close() error {
	return s.conn.Close()
}

func (s *GRPCSender) withRealIP(ctx context.Context) context.Context {
	if s.realIP == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-real-ip", s.realIP)
}

func toProtoMetric(m models.Metrics) *pb.Metric {
	metric := &pb.Metric{Id: m.ID}

	switch m.MType {
	case models.Gauge:
		metric.Type = pb.Metric_GAUGE
		if m.Value != nil {
			metric.Value = *m.Value
		}
	case models.Counter:
		metric.Type = pb.Metric_COUNTER
		if m.Delta != nil {
			metric.Delta = *m.Delta
		}
	}

	return metric
}

func (s *GRPCSender) send(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	req := &pb.UpdateMetricsRequest{
		Metrics: make([]*pb.Metric, 0, len(metrics)),
	}
	for _, m := range metrics {
		req.Metrics = append(req.Metrics, toProtoMetric(m))
	}

	callCtx, cancel := context.WithTimeout(s.withRealIP(ctx), 5*time.Second)
	defer cancel()

	_, err := s.client.UpdateMetrics(callCtx, req)
	return err
}

func (s *GRPCSender) SendBatch(ctx context.Context, metrics []models.Metrics) error {
	return s.send(ctx, metrics)
}

func (s *GRPCSender) SendGauge(ctx context.Context, name string, value float64) error {
	v := value
	return s.send(ctx, []models.Metrics{{ID: name, MType: models.Gauge, Value: &v}})
}

func (s *GRPCSender) SendCounter(ctx context.Context, name string, value int64) error {
	d := value
	return s.send(ctx, []models.Metrics{{ID: name, MType: models.Counter, Delta: &d}})
}

func (s *GRPCSender) SendGaugeJSON(ctx context.Context, name string, value float64) error {
	return s.SendGauge(ctx, name, value)
}

func (s *GRPCSender) SendCounterJSON(ctx context.Context, name string, value int64) error {
	return s.SendCounter(ctx, name, value)
}
