package api

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

type In struct {
	azureClient azureClient
}

func NewIn(azureClient azureClient) In {
	return In{
		azureClient: azureClient,
	}
}

func (i In) CopyBlobToDestination(destinationDir, blobName string, snapshot time.Time) error {
	blobReader, err := i.azureClient.Get(blobName, snapshot)
	if err != nil {
		return err
	}
	defer blobReader.Close()

	file, err := os.Create(filepath.Join(destinationDir, blobName))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, blobReader)
	if err != nil {
		return err
	}

	return nil
}
