package grpcserver

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"DNSPulse_ConfigHub/pkg/datastore"
	pb "DNSPulse_ConfigHub/pkg/gRPC"
	"DNSPulse_ConfigHub/pkg/logger"
	"DNSPulse_ConfigHub/pkg/tools"

	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type server struct {
    pb.UnimplementedConfigHubServiceServer
}

func (s *server) GetSegmentConfig(ctx context.Context, req *pb.GetSegmentConfigRequest) (*pb.SegmentConfStruct, error) {
    if !isValidToken(req.Token, req.SegmentName) {
        return nil, status.Errorf(codes.Unauthenticated, "invalid token")
    }
    segmentConfig, ok := datastore.GetSegmentsConfigBySegment(req.SegmentName)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "segment config not found for: %s", req.SegmentName)
	}
    protoSegmentConfig := convertToProtoSegmentConfig(segmentConfig)

    return protoSegmentConfig, nil
}

func convertToProtoSegmentConfig(segmentConfig datastore.SegmentConfStruct) *pb.SegmentConfStruct {

    return &pb.SegmentConfStruct{
        SegmentName: segmentConfig.SegmentName,
        General: &pb.GeneralConfig{
            CheckInterval: int32(segmentConfig.General.CheckInterval),
            Hash:          segmentConfig.General.Hash,
        },
        Sync: &pb.SyncConfig{
            IsEnabled: segmentConfig.Sync.IsEnable,
            Token:     segmentConfig.Sync.Token,
        },
        Prometheus: &pb.PrometheusConfig{
            Url:         segmentConfig.Prometheus.URL,
            AuthEnabled: segmentConfig.Prometheus.AuthEnabled,
            Username:    segmentConfig.Prometheus.Username,
            Password:    segmentConfig.Prometheus.Password,
            MetricName:  segmentConfig.Prometheus.MetricName,
            RetriesCount: int32(segmentConfig.Prometheus.RetriesCount),
            BufferSize:  int32(segmentConfig.Prometheus.BufferSize),
            Labels: &pb.PrometheusLabelConfig{
                Opcode:             segmentConfig.Prometheus.Labels.Opcode,
                Authoritative:      segmentConfig.Prometheus.Labels.Authoritative,
                Truncated:          segmentConfig.Prometheus.Labels.Truncated,
                Rcode:              segmentConfig.Prometheus.Labels.Rcode,
                RecursionDesired:   segmentConfig.Prometheus.Labels.RecursionDesired,
                RecursionAvailable: segmentConfig.Prometheus.Labels.RecursionAvailable,
                AuthenticatedData:  segmentConfig.Prometheus.Labels.AuthenticatedData,
                CheckingDisabled:   segmentConfig.Prometheus.Labels.CheckingDisabled,
                PollingRate:        segmentConfig.Prometheus.Labels.PollingRate,
                Recursion:          segmentConfig.Prometheus.Labels.Recursion,
            },
        },
        Polling: &pb.PollingConfig{
            Path:          segmentConfig.Polling.Path,
            Hash:          segmentConfig.Polling.Hash,
            Delimiter:     segmentConfig.Polling.Delimeter,
            ExtraDelimiter:segmentConfig.Polling.ExtraDelimeter,
            PollTimeout:   int32(segmentConfig.Polling.PollTimeout),
        },
    }
}

func (s *server) GetCsv(ctx context.Context, req *pb.GetCsvRequest) (*pb.CsvList, error) {
    if !isValidToken(req.Token, req.SegmentName) {
        return nil, status.Errorf(codes.Unauthenticated, "invalid token")
    }
    csvData, ok := datastore.GetPollingHostsBySegment(req.SegmentName)
    if !ok {
        return nil, status.Errorf(codes.NotFound, "CSV data not found for segment: %s", req.SegmentName)
    }
    protoCsvList, err := convertToProtoCsv(csvData)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "error converting CSV data: %v", err)
    }

    return protoCsvList, nil
}

func convertToProtoCsv(csvData []datastore.Csv) (*pb.CsvList, error) {
    protoCsvList := &pb.CsvList{}
    for _, item := range csvData {
        protoCsv := &pb.Csv{
            Server:                  item.Server,
            IpAddress:               item.IPAddress,
            Domain:                  item.Domain,
            Location:                item.Location,
            Site:                    item.Site,
            ServerSecurityZone:      item.ServerSecurityZone,
            Prefix:                  item.Prefix,
            Protocol:                item.Protocol,
            Zonename:                item.Zonename,
            QueryCount:              item.QueryCount,
            ZonenameWithRecursion:   item.ZonenameWithRecursion,
            QueryCountWithRecursion: item.QueryCountWithRecursion,
            ServiceMode:             item.ServiceMode,
        }
        protoCsvList.Csvs = append(protoCsvList.Csvs, protoCsv)
    }
    return protoCsvList, nil
}

func StartGRPCServer() {
    port := datastore.GetConfig().GRPCServer.Port
    if !tools.CheckPortAvailability(port) {
		logger.Logger.Errorf("Port is already in use. Cannot start the gRPC server. Port: %d\n", port)
        return
    }
    grpcServer(strconv.Itoa(port))
}

func grpcServer(port string) {
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
		logger.Logger.Fatalf("failed to listen: %v", err)
    }
	grpcServer :=grpc.NewServer(
        grpc.UnaryInterceptor(authInterceptor),
    )
    fmt.Printf("gRPC Server starting on port: %s\n", port)
    logger.Logger.Infof("gRPC Server starting on port %s", port)
    pb.RegisterConfigHubServiceServer(grpcServer, &server{})
    if err := grpcServer.Serve(lis); err != nil {
        logger.Logger.Fatalf("failed to serve: %v", err)
    }
}
