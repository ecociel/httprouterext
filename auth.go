package httprouterext

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	proto "github.com/ecociel/httprouterext/proto"
	"google.golang.org/grpc"
	"time"
)

var (
	ErrEmptyPrincipal = errors.New("unexpected empty principal")
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

type Timestamp string

func (s Timestamp) String() string {
	return string(s)
}

func TimestampEpoch() Timestamp {
	return Timestamp("1:0000000000000")
}

type Client struct {
	grpcClient   proto.CheckServiceClient
	observeCheck func(ns Namespace, obj Obj, permission Permission, userId UserId, duration time.Duration, ok bool, isError bool)
	observeList  func(ns Namespace, permission Permission, userId UserId, duration time.Duration, isError bool)
}

func New(conn *grpc.ClientConn) *Client {
	return &Client{
		grpcClient: proto.NewCheckServiceClient(conn),
	}
}

func (c *Client) WithObserveCheck(f func(ns Namespace, obj Obj, permission Permission, userId UserId, duration time.Duration, ok bool, isError bool)) *Client {
	c.observeCheck = f
	return c
}
func (c *Client) WithObserveList(f func(ns Namespace, permission Permission, userId UserId, duration time.Duration, isError bool)) *Client {
	c.observeList = f
	return c
}

func (c *Client) List(ctx context.Context, ns Namespace, permission Permission, userId UserId) ([]string, error) {
	begin := time.Now().UnixMilli()
	list, err := c.grpcClient.List(ctx, &proto.ListRequest{
		Ns:     string(ns),
		Rel:    string(permission),
		UserId: string(userId),
		Ts:     TimestampEpoch().String(),
	})
	elapsed := time.Now().UnixMilli() - begin
	if c.observeList != nil {
		c.observeList(ns, permission, userId, time.Duration(elapsed)*time.Millisecond, err != nil)
	}
	if err != nil {
		return nil, fmt.Errorf("list %s,%s,%s: %w", ns, permission, userId, err)
	}
	return list.Objs, nil
}

func (c *Client) Check(ctx context.Context, ns Namespace, obj Obj, permission Permission, userId UserId) (principal Principal, ok bool, err error) {
	return c.CheckWithTimestamp(ctx, ns, obj, permission, userId, Timestamp("1:0000000000000"))
}

func (c *Client) CheckWithTimestamp(ctx context.Context, ns Namespace, obj Obj, permission Permission, userId UserId, ts Timestamp) (principal Principal, ok bool, err error) {
	if permission == Impossible {
		return "", false, nil
	}
	begin := time.Now().UnixMilli()

	res, err := c.grpcClient.Check(ctx, &proto.CheckRequest{
		Ns:     string(ns),
		Obj:    string(obj),
		Rel:    string(permission),
		UserId: string(userId),
		Ts:     string(ts),
	})
	elapsed := time.Now().UnixMilli() - begin
	if c.observeCheck != nil {
		c.observeCheck(ns, obj, permission, userId, time.Duration(elapsed)*time.Millisecond, res.Ok, err != nil)
	}
	if err != nil {
		return "", false, err
	}
	if !res.Ok {
		if res.Principal != nil {
			return Principal((*res.Principal).Id), false, nil
		} else {
			return "", false, nil
		}
	} else {
		if res.Principal != nil {
			return Principal((*res.Principal).Id), true, nil
		} else {
			return "", false, ErrEmptyPrincipal
		}
	}
}

// NaiveBasicClient is a basic auth authenticator that holds a single
// username and password.
type NaiveBasicClient struct {
	username string
	password string
}

func NewNaiveBasicClient(username, password string) *NaiveBasicClient {
	return &NaiveBasicClient{
		username: username,
		password: password,
	}
}

func (c *NaiveBasicClient) Authenticate(_ context.Context, username, password []byte) (bool, error) {
	if string(username) != c.username {
		return false, nil
	}

	return subtle.ConstantTimeCompare(password, []byte(c.password)) == 1, nil
}
