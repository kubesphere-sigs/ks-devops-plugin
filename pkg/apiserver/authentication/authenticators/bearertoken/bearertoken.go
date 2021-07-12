package bearertoken

import (
	"context"
	"k8s.io/apiserver/pkg/authentication/user"
	"time"

	jwt "devops.kubesphere.io/plugin/pkg/apiserver/authentication/token"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

// tokenAuthenticator implements an anonymous auth
type tokenAuthenticator struct{}

func New() authenticator.Token {
	return &tokenAuthenticator{}
}

func (a *tokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (response *authenticator.Response, ok bool, err error) {
	issuer := jwt.NewTokenIssuer("", time.Second)

	var authenticated user.Info
	if authenticated, _, err = issuer.VerifyWithoutClaimsValidation(token); err == nil {
		response = &authenticator.Response{
			User: &user.DefaultInfo{
				Name: authenticated.GetName(),
			}}
		ok = true
	} else {
		ok = false
	}
	return
}
