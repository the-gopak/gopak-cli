package manager

type PackageKey struct {
	Source string
	Name   string
	Kind   string
}

type VersionStatus struct {
	Installed string
	Available string
}
