package grpcserver

import (
	"DNSPulse_ConfigHub/pkg/datastore"
	pb "DNSPulse_ConfigHub/pkg/gRPC"
	"DNSPulse_ConfigHub/pkg/logger"
	"context"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        logger.Logger.Errorf("authInterceptor: metadata is not provided, err: %v", codes.Unauthenticated)
        return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
    }

    tokens, ok := md["authorization"]
    if !ok || len(tokens) == 0 {
        logger.Logger.Errorf("authInterceptor: authorization token is not provided, err: %v", codes.Unauthenticated)
        return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
    }
    token := tokens[0]

	segmentName := extractSegmentName(req)
    if segmentName == "" {
        logger.Logger.Errorf("authInterceptor: segment name is not provided, err: %v", codes.InvalidArgument)
        return nil, status.Errorf(codes.InvalidArgument, "segment name is not provided")
    }

    if !isValidToken(token, segmentName) {
        logger.Logger.Errorf("authInterceptor: invalid token, err: %v", codes.Unauthenticated)
        return nil, status.Errorf(codes.Unauthenticated, "invalid token")
    }
    return handler(ctx, req)
}

func isValidToken(clToken string, segmentName string) bool {
    gRPCSrvToken := datastore.GetConfig().GRPCServer.Token
	srvToken := datastore.GetSegmentsSyncConf(segmentName).Token
	if srvToken == "" {
        return false
    }
    return clToken == gRPCSrvToken
}

func extractSegmentName(req interface{}) string {
    val := reflect.ValueOf(req)
    if val.Kind() == reflect.Ptr && !val.IsNil() {
        val = val.Elem()
    }
    if val.Kind() == reflect.Struct {
        switch req := req.(type) {
        case *pb.GetSegmentConfigRequest:
            return req.SegmentName
        case *pb.GetCsvRequest:
            return req.SegmentName
        default:
            return ""
        }
    }
    return ""
}