package clusterurlmonitor_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/route-monitor-operator/controllers/clusterurlmonitor"
)

func TestDomainExtraction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterUrlMonitor Domain Extraction Suite")
}

var _ = Describe("Domain Extraction", func() {
	Describe("extractDomain", func() {
		var reconciler *clusterurlmonitor.ClusterUrlMonitorReconciler

		BeforeEach(func() {
			reconciler = &clusterurlmonitor.ClusterUrlMonitorReconciler{}
		})

		Context("when using regex patterns", func() {
			It("should extract domain from a rosa hypershift cluster using the default pattern", func() {
				pattern := "^[^.]+\\.(.+)$"
				result, err := reconciler.ExtractDomainForTesting("rosa.example.com", pattern)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("example.com"))
			})

			It("should extract domain from a normal cluster using the default pattern", func() {
				pattern := "^[^.]+\\.(.+)$"
				result, err := reconciler.ExtractDomainForTesting("api.cluster.example.com", pattern)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal("cluster.example.com"))
			})

			It("should return error for invalid pattern", func() {
				pattern := "^[invalid"
				_, err := reconciler.ExtractDomainForTesting("rosa.example.com", pattern)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid domain extract pattern"))
			})

			It("should return error when pattern doesn't match", func() {
				pattern := "^notfound\\.(.+)$"
				_, err := reconciler.ExtractDomainForTesting("rosa.example.com", pattern)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("domain extract pattern did not match"))
			})
		})
	})
})
