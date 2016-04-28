package main_test

import (
	. "github.com/EngineerBetter/cf-plex"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cf-plex", func() {
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
})
