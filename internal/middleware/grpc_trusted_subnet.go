package middleware

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GRPCTrustedSubnetInterceptor(trustedSubnet string) (grpc.UnaryServerInterceptor, error) {
	var subnet *net.IPNet

	if trustedSubnet != "" {
		_, parsed, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("parse trusted subnet %q: %w", trustedSubnet, err)
		}
		subnet = parsed
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if subnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		values := md.Get("x-real-ip")
		if len(values) == 0 {
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip metadata")
		}

		ip := net.ParseIP(values[0])
		if ip == nil || !subnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "ip address is not in trusted subnet")
		}

		return handler(ctx, req)
	}, nil
}
