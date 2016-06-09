package target_test

import (
	. "github.com/EngineerBetter/cf-plex/target"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("target", func() {
	var tmpDir string

	BeforeEach(func() {
		tmpDir, err := ioutil.TempDir("", "plex-target")
		defer os.RemoveAll(tmpDir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Describe("Group management", func() {
		It("allows groups to be added, listed, and deleted", func() {
			AddToGroup(tmpDir, "prod", "https://api.example.com")
			Ω(exists(filepath.Join(tmpDir, "groups", "prod"))).Should(BeTrue(), "group dir should exist after creation")

			RemoveFromGroup(tmpDir, "prod", "https://api.example.com")
			Ω(exists(filepath.Join(tmpDir, "groups", "prod"))).Should(BeFalse(), "group dir should not exist after removal")
		})
	})

	Describe("reserved group names", func() {
		It("forbids groups called default from being added", func() {
			_, err := AddToGroup(tmpDir, "default", "https://api.example.com")
			Ω(err).Should(MatchError("group name default is reserved"))
		})
	})
})

func exists(dir string) bool {
	_, err := os.Stat(dir)

	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
