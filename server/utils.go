package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/appscode/go/log"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	auth "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// write replies to the request with the specified TokenReview object and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
func write(w http.ResponseWriter, info *auth.UserInfo, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-content-type-options", "nosniff")

	resp := auth.TokenReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auth.SchemeGroupVersion.String(),
			Kind:       "TokenReview",
		},
	}

	if err != nil {
		code := http.StatusUnauthorized
		if v, ok := err.(httpStatusCode); ok {
			code = v.Code()
		}
		printStackTrace(err)
		w.WriteHeader(code)
		resp.Status = auth.TokenReviewStatus{
			Authenticated: false,
			Error:         err.Error(),
		}
	} else {
		w.WriteHeader(http.StatusOK)
		resp.Status = auth.TokenReviewStatus{
			Authenticated: true,
			User:          *info,
		}
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		panic(err)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type httpStatusCode interface {
	Code() int
}

func printStackTrace(err error) {
	c, ok := errors.Cause(err).(stackTracer)
	if !ok {
		panic("oops, err does not implement stackTracer")
	}

	st := c.StackTrace()
	log.Errorf("%s\nStacktrace: %+v", err.Error(), st) // top two frames
}

func GetSupportedOrg() []string {
	return []string{
		"Github",
		"Gitlab",
		"Google",
	}
}

// output form : Github/Google/Gitlab
func SupportedOrgPrintForm() string {
	return strings.Join(GetSupportedOrg(), "/")
}

// WithCode annotates err with a new code.
// If err is nil, WithCode returns nil.
func WithCode(err error, code int) error {
	if err == nil {
		return nil
	}
	return &withCode{
		cause: err,
		code:  code,
	}
}

type withCode struct {
	cause error
	code  int
}

func (w *withCode) Error() string { return w.cause.Error() }
func (w *withCode) Cause() error  { return w.cause }
func (w *withCode) Code() int     { return w.code }

func (w *withCode) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
