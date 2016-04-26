package main_test

import (
	. "github.com/EngineerBetter/cf-plex"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"io/ioutil"
	"os"
	"time"
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
		cliPath, err := Build("github.com/EngineerBetter/cf-plex")
		Ω(err).ShouldNot(HaveOccurred())
		session, err := Start(CommandWithEnv(env, cliPath, "api"), GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		session.Wait()
		Ω(session.Out).Should(Say("No api endpoint set. Use 'cf api' to set an endpoint\n"))

		session, err = Start(CommandWithEnv(env, cliPath, "api", "https://api.bosh-lite.com", "--skip-ssl-validation"), GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		session.Wait()
		Ω(session.Out).Should(Say("Setting api endpoint to https://api.bosh-lite.com..."))
		Ω(session.Out).Should(Say("OK"))

		cmd := CommandWithEnv(env, cliPath, "login")
		in, _ := cmd.StdinPipe()
		Ω(err).ShouldNot(HaveOccurred())
		session, err = Start(cmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		time.Sleep(1 * time.Second)
		in.Write([]byte("admin\n"))
		time.Sleep(1 * time.Second)
		in.Write([]byte("admin\n"))
		time.Sleep(1 * time.Second)
		in.Write([]byte("1\n"))
		time.Sleep(1 * time.Second)
		in.Write([]byte("1\n"))
		Eventually(session, "5s").Should(Exit(0))
	})

	It("fails when subprocesses fail", func() {
		env := os.Environ()
		env = SetEnv("CF_HOME", tmpDir, env)
		cliPath, err := Build("github.com/EngineerBetter/cf-plex")
		Ω(err).ShouldNot(HaveOccurred())
		session, err := Start(CommandWithEnv(env, cliPath, "rubbish"), GinkgoWriter, GinkgoWriter)
		Eventually(session).Should(Exit(1))
	})
})
