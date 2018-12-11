/*

Don't alter this file, it was generated.

*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "code.cloudfoundry.org/cf-operator/pkg/kube/client/clientset/versioned/typed/extendedstatefulset/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeExtendedstatefulsetV1alpha1 struct {
	*testing.Fake
}

func (c *FakeExtendedstatefulsetV1alpha1) ExtendedStatefulSets(namespace string) v1alpha1.ExtendedStatefulSetInterface {
	return &FakeExtendedStatefulSets{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeExtendedstatefulsetV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}