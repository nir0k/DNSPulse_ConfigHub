package grpcserver

import (
	"ConfigHub/pkg/datastore"
	pb "ConfigHub/pkg/gRPC"
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
        return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
    }

    // Assuming the token is passed in the "authorization" metadata
    tokens, ok := md["authorization"]
    if !ok || len(tokens) == 0 {
        return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
    }
    token := tokens[0]

	segmentName := extractSegmentName(req)
    if segmentName == "" {
        return nil, status.Errorf(codes.InvalidArgument, "segment name is not provided")
    }

    // Validate the token
    if !isValidToken(token, segmentName) {
        return nil, status.Errorf(codes.Unauthenticated, "invalid token")
    }
    // Continue processing the request
    return handler(ctx, req)
}

func isValidToken(clToken string, segmentName string) bool {
	srvToken := datastore.GetSegmentsSyncConf(segmentName).Token
	if srvToken == "" {
        return false
    }
    return clToken == srvToken
}

func extractSegmentName(req interface{}) string {
    // Use reflection to determine the type of the request and extract segmentName
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