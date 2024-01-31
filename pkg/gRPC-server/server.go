package grpcserver

import (
	"context"
	"net"

	"ConfigHub/pkg/datastore"
	pb "ConfigHub/pkg/gRPC"
	"ConfigHub/pkg/logger"

	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type server struct {
    pb.UnimplementedConfigHubServiceServer
}

func (s *server) GetSegmentConfig(ctx context.Context, req *pb.GetSegmentConfigRequest) (*pb.SegmentConfStruct, error) {
    // Validate the token
    if !isValidToken(req.Token, req.SegmentName) {
        return nil, status.Errorf(codes.Unauthenticated, "invalid token")
    }

    // Fetch the segment configuration from your datastore
    segmentConfig := datastore.GetSegmentsConfigBySEgment(req.SegmentName)
    if segmentConfig == (datastore.SegmentConfStruct{}) {
        return nil, status.Errorf(codes.Internal, "error retrieving segment config")
    }

    // Convert your Golang struct to protobuf message
    protoSegmentConfig := convertToProtoSegmentConfig(&segmentConfig)

    return protoSegmentConfig, nil
}

func convertToProtoSegmentConfig(segmentConfig *datastore.SegmentConfStruct) *pb.SegmentConfStruct {
    // Here you need to convert your Golang struct into the protobuf equivalent.
    // This is a dummy implementation. Replace it with actual field mappings.
    return &pb.SegmentConfStruct{
        SegmentName: segmentConfig.SegmentName,
        // Add other field mappings here...
    }
}

// func (s *server) GetCsv(ctx context.Context, in *pb.GetCsvRequest) (*pb.Csv, error) {
//     // Similar implementation as GetSegmentConfig
// }


func GrpcServer(port string) {
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
		logger.Logger.Fatalf("failed to listen: %v", err)
    }
	grpcServer :=grpc.NewServer(
        grpc.UnaryInterceptor(authInterceptor),
    )
    pb.RegisterConfigHubServiceServer(grpcServer, &server{})
    if err := grpcServer.Serve(lis); err != nil {
        logger.Logger.Fatalf("failed to serve: %v", err)
    }
}