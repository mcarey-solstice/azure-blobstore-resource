package acceptance_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("In", func() {
	var (
		container string
		tempDir   string
	)

	BeforeEach(func() {
		rand.Seed(time.Now().UTC().UnixNano())
		container = fmt.Sprintf("azureblobstore%d", rand.Int())
		createContainer(container)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		deleteContainer(container)
	})

	Context("when given a specific snapshot version and destination directory", func() {
		var (
			snapshotTimestamp *time.Time
		)

		BeforeEach(func() {
			snapshotTimestamp = createBlobWithSnapshot(container, "example.json")
		})

		It("downloads the specific blob version and copies it to destination directory", func() {
			in := exec.Command(pathToIn, tempDir)
			in.Stderr = os.Stderr

			stdin, err := in.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"versioned_file": "example.json"
					},
					"version": { "snapshot": %q }
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
				snapshotTimestamp.Format(time.RFC3339Nano),
			))
			Expect(err).NotTo(HaveOccurred())

			outputJSON, err := in.Output()
			Expect(err).NotTo(HaveOccurred())

			var output struct {
				Version struct {
					Snapshot time.Time `json:"snapshot"`
				} `json:"version"`
				Metadata []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"metadata"`
			}
			err = json.Unmarshal(outputJSON, &output)
			Expect(err).NotTo(HaveOccurred())

			Expect(output.Version.Snapshot).To(Equal(*snapshotTimestamp))
			Expect(output.Metadata[0].Name).To(Equal("filename"))
			Expect(output.Metadata[0].Value).To(Equal("example.json"))
			Expect(output.Metadata[1].Name).To(Equal("url"))
			url, err := url.Parse(output.Metadata[1].Value)
			Expect(err).NotTo(HaveOccurred())
			Expect(url.Hostname()).To(Equal(fmt.Sprintf("%s.blob.core.windows.net", config.StorageAccountName)))
			Expect(url.EscapedPath()).To(Equal(fmt.Sprintf("/%s/example.json", container)))
			Expect(len(url.Query()["snapshot"][0])).To(Equal(28)) // azure is sensetive to trailing zero's
			_, err = os.Stat(filepath.Join(tempDir, "example.json"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when given a very large blob", func() {
		var (
			snapshotTimestamp *time.Time
			fileSize          int64 = 20 * 1000 * 1000
		)

		BeforeEach(func() {
			err := os.Mkdir(filepath.Join(tempDir, "some-resource"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = makeLargeFile(filepath.Join(tempDir, "some-resource", "big_file"), fileSize)
			Expect(err).NotTo(HaveOccurred())

			snapshotTimestamp = uploadBlobWithSnapshot(container, "big_file_on_azure", filepath.Join(tempDir, "some-resource", "big_file"))
		})

		It("downloads the specific blob version and copies it to destination directory", func() {
			in := exec.Command(pathToIn, tempDir)
			in.Stderr = os.Stderr

			stdin, err := in.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"versioned_file": "big_file_on_azure"
					},
					"version": { "snapshot": %q }
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
				snapshotTimestamp.Format(time.RFC3339Nano),
			))
			Expect(err).NotTo(HaveOccurred())

			outputJSON, err := in.Output()
			Expect(err).NotTo(HaveOccurred())

			var output struct {
				Version struct {
					Snapshot time.Time `json:"snapshot"`
				} `json:"version"`
				Metadata []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"metadata"`
			}
			err = json.Unmarshal(outputJSON, &output)
			Expect(err).NotTo(HaveOccurred())

			fileInfo, err := os.Stat(filepath.Join(tempDir, "big_file_on_azure"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Size()).To(Equal(fileSize))
		})
	})
})
