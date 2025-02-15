/*
Copyright 2019 The Knative Authors
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

// This file contains an object which encapsulates k8s clients which are useful for e2e tests.

package test

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"knative.dev/pkg/test"
	"knative.dev/serving-operator/pkg/client/clientset/versioned"
	servingv1alpha1 "knative.dev/serving-operator/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

// Clients holds instances of interfaces for making requests to Knative Serving.
type Clients struct {
	KubeClient                *test.KubeClient
	Dynamic                   dynamic.Interface
	KnativeServingAlphaClient servingv1alpha1.KnativeServingInterface
	Config                    *rest.Config
}

// NewClients instantiates and returns several clientsets required for making request to the
// Knative Serving cluster specified by the combination of clusterName and configPath. Clients can
// make requests within namespace.
func NewClients(configPath string, clusterName string, namespace string) (*Clients, error) {
	clients := &Clients{}
	cfg, err := buildClientConfig(configPath, clusterName)
	if err != nil {
		return nil, err
	}

	// We poll, so set our limits high.
	cfg.QPS = 100
	cfg.Burst = 200

	clients.KubeClient, err = test.NewKubeClient(configPath, clusterName)
	if err != nil {
		return nil, err
	}

	clients.Dynamic, err = dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	clients.KnativeServingAlphaClient, err = newKnativeServingAlphaClients(cfg, namespace)
	if err != nil {
		return nil, err
	}

	clients.Config = cfg
	return clients, nil
}

func buildClientConfig(kubeConfigPath string, clusterName string) (*rest.Config, error) {
	overrides := clientcmd.ConfigOverrides{}
	// Override the cluster name if provided.
	if clusterName != "" {
		overrides.Context.Cluster = clusterName
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&overrides).ClientConfig()
}

func newKnativeServingAlphaClients(cfg *rest.Config, namespace string) (servingv1alpha1.KnativeServingInterface, error) {
	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return cs.ServingV1alpha1().KnativeServings(namespace), nil
}
