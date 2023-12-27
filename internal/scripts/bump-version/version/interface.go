package version

// nolint:revive
// VersionData represents version info needed to generate version files
type VersionData struct {
	VERSION string
}

//go:generate mockgen -destination=../../../mock/scripts/bump-version/version/version.go -package=mock_version . VersionControl,VersionGenerator

// nolint:revive
// VersionControl interface for interacting with version control systems
type VersionControl interface {
	Add(filePath string) error
	Commit(message string) error
	Tag(version string) error
}

// nolint:revive
// VersionGenerator interface for generating version files
type VersionGenerator interface {
	Generate(data VersionData) error
}
