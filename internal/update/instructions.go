package update

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v3/internal/system"
)

const WindowsDistributionHoldMessage = "Windows binary distribution and Scoop are temporarily unavailable until publicly trusted Authenticode signing is enforced."

// GentleAIModulePath returns the Go module path for gentle-ai. If versionOrMajor
// is provided, it derives the module suffix for the corresponding major version
// (unsuffixed for v0/v1, /vN for N >= 2), accepts an explicit module or import
// path, or falls back to the default v3 module path.
func GentleAIModulePath(tool ToolInfo, versionOrMajor ...string) string {
	repo := strings.ToLower(fmt.Sprintf("github.com/%s/%s", strings.TrimSpace(tool.Owner), strings.TrimSpace(tool.Repo)))
	if repo == "github.com//" {
		repo = "github.com/gentleman-programming/gentle-ai"
	}
	if len(versionOrMajor) > 0 && strings.TrimSpace(versionOrMajor[0]) != "" {
		target := strings.TrimSpace(versionOrMajor[0])
		if i := strings.Index(target, "@"); i >= 0 {
			target = target[:i]
		}
		if strings.Contains(target, "/") {
			return strings.TrimSuffix(strings.TrimSuffix(target, "/cmd/gentle-ai"), "/cmd")
		}
		if major := extractMajorVersion(target); major >= 4 {
			return fmt.Sprintf("%s/v%d", repo, major)
		}
	}
	return repo + "/v3"
}

// GentleAIImportPath returns the Go import path for the gentle-ai binary.
func GentleAIImportPath(tool ToolInfo, versionOrMajor ...string) string {
	return GentleAIModulePath(tool, versionOrMajor...) + "/cmd/gentle-ai"
}

func extractMajorVersion(version string) int {
	v := strings.TrimPrefix(strings.TrimSpace(version), "v")
	if i := strings.IndexAny(v, ".-"); i >= 0 {
		v = v[:i]
	}
	n, _ := strconv.Atoi(v)
	return n
}

// GentleAISourceInstallCommand returns the safe source-install fallback for an
// exact release, beta main build, or the latest release when version is empty.
// In beta, modulePath is strictly required and the verified commit SHA is pinned
// in the install target (@<SHA>).
func GentleAISourceInstallCommand(version string, modulePath ...string) (string, error) {
	version, module := strings.TrimSpace(version), ""
	if len(modulePath) > 0 {
		module = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(modulePath[0]), "/cmd/gentle-ai"), "/cmd")
	}
	if strings.HasPrefix(version, "main@") || version == "main" {
		sha := strings.TrimSpace(strings.TrimPrefix(version, "main@"))
		if module == "" || sha == "" || sha == "main" {
			return "", fmt.Errorf("cannot derive source install command for beta build %q: missing Go module path or commit SHA", version)
		}
		return fmt.Sprintf("go install %s/cmd/gentle-ai@%s", module, sha), nil
	}
	target := "latest"
	if version != "" && version != "latest" {
		target = "v" + strings.TrimPrefix(version, "v")
	}
	if module == "" {
		module = GentleAIModulePath(ToolInfo{}, version)
	}
	return fmt.Sprintf("go install %s/cmd/gentle-ai@%s", module, target), nil
}

// updateHint returns a platform-specific instruction string for updating the given tool.
func updateHint(tool ToolInfo, profile system.PlatformProfile) string {
	switch tool.Name {
	case "gentle-ai":
		return gentleAIHint(profile)
	case "engram":
		return engramHint(profile)
	case "gga":
		return ggaHint(profile)
	case "opencode-subagent-statusline", "opencode-sdd-engram-manage":
		return "gentle-ai upgrade updates ~/.config/opencode npm deps, clears this plugin's @latest cache, then requires OpenCode restart/reload"
	default:
		return ""
	}
}

func updateHintForOwnership(tool ToolInfo, profile system.PlatformProfile, ownership HomebrewOwnership) string {
	if profile.PackageManager == "brew" && ownership != HomebrewNone {
		return fmt.Sprintf("brew upgrade --%s %s", ownership, tool.Name)
	}
	return updateHint(tool, profile)
}

func openCodeRegisteredNotMaterializedHint(tool ToolInfo) string {
	pkg := strings.TrimSpace(tool.NpmPackage)
	if pkg == "" {
		pkg = tool.Name
	}
	return fmt.Sprintf("registered in ~/.config/opencode/tui.json; pending npm dependency materialization for %s. Run gentle-ai upgrade to install/update ~/.config/opencode dependencies, then restart or reload OpenCode; if it stays pending, check OpenCode logs for package or peer dependency errors.", pkg)
}

// gentleAIHint is the stable-channel instruction only. When the checker
// resolves a main-head beta target, applyBetaMainHeadStatus overrides this
// hint with GentleAISourceInstallCommand so the printed instruction installs
// the advertised target instead of the latest stable release.
func gentleAIHint(profile system.PlatformProfile) string {
	if profile.PackageManager == "brew" && homebrewPackageInstalled("gentle-ai") {
		return "brew upgrade gentle-ai"
	}

	switch profile.OS {
	case "linux":
		return "curl -fsSL https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/scripts/install.sh | bash"
	case "darwin":
		return "gentle-ai upgrade (downloads pre-built binary)"
	case "windows":
		cmd, _ := GentleAISourceInstallCommand("")
		return WindowsDistributionHoldMessage + " Install/update from source with Go 1.25.10+: " + cmd
	default:
		return ""
	}
}

func engramHint(profile system.PlatformProfile) string {
	if profile.PackageManager == "brew" && homebrewPackageInstalled("engram") {
		return "brew upgrade engram"
	}
	return "gentle-ai upgrade (downloads pre-built binary)"
}

func ggaHint(profile system.PlatformProfile) string {
	if profile.PackageManager == "brew" && homebrewPackageInstalled("gga") {
		return "brew upgrade gga"
	}
	return "See https://github.com/Gentleman-Programming/gentleman-guardian-angel"
}
