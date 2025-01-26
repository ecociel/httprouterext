package httprouterext

import (
	"context"
	"log"
)

type User interface {
	Principal() string
}

type user struct {
	ns        Namespace
	obj       Obj
	principal Principal
	ctx       context.Context
	check     func(ctx context.Context, ns Namespace, obj Obj, permission Permission, userId UserId) (principal Principal, ok bool, err error)
}

func (u *user) Principal() string {
	return string(u.principal)
}

func (u *user) HasPermission(args ...string) bool {
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
		log.Printf("user check: %s %s %s: %v", ns, obj, permission, err)
	}
	return ok
}
