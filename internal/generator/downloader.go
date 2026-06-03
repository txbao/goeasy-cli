package generator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadTemplate(templateName, version, targetDir string) error {
	base := os.Getenv("GoEasy_TEMPLATE_URL")
	if base == "" {
		base = "https://yourdomain.com/templates"
	}
	base = strings.TrimRight(base, "/")
	url := fmt.Sprintf("%s/%s-%s.zip", base, templateName, version)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	tmpZip := filepath.Join(os.TempDir(), "goesy-"+templateName+"-"+version+".zip")
	f, err := os.Create(tmpZip)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	f.Close()

	return Unzip(tmpZip, targetDir)
}
