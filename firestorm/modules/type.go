package modules

type Module struct {
	Name    string
	Version string
	Files   map[string]string
}

func NewPackage(name string, version string) Module {
	return Module{
		Name:    name,
		Version: version,
		Files:   loadFileList(name, version),
	}
}
