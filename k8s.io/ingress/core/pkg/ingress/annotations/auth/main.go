/*
Copyright 2015 The Kubernetes Authors.

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
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/pkg/errors"
	api "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/file"
	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	ing_errors "k8s.io/ingress/core/pkg/ingress/errors"
	"k8s.io/ingress/core/pkg/ingress/resolver"
)

const (
	authType   = "ingress.kubernetes.io/auth-type"
	authSecret = "ingress.kubernetes.io/auth-secret"
	authRealm  = "ingress.kubernetes.io/auth-realm"
)

var (
	authTypeRegex = regexp.MustCompile(`basic|digest`)
	// AuthDirectory default directory used to store files
	// to authenticate request
	AuthDirectory = "/etc/ingress-controller/auth"
)

// BasicDigest returns authentication configuration for an Ingress rule
type BasicDigest struct {
	Type    string `json:"type"`
	Realm   string `json:"realm"`
	File    string `json:"file"`
	Secured bool   `json:"secured"`
	FileSHA string `json:"fileSha"`
}

// Equal tests for equality between two BasicDigest types
func (bd1 *BasicDigest) Equal(bd2 *BasicDigest) bool {
	if bd1 == bd2 {
		return true
	}
	if bd1 == nil || bd2 == nil {
		return false
	}
	if bd1.Type != bd2.Type {
		return false
	}
	if bd1.Realm != bd2.Realm {
		return false
	}
	if bd1.File != bd2.File {
		return false
	}
	if bd1.Secured != bd2.Secured {
		return false
	}
	if bd1.FileSHA != bd2.FileSHA {
		return false
	}

	return true
}

type auth struct {
	secretResolver resolver.Secret
	authDirectory  string
}

// NewParser creates a new authentication annotation parser
func NewParser(authDirectory string, sr resolver.Secret) parser.IngressAnnotation {
	os.MkdirAll(authDirectory, 0755)

	currPath := authDirectory
	for currPath != "/" {
		currPath = path.Dir(currPath)
		err := os.Chmod(currPath, 0755)
		if err != nil {
			break
		}
	}

	return auth{sr, authDirectory}
}

// Parse parses the annotations contained in the ingress
// rule used to add authentication in the paths defined in the rule
// and generated an htpasswd compatible file to be used as source
// during the authentication process
func (a auth) Parse(ing *extensions.Ingress) (interface{}, error) {
	at, err := parser.GetStringAnnotation(authType, ing)
	if err != nil {
		return nil, err
	}

	if !authTypeRegex.MatchString(at) {
		return nil, ing_errors.NewLocationDenied("invalid authentication type")
	}

	s, err := parser.GetStringAnnotation(authSecret, ing)
	if err != nil {
		return nil, ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "error reading secret name from annotation"),
		}
	}

	name := fmt.Sprintf("%v/%v", ing.Namespace, s)
	secret, err := a.secretResolver.GetSecret(name)
	if err != nil {
		return nil, ing_errors.LocationDenied{
			Reason: errors.Wrapf(err, "unexpected error reading secret %v", name),
		}
	}

	realm, _ := parser.GetStringAnnotation(authRealm, ing)

	passFile := fmt.Sprintf("%v/%v-%v.passwd", a.authDirectory, ing.GetNamespace(), ing.GetName())
	err = dumpSecret(passFile, secret)
	if err != nil {
		return nil, err
	}

	return &BasicDigest{
		Type:    at,
		Realm:   realm,
		File:    passFile,
		Secured: true,
		FileSHA: file.SHA1(passFile),
	}, nil
}

// dumpSecret dumps the content of a secret into a file
// in the expected format for the specified authorization
func dumpSecret(filename string, secret *api.Secret) error {
	val, ok := secret.Data["auth"]
	if !ok {
		return ing_errors.LocationDenied{
			Reason: errors.Errorf("the secret %v does not contain a key with value auth", secret.Name),
		}
	}

	// TODO: check permissions required
	err := ioutil.WriteFile(filename, val, 0777)
	if err != nil {
		return ing_errors.LocationDenied{
			Reason: errors.Wrap(err, "unexpected error creating password file"),
		}
	}

	return nil
}
