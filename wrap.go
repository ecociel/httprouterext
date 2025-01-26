package httprouterext

import (
	"context"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Resource interface {
	Requires(principalOrToken string, method string) (ns Namespace, obj Obj, permission Permission)
}

type responseWriterWrapper struct {
	http.ResponseWriter
	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
	userAgent             string
	headersSent           bool
}

func Observe(meter Meter, w http.ResponseWriter, r *http.Request, f func(w http.ResponseWriter) error) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	rw := &responseWriterWrapper{
		ResponseWriter: w,
		ip:             clientIP,
		time:           time.Time{},
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		elapsedTime:    time.Duration(0),
		userAgent:      r.UserAgent(),
	}
	startTime := time.Now()
	err := f(rw)
	finishTime := time.Now()
	rw.time = finishTime.UTC()
	rw.elapsedTime = finishTime.Sub(startTime)

	if err != nil {
		errMsg := mapError(meter, err, rw, r)
		if errMsg != "" {
			log.Printf("%s %s: error=%s identity=%s duration=%s", r.Method, r.RequestURI, errMsg, "-", rw.elapsedTime.String())
		}
	}
}

func mapError(meter Meter, err error, w *responseWriterWrapper, req *http.Request) (errMsg string) {
	// If the handler returns an error, we try our best here
	// to map it to more than just Internal Server Error
	// First, if the error contains a hint to the caller, we
	// report that hint to the caller using generic 400 status code.
	// Also, we do not log these kinds of errors since given there

	//cause := cerr.Cause(err)
	//
	//if errors.Is(cause, cerr.ErrNotFound) {
	//	h.Respond404(w, cause.Error())
	//	mx.ResponseErrors.WithLabelValues("404", obj.OperationId(req)).Add(1)
	//	return
	//}
	//switch cause.(type) {
	//case h.Problemer:
	//	h.RespondProblem(w, cause.(h.Problemer).Problem())
	//	mx.ResponseErrors.WithLabelValues(strconv.Itoa(cause.(h.Problemer).Problem().Status), obj.OperationId(req)).Add(1)
	//	return
	//case cerr.Hinter:
	//	// Ignore potential wrapping as hinters are for conveying original
	//	// cause + hint to the user
	//	h.Respond400(w, fmt.Sprintf("%s: %v", cause.(cerr.Hinter).Hint(), err))
	//	mx.ResponseErrors.WithLabelValues("400", obj.OperationId(req)).Add(1)
	//	return
	//default:
	//	Respond500WithError(w, err)
	//	mx.ResponseErrors.WithLabelValues("500", obj.OperationId(req)).Add(1)
	//	errMsg = fmt.Sprintf("%v", err)
	//	return errMsg
	//}

	http.Error(w, err.Error(), http.StatusInternalServerError)
	errMsg = fmt.Sprintf("%v", err)
	return errMsg
}

type HandlerFunc func(http.ResponseWriter, *http.Request, httprouter.Params, Resource, User) error

type Meter interface {
}
type Wrapper interface {
	Meter
	Check(ctx context.Context, ns Namespace, obj Obj, permission Permission, userId UserId) (principal Principal, ok bool, err error)
	List(ctx context.Context, ns Namespace, permission Permission, userId UserId) ([]string, error)
}

// TODO const None = Permission("none")
const Impossible = Permission("impossible")

func Wrap(wrapper Wrapper, extract func(r *http.Request, p httprouter.Params) (Resource, error), hdl HandlerFunc) httprouter.Handle {
	return httprouter.Handle(func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {

		sessionCookie, err := r.Cookie("session")
		if errors.Is(err, http.ErrNoCookie) {
			back := url.QueryEscape(r.RequestURI)
			uri := fmt.Sprintf("/signin?back=%s", back)
			http.Redirect(rw, r, uri, http.StatusSeeOther)
			return
		}
		token := sessionCookie.Value

		Observe(wrapper, rw, r, func(w http.ResponseWriter) error {
			resource, err := extract(r, p)
			if err != nil {
				return fmt.Errorf("extract: %w", err)
			}
			ns, obj, permission := resource.Requires(token, r.Method)
			fmt.Printf("Access - %s,%s,%s\n", ns, obj, permission)

			principal, ok, err := wrapper.Check(r.Context(), ns, obj, permission, UserId(token))
			if err != nil {
				return fmt.Errorf("check: %w", err)
			}
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				return nil
			}

			user := user{
				ns:        ns,
				obj:       obj,
				principal: principal,
				ctx:       r.Context(),
				check:     wrapper.Check,
				list:      wrapper.List,
			}

			return hdl(w, r, p, resource, &user)
		})
	})
}
