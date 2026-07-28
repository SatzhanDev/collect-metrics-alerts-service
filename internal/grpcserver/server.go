package grpcserver

import (
	"context"

	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/logger"
	models "github.com/SatzhanDev/collect-metrics-alerts-service/internal/models"
	pb "github.com/SatzhanDev/collect-metrics-alerts-service/internal/proto"
	"github.com/SatzhanDev/collect-metrics-alerts-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	svc service.Service
}

func NewMetricsServer(svc service.Service) *MetricsServer {
	return &MetricsServer{svc: svc}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	batch := make([]models.Metrics, 0, len(req.GetMetrics()))

	for _, m := range req.GetMetrics() {
		metric := models.Metrics{ID: m.GetId()}

		switch m.GetType() {
		case pb.Metric_GAUGE:
			metric.MType = models.Gauge
			v := m.GetValue()
			metric.Value = &v
		case pb.Metric_COUNTER:
			metric.MType = models.Counter
			d := m.GetDelta()
			metric.Delta = &d
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type for id %q", m.GetId())
		}

		batch = append(batch, metric)
	}

	if err := s.svc.UpdateBatch(ctx, batch); err != nil {
		logger.Log.Error("update batch failed", zap.String("component", "grpc"), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update metrics")
	}

	return &pb.UpdateMetricsResponse{}, nil
}
