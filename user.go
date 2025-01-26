package httprouterext

import (
	"context"
	"fmt"
	"log"
)

type User interface {
	Principal() string
	HasPermission(args ...string) (bool, error)
	List(ns string, permission string) ([]string, error)
}

type user struct {
	ns        Namespace
	obj       Obj
	principal Principal
	ctx       context.Context
	check     func(ctx context.Context, ns Namespace, obj Obj, permission Permission, userId UserId) (principal Principal, ok bool, err error)
	list      func(ctx context.Context, ns Namespace, permission Permission, userId UserId) ([]string, error)
}

func (u *user) Principal() string {
	return string(u.principal)
}

func (u *user) HasPermission(args ...string) (bool, error) {
	var ns Namespace
	var obj Obj
	var permission Permission

	switch len(args) {
	case 1:
		ns = u.ns
		obj = u.obj
		permission = Permission(args[0])
		break
	case 2:
		obj = Obj(args[0])
		permission = Permission(args[1])
		break
	case 3:
		ns = Namespace(args[0])
		obj = Obj(args[1])
		permission = Permission(args[2])
		break
	default:
		panic("HasPermission requires 1 or 3 arguments")
	}
	log.Printf("user check2: %s %s %s", ns, obj, permission)
	_, ok, err := u.check(u.ctx, ns, obj, permission, UserId(u.principal))
	if err != nil {
		return false, fmt.Errorf("user check: %s %s %s: %w", ns, obj, permission, err)
	}
	return ok, nil
}

func (u *user) List(ns string, permission string) ([]string, error) {
	log.Printf("list: %s %s", ns, permission)
	objs, err := u.list(u.ctx, Namespace(ns), Permission(permission), UserId(u.principal))
	if err != nil {
		return nil, fmt.Errorf("list: %s %s: %w", ns, permission, err)
	}
	return objs, nil
}
