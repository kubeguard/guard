package client

import (
	"appscode.com/api/health"
	operation "appscode.com/api/operation/v1alpha1"
	"google.golang.org/grpc"
)

// single client service in api. returned directly the api client.
type loneClientInterface interface {
	Health() health.HealthClient
	Operation() operation.OperationsClient
}

type loneClientServices struct {
	healthClient    health.HealthClient
	operationClient operation.OperationsClient
}

func newLoneClientService(conn *grpc.ClientConn) loneClientInterface {
	return &loneClientServices{
		healthClient:    health.NewHealthClient(conn),
		operationClient: operation.NewOperationsClient(conn),
	}
}

func (s *loneClientServices) Health() health.HealthClient {
	return s.healthClient
}

func (s *loneClientServices) Operation() operation.OperationsClient {
	return s.operationClient
}
