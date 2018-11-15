package api_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/azure-blobstore-resource/api"
	"github.com/pivotal-cf/azure-blobstore-resource/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("In", func() {
	var (
		azureClient *fakes.AzureClient
		in          api.In

		tempDir string
	)

	BeforeEach(func() {
		azureClient = &fakes.AzureClient{}
		in = api.NewIn(azureClient)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("CopyBlobToDestination", func() {
		var (
			snapshot time.Time
		)

		BeforeEach(func() {
			azureClient.GetCall.Returns.BlobReader = ioutil.NopCloser(bytes.NewReader([]byte(`{"key": "value"}`)))
			snapshot = time.Date(2017, time.January, 01, 01, 01, 01, 01, time.UTC)
		})

		It("copies blob from azure blobstore to local destination directory", func() {
			err := in.CopyBlobToDestination(tempDir, "example.json", snapshot)
			Expect(err).NotTo(HaveOccurred())

			Expect(azureClient.GetCall.CallCount).To(Equal(1))
			Expect(azureClient.GetCall.Receives.BlobName).To(Equal("example.json"))
			Expect(azureClient.GetCall.Receives.Snapshot).To(Equal(snapshot))

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "example.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(`{"key": "value"}`))
		})

		Context("when an error occurs", func() {
			Context("when azure client fails to get a blob", func() {
				It("returns an error", func() {
					azureClient.GetCall.Returns.Error = errors.New("failed to get blob")
					err := in.CopyBlobToDestination(tempDir, "example.json", snapshot)
					Expect(err).To(MatchError("failed to get blob"))
				})
			})

			Context("when it fails to create a file into the destination dir", func() {
				It("returns an error", func() {
					err := in.CopyBlobToDestination("/fake/dest/dir", "example.json", snapshot)
					Expect(err).To(MatchError("open /fake/dest/dir/example.json: no such file or directory"))
				})
			})
		})
	})
})
