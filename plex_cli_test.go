package main_test

import (
	. "github.com/EngineerBetter/cf-plex"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var _ = Describe("cf-plex", func() {
	var tmpDir string
	var cfUsername string
	var cfPassword string

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "plex")
		Ω(err).ShouldNot(HaveOccurred())

		cfUsername = os.Getenv("CF_USERNAME")
		if cfUsername == "" {
			cfUsername = "testing@engineerbetter.com"
		}

		cfPassword = os.Getenv("CF_PASSWORD")
		Ω(cfPassword).ShouldNot(BeZero(), "CF_PASSWORD env var must be set")
	})

	AfterEach(func() {
		Ω(os.RemoveAll(tmpDir)).Should(Succeed())
	})

	It("runs commands against multiple Cloud Foundry instances", func() {
		env := os.Environ()
		env = SetEnv("CF_PLEX_HOME", tmpDir, env)
		cliPath, err := Build("github.com/EngineerBetter/cf-plex")
		Ω(err).ShouldNot(HaveOccurred())

		session, _ := startSession(env, cliPath, "apps")
		session.Wait("1s")
		Ω(session.Err).Should(Say("No APIs have been set"))

		addApi("https://api.run.pivotal.io", cfUsername, cfPassword, env, cliPath)
		addApi("https://api.eu-gb.bluemix.net", cfUsername, cfPassword, env, cliPath)

		session, _ = startSession(env, cliPath, "list-apis")
		session.Wait("1s")
		Ω(session.Out).Should(Say("https___api.eu-gb.bluemix.net"), "APIs should be alphabetically listed")
		Ω(session.Out).Should(Say("https___api.run.pivotal.io"))
		Ω(string(session.Buffer().Contents())).ShouldNot(ContainSubstring(tmpDir))

		session, in := startSession(env, cliPath, "delete-org", "does-not-exist")
		confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
		Eventually(session, "5s").Should(Say("Delete cancelled"))

		confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
		Eventually(session, "5s").Should(Say("Delete cancelled"))
		Eventually(session).Should(Exit(0))

		removeApi("https://api.run.pivotal.io", env, cliPath)
		removeApi("https://api.eu-gb.bluemix.net", env, cliPath)

		session, _ = startSession(env, cliPath, "apps")
		Eventually(session.Err).Should(Say("No APIs have been set"))
	})

	It("fails when subprocesses fail", func() {
		env := os.Environ()
		env = SetEnv("CF_PLEX_HOME", tmpDir, env)
		cliPath, err := Build("github.com/EngineerBetter/cf-plex")
		Ω(err).ShouldNot(HaveOccurred())
		session, _ := Start(CommandWithEnv(env, cliPath, "rubbish"), GinkgoWriter, GinkgoWriter)
		Eventually(session).Should(Exit(1))
	})

	It("does not run a command after it has failed against one API", func() {
		env := os.Environ()
		env = SetEnv("CF_PLEX_HOME", tmpDir, env)
		cliPath, err := Build("github.com/EngineerBetter/cf-plex")
		Ω(err).ShouldNot(HaveOccurred())

		addApi("https://api.run.pivotal.io", cfUsername, cfPassword, env, cliPath)
		addApi("https://api.eu-gb.bluemix.net", cfUsername, cfPassword, env, cliPath)

		session, _ := startSession(env, cliPath, "target", "-s", "does-not-exist")
		session.Wait()
		output := string(session.Buffer().Contents())
		Ω(strings.Count(output, "FAILED")).ShouldNot(BeNumerically(">", 1))
	})
})

func startSession(env []string, args ...string) (*Session, io.Writer) {
	cmd := CommandWithEnv(env, args...)
	in, err := cmd.StdinPipe()
	Ω(err).ShouldNot(HaveOccurred())
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	return session, in
}

func addApi(api, cfUsername, cfPassword string, env []string, cliPath string) {
	session, _ := startSession(env, cliPath, "add-api", api, cfUsername, cfPassword)
	session.Wait("5s")
	Ω(session.Out).Should(Say("Setting api endpoint to " + api + "...\nOK"))
	Ω(session.Out).Should(Say("Authenticating...\nOK"))
}

func removeApi(api string, env []string, cliPath string) {
	session, _ := startSession(env, cliPath, "remove-api", api)
	session.Wait("1s")
	Ω(session.Out).Should(Say("Removed " + api))
}

func confirm(expectedPrompt, input string, session *Session, in io.Writer) {
	time.Sleep(1 * time.Second)
	Ω(session).Should(Say(expectedPrompt))
	in.Write([]byte(input + "\n"))
}
