package admin

import (
	"context"
	"sort"

	pb "gRPCServerDemo/gen/admin"

	"google.golang.org/grpc"
)

// Handler implements the AdminService gRPC service.
type Handler struct {
	pb.UnimplementedAdminServiceServer
	server *grpc.Server
}

// NewHandler creates a new Handler with the given gRPC server reference.
func NewHandler(server *grpc.Server) *Handler {
	return &Handler{server: server}
}

// ListServices returns the full names of all registered gRPC services on the server.
func (h *Handler) ListServices(ctx context.Context, req *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
	serviceInfo := h.server.GetServiceInfo()
	serviceNames := make([]string, 0, len(serviceInfo))
	for name := range serviceInfo {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)
	return &pb.ListServicesResponse{ServiceNames: serviceNames}, nil
}
