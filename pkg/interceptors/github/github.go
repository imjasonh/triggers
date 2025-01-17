/*
Copyright 2019 The Tekton Authors

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

package github

import (
	"errors"
	"net/http"

	gh "github.com/google/go-github/github"

	"github.com/tektoncd/triggers/pkg/interceptors"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type Interceptor struct {
	KubeClientSet          kubernetes.Interface
	Logger                 *zap.SugaredLogger
	Github                 *triggersv1.GithubInterceptor
	EventListenerNamespace string
}

func NewInterceptor(gh *triggersv1.GithubInterceptor, k kubernetes.Interface, ns string, l *zap.SugaredLogger) interceptors.Interceptor {
	return &Interceptor{
		Logger:                 l,
		Github:                 gh,
		KubeClientSet:          k,
		EventListenerNamespace: ns,
	}
}

func (w *Interceptor) ExecuteTrigger(payload []byte, request *http.Request, _ *triggersv1.EventListenerTrigger, _ string) ([]byte, error) {
	// No secret set, just continue
	if w.Github.SecretRef == nil {
		return payload, nil
	}
	header := request.Header.Get("X-Hub-Signature")
	if header == "" {
		return nil, errors.New("no X-Hub-Signature header set")
	}

	secretToken, err := interceptors.GetSecretToken(w.KubeClientSet, w.Github.SecretRef, w.EventListenerNamespace)
	if err != nil {
		return nil, err
	}
	if err := gh.ValidateSignature(header, payload, secretToken); err != nil {
		return nil, err
	}

	return payload, nil
}
