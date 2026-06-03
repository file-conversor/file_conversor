// internal/deps/ghostscript_dep.go

package deps

var GhostScriptDep = &Dependency{
	Name:     "Ghostscript",
	Binary:   "gs",
	Homepage: "https://www.ghostscript.com/",
	License:  AGPLv3,
	Packages: []Package{
		{
			Name:    "ghostscript",
			Manager: BrewPkgMgr,
		},
		{
			Name:    "ghostscript",
			Manager: ScoopPkgMgr,
		},
	},
}
