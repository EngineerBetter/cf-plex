package main_test

import (
	"io"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/EngineerBetter/cf-plex/cfcli"
	"github.com/EngineerBetter/cf-plex/clipr"
	"github.com/EngineerBetter/cf-plex/env"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var timeout = "10s"
var addUsageMatcher = "cf-plex add-api \\[-g <group>\\] <apiUrl> \\[<username> <password>\\]"
var listUsageMatcher = "cf-plex list-apis"
var removeUsageMatcher = "cf-plex remove-api \\[-g <group>\\] <apiUrl>"

var _ = Describe("cf-plex", func() {

	var tmpDir string
	var cfUsername string
	var cfPassword string
	var cliPath string
	var envVars []string

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "plex-test")
		Ω(err).ShouldNot(HaveOccurred())

		cfUsername = os.Getenv("CF_USERNAME")
		if cfUsername == "" {
			cfUsername = "testing@engineerbetter.com"
		}

		cfPassword = os.Getenv("CF_PASSWORD")
		Ω(cfPassword).ShouldNot(BeZero(), "CF_PASSWORD env var must be set")

		envVars = env.Set("CF_PLEX_HOME", tmpDir, os.Environ())
		cliPath, err = Build("github.com/EngineerBetter/cf-plex")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		Ω(os.RemoveAll(tmpDir)).Should(Succeed())
	})

	Describe("plugin availability", func() {
		var tmpCfHome string
		var err error

		BeforeEach(func() {
			tmpCfHome, err = ioutil.TempDir("", "plex.cf")
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			Ω(os.RemoveAll(tmpCfHome)).Should(Succeed())
		})

		It("can use existing plugins", func() {
			envVars = env.Set("CF_PLUGIN_HOME", tmpCfHome, os.Environ())
			envVars = env.Set("CF_HOME", tmpCfHome, envVars)

			server := httptest.NewServer(nil)
			defer server.Close()
			clipr.Configure(server.Config, server.URL, "clipr/fixtures/osx/echo", "clipr/fixtures/linux64/echo")

			session, _ := startSession(envVars, "cf", "add-plugin-repo", "test", server.URL)
			Eventually(session).Should(Say("added as 'test'"))
			session, in := startSession(envVars, "cf", "install-plugin", "echo", "-r", "test")
			confirm("(Do you want to install the plugin echo?)", "y", session, in)

			Eventually(session).Should(Say("Plugin EchoDemo v0.1.4 successfully installed"))

			addApi("https://api.run.pivotal.io", cfUsername, cfPassword, envVars, cliPath)
			addApi("https://api.eu-gb.bluemix.net", cfUsername, cfPassword, envVars, cliPath)

			session, _ = startSession(envVars, cliPath, "echo", "foobar")
			expectRunning(session, "cf echo foobar", "https___api.eu-gb.bluemix.net")
			Eventually(session).Should(Say("foobar"))
			expectRunning(session, "cf echo foobar", "https___api.run.pivotal.io")
			Eventually(session).Should(Say("foobar"))
		})
	})

	Describe("adding apis", func() {
		Context("when the username and password are absent", func() {
			It("assumes the user wants interactive login", func() {
				session, in := startSession(envVars, cliPath, "add-api", "https://api.run.pivotal.io")
				confirm("Email>", cfUsername, session, in)
				confirm("Password>", cfPassword, session, in)
				Eventually(session, timeout).Should(Say("Authenticating...\nOK"))
			})
		})

		Context("when the username is absent", func() {
			It("outputs a useful errror message", func() {
				session, _ := startSession(envVars, cliPath, "add-api", "https://api.run.pivotal.io", cfPassword)
				Eventually(session).Should(Say("Usage: " + addUsageMatcher))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the password is absent", func() {
			It("outputs a useful errror message", func() {
				session, _ := startSession(envVars, cliPath, "add-api", "https://api.run.pivotal.io", cfUsername)
				Eventually(session).Should(Say("Usage: " + addUsageMatcher))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the API is absent", func() {
			It("outputs a useful errror message", func() {
				session, _ := startSession(envVars, cliPath, "add-api", cfUsername, cfPassword)
				Eventually(session).Should(Say("Usage: " + addUsageMatcher))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when -g is specified", func() {
			It("requires a group name", func() {
				session, _ := startSession(envVars, cliPath, "add-api", "-g", "https://api.run.pivotal.io", cfUsername, cfPassword)
				Eventually(session).Should(Say("Usage: " + addUsageMatcher))
				Eventually(session).Should(Exit(1))
			})

			It("adds a group", func() {
				session, _ := startSession(envVars, cliPath, "add-api", "-g", "nonprod", "https://api.run.pivotal.io", cfUsername, cfPassword)
				Eventually(session, "5s").Should(Say("Added https://api.run.pivotal.io to group 'nonprod'"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("removing an api", func() {
		Context("when the api is not provided", func() {
			It("outputs a useful error message", func() {
				session, _ := startSession(envVars, cliPath, "remove-api")
				Eventually(session).Should(Say("Usage: " + removeUsageMatcher))
			})
		})

		Context("when -g is specified", func() {
			It("requires a group name", func() {
				session, _ := startSession(envVars, cliPath, "remove-api", "-g", "https://api.run.pivotal.io")
				Eventually(session).Should(Say("Usage: " + removeUsageMatcher))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Describe("running cf commands", func() {
		It("runs commands against multiple Cloud Foundry instances", func() {
			session, _ := startSession(envVars, cliPath, "apps")
			session.Wait("1s")
			Ω(session.Err).Should(Say("No APIs have been set"))

			addApi("https://api.run.pivotal.io", cfUsername, cfPassword, envVars, cliPath)
			addApi("https://api.eu-gb.bluemix.net", cfUsername, cfPassword, envVars, cliPath)

			session, _ = startSession(envVars, cliPath, "list-apis")
			session.Wait("1s")
			Ω(session.Out).Should(Say("https://api.eu-gb.bluemix.net"), "APIs should be alphabetically listed")
			Ω(session.Out).Should(Say("https://api.run.pivotal.io"))
			Ω(string(session.Buffer().Contents())).ShouldNot(ContainSubstring(tmpDir))

			session, in := startSession(envVars, cliPath, "delete-org", "does-not-exist")
			expectRunning(session, "cf delete-org does-not-exist", "https___api.eu-gb.bluemix.net")
			confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
			Eventually(session, timeout).Should(Say("Delete cancelled"))

			expectRunning(session, "cf delete-org does-not-exist", "https___api.run.pivotal.io")
			confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
			Eventually(session, timeout).Should(Say("Delete cancelled"))
			Eventually(session).Should(Exit(0))

			removeApi("https://api.run.pivotal.io", envVars, cliPath)
			removeApi("https://api.eu-gb.bluemix.net", envVars, cliPath)

			session, _ = startSession(envVars, cliPath, "apps")
			Eventually(session.Err).Should(Say("No APIs have been set"))
		})

		It("fails when subprocesses fail", func() {
			session, _ := startSession(envVars, cliPath, "rubbish")
			Eventually(session).Should(Exit(1))
		})

		It("does not run a command after it has failed against one API", func() {
			addApi("https://api.run.pivotal.io", cfUsername, cfPassword, envVars, cliPath)
			addApi("https://api.eu-gb.bluemix.net", cfUsername, cfPassword, envVars, cliPath)

			session, _ := startSession(envVars, cliPath, "target", "-s", "does-not-exist")
			session.Wait()
			output := string(session.Buffer().Contents())
			Ω(strings.Count(output, "FAILED")).ShouldNot(BeNumerically(">", 1))
		})

		Context("when --force is supplied", func() {
			It("runs against all targets even if the first command fails", func() {
				addApi("https://api.run.pivotal.io", cfUsername, cfPassword, envVars, cliPath)
				addApi("https://api.eu-gb.bluemix.net", cfUsername, cfPassword, envVars, cliPath)
				// BlueMix has org named after user. Neither public CF allows us to create orgs
				session, _ := startSession(envVars, cliPath, "target", "-o", "testing@engineerbetter.com", "--force")
				session.Wait(timeout)
				output := string(session.Buffer().Contents())
				Ω(strings.Count(output, "FAILED")).Should(BeNumerically("==", 1))
			})
		})
	})

	Describe("Specifying APIs via CF_PLEX_APIS", func() {
		var cfEnvs string

		Context("when using default separators", func() {
			BeforeEach(func() {
				cfEnvs = cfUsername + "^" + cfPassword + ">https://api.run.pivotal.io;" + cfUsername + "^" + cfPassword + ">https://api.eu-gb.bluemix.net"
				envVars = append(envVars, "CF_PLEX_APIS="+cfEnvs)
			})

			It("Runs commands against APIs in ENV, logging in only once", func() {
				session, in := startSession(envVars, cliPath, "delete-org", "does-not-exist")
				Eventually(session, timeout).Should(Say("Setting api endpoint to https://api.run.pivotal.io...\nOK"))
				Eventually(session, timeout).Should(Say("Authenticating...\nOK"))
				Eventually(session, timeout).Should(Say("Setting api endpoint to https://api.eu-gb.bluemix.net"))
				Eventually(session, timeout).Should(Say("Authenticating...\nOK"))
				expectRunning(session, "cf delete-org does-not-exist", "https___api.run.pivotal.io")
				confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
				Eventually(session, timeout).Should(Say("Delete cancelled"))

				expectRunning(session, "cf delete-org does-not-exist", "https___api.eu-gb.bluemix.net")
				confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
				Eventually(session, timeout).Should(Say("Delete cancelled"))
				Eventually(session).Should(Exit(0))

				session, _ = startSession(envVars, cliPath, "apps")
				Ω(session.Wait(timeout).Out.Contents()).ShouldNot(ContainSubstring("Not logged in"))
				Ω(session.Wait(timeout).Out.Contents()).ShouldNot(ContainSubstring("Authenticating..."))
			})

			It("Disallows add-api", func() {
				session, _ := startSession(envVars, cliPath, "add-api", "https://api.run.pivotal.io")
				Eventually(session).Should(Say("Managing APIs is not allowed when CF_PLEX_APIS is set"))
				Eventually(session).Should(Exit(1))
			})

			It("Disallows list-apis", func() {
				session, _ := startSession(envVars, cliPath, "list-apis")
				Eventually(session).Should(Say("Managing APIs is not allowed when CF_PLEX_APIS is set"))
				Eventually(session).Should(Exit(1))
			})

			It("Disallows remove-api", func() {
				session, _ := startSession(envVars, cliPath, "remove-api", "https://api.run.pivotal.io")
				Eventually(session).Should(Say("Managing APIs is not allowed when CF_PLEX_APIS is set"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when custom separators are defined", func() {
			BeforeEach(func() {
				cfEnvs = cfUsername + "-foo-" + cfPassword + "_https://api.run.pivotal.io|" + cfUsername + "-foo-" + cfPassword + "_https://api.eu-gb.bluemix.net"
				envVars = append(envVars, "CF_PLEX_SEP_TRIPLE=|")
				envVars = append(envVars, "CF_PLEX_SEP_CREDS_API=_")
				envVars = append(envVars, "CF_PLEX_SEP_USER_PASS=-foo-")
				envVars = append(envVars, "CF_PLEX_APIS="+cfEnvs)
			})

			It("still works", func() {
				session, in := startSession(envVars, cliPath, "delete-org", "does-not-exist")
				Eventually(session, timeout).Should(Say("Setting api endpoint to https://api.run.pivotal.io...\nOK"))
				Eventually(session, timeout).Should(Say("Authenticating...\nOK"))
				Eventually(session, timeout).Should(Say("Setting api endpoint to https://api.eu-gb.bluemix.net"))
				Eventually(session, timeout).Should(Say("Authenticating...\nOK"))
				expectRunning(session, "cf delete-org does-not-exist", "https___api.run.pivotal.io")
				confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
				Eventually(session, timeout).Should(Say("Delete cancelled"))

				expectRunning(session, "cf delete-org does-not-exist", "https___api.eu-gb.bluemix.net")
				confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
				Eventually(session, timeout).Should(Say("Delete cancelled"))
				Eventually(session).Should(Exit(0))
				Ω(string(session.Buffer().Contents())).ShouldNot(ContainSubstring(cfPassword))
			})
		})
	})

	Describe("group management", func() {
		It("errs when a unrecognised group is referenced", func() {
			session, _ := startSession(envVars, cliPath, "-g", "nonprod", "delete-org", "does-not-exist")
			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("Group 'nonprod' not recognised"))
		})

		It("allows groups to be added, listed, run against, and removed", func() {
			session, _ := startSession(envVars, cliPath, "add-api", "-g", "nonprod", "https://api.run.pivotal.io", cfUsername, cfPassword)
			Eventually(session, "5s").Should(Exit(0))
			session, _ = startSession(envVars, cliPath, "list-apis")
			Eventually(session).Should(Say("nonprod"))
			Eventually(session).Should(Say("\thttps://api.run.pivotal.io"))
			Eventually(session).Should(Exit(0))

			session, _ = startSession(envVars, cliPath, "delete-org", "my-org")
			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("-g <group> is mandatory whenever groups have been added. Use '-g default' to target APIs without an explicit group."))

			session, in := startSession(envVars, cliPath, "-g", "nonprod", "delete-org", "does-not-exist")
			expectRunning(session, "cf delete-org does-not-exist", "https___api.run.pivotal.io")
			confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
			Eventually(session, timeout).Should(Say("Delete cancelled"))
			Eventually(session).Should(Exit(0))

			session, _ = startSession(envVars, cliPath, "remove-api", "-g", "nonprod", "https://api.run.pivotal.io")
			Eventually(session).Should(Say("Removed https://api.run.pivotal.io from 'nonprod'"))
			Eventually(session).Should(Exit(0))

			session, _ = startSession(envVars, cliPath, "list-apis")
			Eventually(session).Should(Exit(0))
			Eventually(session).ShouldNot(Say("nonprod"))
			Eventually(session).ShouldNot(Say("\thttps://api.run.pivotal.io"))
		})

		It("does not run commands against APIs not in the group", func() {
			addApi("https://api.eu-gb.bluemix.net", cfUsername, cfPassword, envVars, cliPath)
			session, _ := startSession(envVars, cliPath, "add-api", "-g", "nonprod", "https://api.run.pivotal.io", cfUsername, cfPassword)
			Eventually(session, timeout).Should(Exit(0))
			session, in := startSession(envVars, cliPath, "-g", "nonprod", "delete-org", "does-not-exist")
			expectRunning(session, "cf delete-org does-not-exist", "https___api.run.pivotal.io")
			confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
			Eventually(session, timeout).Should(Exit(0))
			Ω(string(session.Buffer().Contents())).ShouldNot(ContainSubstring("api.eu-gb.bluemix.net"))
		})

		It("does not treat batch APIs as a group", func() {
			cfEnvs := cfUsername + "^" + cfPassword + ">https://api.eu-gb.bluemix.net"
			cfPlexApisEnvVars := append(envVars, "CF_PLEX_APIS="+cfEnvs)
			session, in := startSession(cfPlexApisEnvVars, cliPath, "delete-org", "does-not-exist")
			Eventually(session, timeout).Should(Say("Authenticating...\nOK"))
			expectRunning(session, "cf delete-org does-not-exist", "https___api.eu-gb.bluemix.net")
			confirm("Really delete the org does-not-exist and everything associated with it?", "n", session, in)
			Eventually(session).Should(Exit(0))

			session, _ = startSession(envVars, cliPath, "add-api", "-g", "nonprod", "https://api.run.pivotal.io", cfUsername, cfPassword)
			Eventually(session, timeout).Should(Exit(0))

			session, _ = startSession(envVars, cliPath, "list-apis")
			Eventually(session).Should(Exit(0))
			Eventually(session).ShouldNot(Say("batch"))
			Eventually(session).ShouldNot(Say("\thttps://api.eu-gb.bluemix.net"))
		})
	})

	Describe("asking for help", func() {
		It("outputs usage of all commands", func() {
			session, _ := startSession(envVars, cliPath)
			expectUsage(session)
		})
	})

	Context("when no sub-command is specified", func() {
		It("outputs usage of all commands", func() {
			session, _ := startSession(envVars, cliPath)
			expectUsage(session)
		})
	})
})

func startSession(envVars []string, args ...string) (*Session, io.Writer) {
	cmd := cfcli.CommandWithEnv(envVars, args...)
	in, err := cmd.StdinPipe()
	Ω(err).ShouldNot(HaveOccurred())
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	return session, in
}

func addApi(api, cfUsername, cfPassword string, envVars []string, cliPath string) {
	session, _ := startSession(envVars, cliPath, "add-api", api, cfUsername, cfPassword)
	session.Wait(timeout)
	Ω(session.Out).Should(Say("Setting api endpoint to " + api + "...\nOK"))
	Ω(session.Out).Should(Say("Authenticating...\nOK"))
	Ω(string(session.Buffer().Contents())).ShouldNot(ContainSubstring(cfPassword))
}

func removeApi(api string, envVars []string, cliPath string) {
	session, _ := startSession(envVars, cliPath, "remove-api", api)
	session.Wait("1s")
	Ω(session.Out).Should(Say("Removed " + api))
}

func confirm(expectedPrompt, input string, session *Session, in io.Writer) {
	Eventually(session, timeout).Should(Say(expectedPrompt))
	in.Write([]byte(input + "\n"))
}

func expectUsage(session *Session) {
	Eventually(session).Should(Say("Usage:"))
	Eventually(session).Should(Say("cf-plex \\[-g <group>\\] <cf cli command> \\[--force\\]"))
	Eventually(session).Should(Say(addUsageMatcher))
	Eventually(session).Should(Say(listUsageMatcher))
	Eventually(session).Should(Say(removeUsageMatcher))
}

func expectRunning(session *Session, cmd, api string) {
	Eventually(session).Should(Say("\n\nRunning '" + cmd + "' on " + api))
}
