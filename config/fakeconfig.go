package config

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/bpineau/katafygio/pkg/log"
)

var (
	// FakeResyncInterval is the interval between resyncs during unit tests
	FakeResyncInterval = time.Duration(time.Second)

	// Labels used to filter objets during unit tests runs
	Labels = map[string]string{"foo": "bar", "spam": "egg"}
)

// FakeConfig returns a configuration struct using a fake clientset, for unit tests
func FakeConfig(objects ...runtime.Object) *KdnConfig {
	c := &KdnConfig{
		DryRun:     true,
		Logger:     log.New("", "", "test"),
		ClientSet:  fake.NewSimpleClientset(objects...),
		LocalDir:   "/tmp/katafygio",
		Filter:     "foo=bar,spam=egg",
		ResyncIntv: FakeResyncInterval,
	}

	return c
}

// FakeClientSet provides a fake.NewSimpleClientset, useful for testing without a real cluster
func FakeClientSet() *fake.Clientset {
	return fake.NewSimpleClientset()
}
