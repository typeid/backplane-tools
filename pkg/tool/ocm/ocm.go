package ocm

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	gogithub "github.com/google/go-github/github"
	"github.com/tnierman/backplane-tools/pkg/source/github"
)

// Tool implements the interface to manage the 'ocm-cli' binary
type Tool struct {
	source *github.GithubTool
}

func NewTool() *Tool {
	t := &Tool{
		source: github.NewGithubTool("openshift-online", "ocm-cli"),
	}
	return t
}

func (t *Tool) Name() string {
	return "ocm"
}

func (t *Tool) Install(rootDir string) error {
	// Pull latest release from GH
	release, err := t.source.FetchLatestRelease()
	if err != nil {
		return err
	}

	// Determine which assets to download
	var checksumAsset gogithub.ReleaseAsset
	var ocmBinaryAsset gogithub.ReleaseAsset

assetLoop:
	for _, asset := range release.Assets {
		// Exclude assets that do not match system OS
		if !strings.Contains(asset.GetName(), runtime.GOOS) {
			continue assetLoop
		}
		// Exclude assets that do not match system architecture
		if !strings.Contains(asset.GetName(), runtime.GOARCH) {
			continue assetLoop
		}

		if strings.Contains(asset.GetName(), "sha256") {
			if checksumAsset.GetName() != "" {
				return fmt.Errorf("detected duplicate ocm-cli checksum assets")
			}
			checksumAsset = asset
		} else {
			if ocmBinaryAsset.GetName() != "" {
				return fmt.Errorf("detected duplicate ocm-cli binary assets")
			}
			ocmBinaryAsset = asset
		}
	}

	if checksumAsset.GetName() == "" || ocmBinaryAsset.GetName() == "" {
		return fmt.Errorf("failed to find ocm-cli or it's checksum")
	}

	// Download the arch- & os-specific assets
	versionedDir := filepath.Join(rootDir, "ocm", release.GetTagName())
	err = os.MkdirAll(versionedDir, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("failed to create version-specific directory '%s': %w", versionedDir, err)
	}

	//err = t.source.DownloadReleaseAssets([]gogithub.ReleaseAsset{checksumAsset, ocmBinaryAsset}, versionedDir)
	//if err != nil {
	//	return fmt.Errorf("failed to download one or more assets: %w", err)
	//}

	// Verify checksum of downloaded assets
	ocmBinaryFilepath := filepath.Join(versionedDir, ocmBinaryAsset.GetName())
	fileBytes, err := ioutil.ReadFile(ocmBinaryFilepath)
	if err != nil {
		return fmt.Errorf("failed to read ocm-cli binary file '%s' while generating sha256sum: %w", ocmBinaryFilepath, err)
	}
	sumBytes := sha256.Sum256(fileBytes)
	binarySum := fmt.Sprintf("%x", sumBytes[:])
	fmt.Println("sum: ", binarySum)

	checksumFilePath := filepath.Join(versionedDir, checksumAsset.GetName())
	checksumBytes, err := ioutil.ReadFile(checksumFilePath)
	if err != nil {
		return fmt.Errorf("failed to read ocm-cli checksum file '%s': %w", checksumFilePath, err)
	}
	checksum := strings.Split(string(checksumBytes), " ")[0]
	fmt.Println("checksum: ", checksum)
	if strings.TrimSpace(binarySum) != strings.TrimSpace(checksum) {
		fmt.Printf("WARNING: Checksum for ocm-cli does not match the calculated value. Please retry installation. If issue persists, this tool can be downloaded manually at %s\n", ocmBinaryAsset.GetBrowserDownloadURL())
	}
	return nil
}

func (t *Tool) Configure() error {
	return nil
}

func (t *Tool) Remove() error {
	return nil
}