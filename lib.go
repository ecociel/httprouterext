package guard

import (
	"context"
	proto "github.com/ecociel/guard-go-client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Namespace string
type Obj string
type Permission string
type UserId string
type Principal string

type Resource interface {
	Requires(principalOrToken string, method string) (ns Namespace, obj Obj, permission Permission)
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

		return "", false, err
	}
	return Principal(res.Principal.Id), true, nil
}
func (c *Client) List(ctx context.Context, ns, permission, userId string) ([]Obj, error) {
	list, err := c.grpcClient.List(ctx, &proto.ListRequest{
		Ns:         ns,
		Permission: permission,
		UserId:     userId,
	})
	if err != nil {
		return nil, err
	}
	return toObjects(list.Obj), err
}

func toObjects(list []string) []Obj {
	var objects []Obj
	for _, l := range list {
		objects = append(objects, Obj(l))
	}
	return objects
}
