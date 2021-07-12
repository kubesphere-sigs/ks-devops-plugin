package anonymous

import (
	"devops.kubesphere.io/plugin/pkg/apiserver/query"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"net/http"
)

// Authenticator implements an anonymous auth
type Authenticator struct{}

func NewAuthenticator() authenticator.Request {
	return &Authenticator{}
}

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	if auth := query.GetAuthorization(req.Header); auth == "" {
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name:   "anonymous",
				Groups: []string{user.AllAuthenticated},
			},
		}, true, nil
	}
	return nil, false, nil
}
