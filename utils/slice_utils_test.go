package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checking books out of the library", func() {
	Context("and the book is available", func() {
		It("lends it to the reader", func() {
			Expect(true).To(Equal(true))
		})
	})
})

//func TestFind(t *testing.T) {
//	b := []string{"a", "b", "c"}
//
//	index, found := Find(b, "b")
//
//	if index != 1 || !found {
//		t.Errorf("Cannot find")
//	}
//}
