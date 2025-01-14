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

type Client struct {
	conn   *grpc.ClientConn
	client proto.CheckServiceClient
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		client: proto.NewCheckServiceClient(conn),
	}
}

func (c *Client) Check(ctx context.Context, ns Namespace, obj Obj, permission Permission, userId UserId) (principal Principal, ok bool, err error) {

	res, err := c.client.Check(ctx, &proto.CheckRequest{
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
