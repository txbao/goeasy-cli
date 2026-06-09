package generator

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FetchProtoFromURL 下载远程 .proto 到 api/proto/imported/ 并返回相对项目根的路径。
func FetchProtoFromURL(projectDir, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("--from-url is required")
	}
	abs, err := filepath.Abs(projectDir)
	if err != nil {
		return "", err
	}
	module, err := readModulePath(abs)
	if err != nil {
		return "", err
	}
	if !strings.Contains(rawURL, "://") {
		body, err := os.ReadFile(rawURL)
		if err != nil {
			return "", err
		}
		return saveFetchedProto(abs, module, body, filepath.Base(rawURL))
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid --from-url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "file" {
		return "", fmt.Errorf("unsupported URL scheme %q (use http/https/file)", u.Scheme)
	}
	name := filepath.Base(u.Path)
	if name == "" || name == "." || name == "/" {
		name = "remote.proto"
	}

	var body []byte
	switch u.Scheme {
	case "file":
		path := u.Path
		if runtimePath := u.Host; runtimePath != "" && len(path) > 0 && path[0] == '/' {
			path = runtimePath + path
		} else if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		body, err = os.ReadFile(filepath.FromSlash(path))
	default:
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(rawURL)
		if err != nil {
			return "", fmt.Errorf("fetch proto: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("fetch proto: HTTP %s", resp.Status)
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
	}
	if len(body) == 0 {
		return "", fmt.Errorf("empty proto from %s", rawURL)
	}
	return saveFetchedProto(abs, module, body, name)
}

var goPackageRE = regexp.MustCompile(`(?m)^option\s+go_package\s*=\s*".*";`)

func rewriteGoPackage(body []byte, projectModule, protoBase string) []byte {
	protoBase = strings.TrimSuffix(protoBase, ".proto")
	suffix := strings.ReplaceAll(protoBase, "-", "_") + "pb"
	newLine := fmt.Sprintf(`option go_package = "%s/api/proto/gen/imported/%s;%s";`, projectModule, protoBase, suffix)
	s := string(body)
	if goPackageRE.MatchString(s) {
		return []byte(goPackageRE.ReplaceAllString(s, newLine))
	}
	return []byte(strings.TrimSpace(s) + "\n\n" + newLine + "\n")
}

func saveFetchedProto(projectAbs, projectModule string, body []byte, name string) (string, error) {
	if !strings.HasSuffix(strings.ToLower(name), ".proto") {
		name += ".proto"
	}
	destDir := filepath.Join(projectAbs, "api", "proto", "imported")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(destDir, name)
	body = rewriteGoPackage(body, projectModule, name)
	if err := os.WriteFile(dest, body, 0644); err != nil {
		return "", err
	}
	rel, err := filepath.Rel(projectAbs, dest)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	fmt.Printf("  fetched %s\n", rel)
	return rel, nil
}
