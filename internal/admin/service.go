package admin

import (
	"context"
	"sort"

	genadmin "gRPCServerDemo/gen/admin"
	"google.golang.org/grpc"
)

// AdminServiceImpl implements the AdminService gRPC service.
type AdminServiceImpl struct {
	genadmin.UnimplementedAdminServiceServer
	server *grpc.Server
}

// NewAdminServiceImpl creates a new AdminServiceImpl with the given gRPC server reference.
func NewAdminServiceImpl(server *grpc.Server) *AdminServiceImpl {
	return &AdminServiceImpl{server: server}
}

// ListServices returns the full names of all registered gRPC services on the server.
func (s *AdminServiceImpl) ListServices(ctx context.Context, req *genadmin.ListServicesRequest) (*genadmin.ListServicesResponse, error) {
	serviceInfo := s.server.GetServiceInfo()
	serviceNames := make([]string, 0, len(serviceInfo))
	for name := range serviceInfo {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)
	return &genadmin.ListServicesResponse{ServiceNames: serviceNames}, nil
}