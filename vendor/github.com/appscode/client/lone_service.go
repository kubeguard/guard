package client

import (
	"github.com/appscode/api/health"
	mailinglist "github.com/appscode/api/mailinglist/v1beta1"
	operation "github.com/appscode/api/operation/v1beta1"
	"google.golang.org/grpc"
)

// single client service in api. returned directly the api client.
type loneClientInterface interface {
	Health() health.HealthClient
	MailingList() mailinglist.MailingListClient
	Operation() operation.OperationsClient
}

type loneClientServices struct {
	healthClient      health.HealthClient
	mailingListClient mailinglist.MailingListClient
	operationClient   operation.OperationsClient
}

func newLoneClientService(conn *grpc.ClientConn) loneClientInterface {
	return &loneClientServices{
		healthClient:      health.NewHealthClient(conn),
		mailingListClient: mailinglist.NewMailingListClient(conn),
		operationClient:   operation.NewOperationsClient(conn),
	}
}

func (s *loneClientServices) Health() health.HealthClient {
	return s.healthClient
}

func (s *loneClientServices) MailingList() mailinglist.MailingListClient {
	return s.mailingListClient
}

func (s *loneClientServices) Operation() operation.OperationsClient {
	return s.operationClient
}
