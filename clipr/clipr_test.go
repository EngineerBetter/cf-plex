package clipr_test

import (
	. "github.com/EngineerBetter/cf-plex/clipr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bitly/go-simplejson"

	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("CLIPR", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewServer(nil)
		Configure(server.Config, server.URL, "fixtures/osx/echo")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("/list", func() {
		It("returns a listing containing the echo plugin", func() {
			resp, err := http.Get(server.URL + "/list")
			Ω(err).ShouldNot(HaveOccurred())

			json, err := simplejson.NewFromReader(resp.Body)
			Ω(err).ShouldNot(HaveOccurred())

			echoNode := json.Get("plugins").GetIndex(0)
			Ω(echoNode.Get("name").MustString()).Should(Equal("echo"))
			bins, err := echoNode.Get("binaries").Array()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bins).Should(ContainElement(SatisfyAll(
				HaveKeyWithValue("platform", "osx"),
				HaveKeyWithValue("url", server.URL+"/bin/osx/echo"),
			)))
			Ω(bins).Should(ContainElement(SatisfyAll(
				HaveKeyWithValue("platform", "win64"),
				HaveKeyWithValue("url", server.URL+"/bin/windows64/echo.exe"),
			)))
		})
	})

	Describe("serving binaries", func() {
		It("serves the plugin binaries as advertised", func() {
			resp, err := http.Get(server.URL + "/list")
			Ω(err).ShouldNot(HaveOccurred())

			json, err := simplejson.NewFromReader(resp.Body)
			Ω(err).ShouldNot(HaveOccurred())

			firstUrl := json.Get("plugins").GetIndex(0).Get("binaries").GetIndex(0).Get("url").MustString()
			resp, err = http.Get(firstUrl)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(resp.StatusCode).Should(Equal(200))
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			Ω(err).ShouldNot(HaveOccurred())
			fileBytes, err := ioutil.ReadFile("fixtures/osx/echo")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(bodyBytes).Should(Equal(fileBytes))
		})
	})
})
