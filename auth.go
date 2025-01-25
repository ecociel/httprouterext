package httprouterext

import (
	"context"
	"fmt"
	proto "github.com/ecociel/guard-go-client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Namespace string

func (s Namespace) String() string {
	return string(s)
}

type Obj string

func (s Obj) String() string {
	return string(s)
}

type Permission string

func (s Permission) String() string {
	return string(s)
}

type UserId string

func (s UserId) String() string {
	return string(s)
}

type Principal string

func (s Principal) String() string {
	return string(s)
}

type Client struct {
	grpcClient proto.CheckServiceClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		grpcClient: proto.NewCheckServiceClient(conn),
	}
}

func (c *Client) Check(ctx context.Context, ns Namespace, obj Obj, permission Permission, userId UserId) (principal Principal, ok bool, err error) {
	if permission == Impossible {
		return "", false, nil
	}

	res, err := c.grpcClient.Check(ctx, &proto.CheckRequest{
		Ns:         string(ns),
		Obj:        string(obj),
		Permission: string(permission),
		UserId:     string(userId),
		//Timestamp:  &timestamp,
	})
	if err != nil {
		s := status.Convert(err)
		if s.Code() == codes.NotFound {
			return "", false, nil
		}

		return "", false, fmt.Errorf("check %s,%s,%s,%s: %w", ns, obj, permission, userId, err)
	}
	return Principal(res.Principal.Id), true, nil
}

func (c *Client) List(ctx context.Context, ns Namespace, permission Permission, userId UserId) ([]Obj, error) {
	list, err := c.grpcClient.List(ctx, &proto.ListRequest{
		Ns:         string(ns),
		Permission: string(permission),
		UserId:     string(userId),
	})
	if err != nil {
		return nil, fmt.Errorf("list %s,%s,%s: %w", ns, permission, userId, err)
	}
	result := make([]Obj, len(list.Obj))
	for i := range list.Obj {
		result[i] = Obj(list.Obj[i])
	}
	return result, nil
}
