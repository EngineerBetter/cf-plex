package main_test

import (
	. "github.com/EngineerBetter/cf-plex"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
)

var _ = Describe("cf-plex", func() {
	var tmpDir string

	BeforeSuite(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "plex")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		Ω(os.RemoveAll(tmpDir)).Should(Succeed())
	})

	Describe("SetEnv", func() {
		Context("when the env var is already set", func() {
			It("replaces the value", func() {
				env := []string{"KEY=value", "CF_HOME=foo"}
				env = SetEnv("CF_HOME", "bar", env)
				Ω(env).Should(ContainElement("CF_HOME=bar"))
				Ω(env).Should(ContainElement("KEY=value"))
				Ω(env).ShouldNot(ContainElement("CF_HOME=foo"))
			})
		})

		Context("when the env var is not set already", func() {
			It("adds the value", func() {
				env := []string{"KEY=value"}
				env = SetEnv("CF_HOME", "bar", env)
				Ω(env).Should(ContainElement("CF_HOME=bar"))
				Ω(env).Should(ContainElement("KEY=value"))
			})
		})
	})

	It("calls external things", func() {
		env := os.Environ()
		env = SetEnv("CF_HOME", tmpDir, env)
		cmd := CommandWithEnv(env, "cf", "api")

		Ω(Output(cmd)).Should(Equal("No api endpoint set. Use 'cf api' to set an endpoint\n"))

		cmd = CommandWithEnv(env, "cf", "api", "https://api.bosh-lite.com", "--skip-ssl-validation")
		err := cmd.Run()
		Ω(err).ShouldNot(HaveOccurred())
	})
})
