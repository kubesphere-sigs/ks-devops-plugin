/*

 Copyright 2020 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/

package auth

import (
	"fmt"
	"net/mail"

	"golang.org/x/crypto/bcrypt"

	informers "devops.kubesphere.io/plugin/pkg/client/informers/externalversions"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	authoptions "devops.kubesphere.io/plugin/pkg/apiserver/authentication/options"
	kubesphere "devops.kubesphere.io/plugin/pkg/client/clientset/versioned"
	iamv1alpha2listers "devops.kubesphere.io/plugin/pkg/client/listers/iam/v1alpha2"
)

var (
	RateLimitExceededError  = fmt.Errorf("auth rate limit exceeded")
	IncorrectPasswordError  = fmt.Errorf("incorrect password")
	AccountIsNotActiveError = fmt.Errorf("account is not active")
)

type PasswordAuthenticator interface {
	Authenticate(username, password string) (authuser.Info, string, error)
}

type OAuthAuthenticator interface {
	Authenticate(provider, code string) (authuser.Info, string, error)
}

type passwordAuthenticator struct {
	ksClient    kubesphere.Interface
	userGetter  *userGetter
	authOptions *authoptions.AuthenticationOptions
}

type oauth2Authenticator struct {
	ksClient    kubesphere.Interface
	userGetter  *userGetter
	authOptions *authoptions.AuthenticationOptions
}

type userGetter struct {
	userLister iamv1alpha2listers.UserLister
}

func NewPasswordAuthenticator(ksClient kubesphere.Interface,
	userLister iamv1alpha2listers.UserLister,
	options *authoptions.AuthenticationOptions) PasswordAuthenticator {
	passwordAuthenticator := &passwordAuthenticator{
		ksClient:    ksClient,
		userGetter:  &userGetter{userLister: userLister},
		authOptions: options,
	}
	return passwordAuthenticator
}

func NewOAuthAuthenticator(ksClient kubesphere.Interface,
	ksInformer informers.SharedInformerFactory,
	options *authoptions.AuthenticationOptions) OAuthAuthenticator {
	oauth2Authenticator := &oauth2Authenticator{
		ksClient:    ksClient,
		userGetter:  &userGetter{userLister: ksInformer.Iam().V1alpha2().Users().Lister()},
		authOptions: options,
	}
	return oauth2Authenticator
}

func (p *passwordAuthenticator) Authenticate(username, password string) (authuser.Info, string, error) {
	// empty username or password are not allowed
	if username == "" || password == "" {
		return nil, "", IncorrectPasswordError
	}

	// kubesphere account
	user, err := p.userGetter.findUser(username)
	if err != nil {
		// ignore not found error
		if !errors.IsNotFound(err) {
			klog.Error(err)
			return nil, "", err
		}
	}

	// check user status
	if user != nil && (user.Status.State == nil || *user.Status.State != iamv1alpha2.UserActive) {
		if user.Status.State != nil && *user.Status.State == iamv1alpha2.UserAuthLimitExceeded {
			klog.Errorf("%s, username: %s", RateLimitExceededError, username)
			return nil, "", RateLimitExceededError
		} else {
			// state not active
			klog.Errorf("%s, username: %s", AccountIsNotActiveError, username)
			return nil, "", AccountIsNotActiveError
		}
	}

	// if the password is not empty, means that the password has been reset, even if the user was mapping from IDP
	if user != nil && user.Spec.EncryptedPassword != "" {
		if err = PasswordVerify(user.Spec.EncryptedPassword, password); err != nil {
			klog.Error(err)
			return nil, "", err
		}
		u := &authuser.DefaultInfo{
			Name: user.Name,
		}
		// check if the password is initialized
		if uninitialized := user.Annotations[iamv1alpha2.UninitializedAnnotation]; uninitialized != "" {
			u.Extra = map[string][]string{
				iamv1alpha2.ExtraUninitialized: {uninitialized},
			}
		}
		return u, "", nil
	}

	return nil, "", IncorrectPasswordError
}

func (o oauth2Authenticator) Authenticate(provider, code string) (authuser.Info, string, error) {
	return nil, "", errors.NewNotFound(iamv1alpha2.Resource("user"), "authenticated.GetUsername()")
}

func PasswordVerify(encryptedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password)); err != nil {
		return IncorrectPasswordError
	}
	return nil
}

// findUser
func (u *userGetter) findUser(username string) (*iamv1alpha2.User, error) {
	if _, err := mail.ParseAddress(username); err != nil {
		return u.userLister.Get(username)
	} else {
		users, err := u.userLister.List(labels.Everything())
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		for _, find := range users {
			if find.Spec.Email == username {
				return find, nil
			}
		}
	}

	return nil, errors.NewNotFound(iamv1alpha2.Resource("user"), username)
}

func (u *userGetter) findLinkedAccount(idp, uid string) (*iamv1alpha2.User, error) {
	selector := labels.SelectorFromSet(labels.Set{
		iamv1alpha2.IdentifyProviderLabel: idp,
		iamv1alpha2.OriginUIDLabel:        uid,
	})

	users, err := u.userLister.List(selector)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	if len(users) != 1 {
		return nil, errors.NewNotFound(iamv1alpha2.Resource("user"), uid)
	}

	return users[0], err
}
