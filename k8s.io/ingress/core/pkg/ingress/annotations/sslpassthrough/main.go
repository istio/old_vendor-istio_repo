/*
Copyright 2016 The Kubernetes Authors.

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

package sslpassthrough

import (
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/ingress/core/pkg/ingress/annotations/parser"
	ing_errors "k8s.io/ingress/core/pkg/ingress/errors"
)

const (
	passthrough = "ingress.kubernetes.io/ssl-passthrough"
)

type sslpt struct {
}

// NewParser creates a new SSL passthrough annotation parser
func NewParser() parser.IngressAnnotation {
	return sslpt{}
}

// ParseAnnotations parses the annotations contained in the ingress
// rule used to indicate if is required to configure
func (a sslpt) Parse(ing *extensions.Ingress) (interface{}, error) {
	if ing.GetAnnotations() == nil {
		return false, ing_errors.ErrMissingAnnotations
	}

	return parser.GetBoolAnnotation(passthrough, ing)
}
