package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type dockerPlatform struct {
	OS           string
	Architecture string
	Variant      string
}

func dockerPlatformForRuntime() dockerPlatform {
	p := dockerPlatform{OS: runtime.GOOS}
	switch runtime.GOARCH {
	case "amd64":
		p.Architecture = "amd64"
	case "arm64":
		p.Architecture = "arm64"
	case "arm":
		p.Architecture = "arm"
		// Best-effort: runtime doesn't expose GOARM directly; keep empty to match any.
	default:
		p.Architecture = runtime.GOARCH
	}
	return p
}

func dockerInstalledConfigDigest(ref string) (string, error) {
	out, err := runDocker("image", "inspect", "--format", "{{.Id}}", ref)
	if err != nil {
		return "", err
	}
	return extractDigest(strings.TrimSpace(out)), nil
}

func dockerRemoteConfigDigestForCurrentPlatform(ref string) (string, error) {
	out, err := runDocker("manifest", "inspect", "--verbose", ref)
	if err != nil {
		return "", err
	}
	p := dockerPlatformForRuntime()
	return configDigestFromManifestInspectVerbose([]byte(out), p)
}

func configDigestFromManifestInspectVerbose(raw []byte, p dockerPlatform) (string, error) {
	// docker manifest inspect --verbose returns an array of entries.
	// Each entry includes:
	// - Descriptor.platform.{os,architecture,variant}
	// - OCIManifest.config.digest  (this is the image ID/config digest)
	var entries []struct {
		Descriptor struct {
			Platform struct {
				Architecture string `json:"architecture"`
				OS           string `json:"os"`
				Variant      string `json:"variant"`
			} `json:"platform"`
		} `json:"Descriptor"`
		OCIManifest struct {
			Config struct {
				Digest string `json:"digest"`
			} `json:"config"`
		} `json:"OCIManifest"`
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&entries); err != nil {
		// fall back to permissive decode
		if err2 := json.Unmarshal(raw, &entries); err2 != nil {
			return "", fmt.Errorf("parse docker manifest inspect --verbose: %w", err)
		}
	}

	for _, e := range entries {
		pl := e.Descriptor.Platform
		if pl.OS != "" && p.OS != "" && pl.OS != p.OS {
			continue
		}
		if pl.Architecture != "" && p.Architecture != "" && pl.Architecture != p.Architecture {
			continue
		}
		if p.Variant != "" && pl.Variant != "" && pl.Variant != p.Variant {
			continue
		}
		return extractDigest(e.OCIManifest.Config.Digest), nil
	}
	return "", nil
}

func extractDigest(s string) string {
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, '@'); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimSpace(s)
}

func runDocker(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("docker %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("docker %s failed: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
