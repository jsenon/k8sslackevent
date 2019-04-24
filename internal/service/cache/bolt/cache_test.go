package bolt_test

import (
	"testing"

	"github.com/jsenon/k8sslackevent/internal/service/cache/bolt"
	. "github.com/onsi/gomega"
)

func TestNominalCaseNotSended(t *testing.T) {
	g := NewGomegaWithT(t)

	caching := bolt.NewCache()
	err := caching.Init()
	g.Expect(err).To(BeNil())

	msg := "this is the nominal case test"
	g.Expect(caching.CheckIfSended(msg)).To(BeFalse())

	err = caching.SaveMsg(msg)
	g.Expect(err).To(BeNil())

	g.Expect(caching.CheckIfSended(msg)).To(BeTrue())
}

