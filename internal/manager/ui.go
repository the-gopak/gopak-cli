package manager

type PackageKey struct{
	Source string
	Name string
	Kind string
}

type VersionStatus struct{
	Installed string
	Available string
}

type UpdateReporter interface{
	OnInit(groups map[string][]string)
	OnInstalled(k PackageKey, version string)
	OnAvailable(k PackageKey, version string)
	OnPhaseDone(name string)
	ConfirmProceed() bool
	OnUpdateStart()
	OnPackageUpdated(k PackageKey, ok bool, errMsg string)
	OnDone()
}
