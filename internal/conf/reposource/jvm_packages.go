package reposource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MavenModule struct {
	GroupID    string
	ArtifactID string
}

func (m *MavenModule) Equal(other *MavenModule) bool {
	return m == other || (m != nil && other != nil && *m == *other)
}

func (m *MavenModule) IsJDK() bool {
	return m.Equal(jdkModule())
}

func (m *MavenModule) MatchesDependencyString(dependency string) bool {
	return strings.HasPrefix(dependency, fmt.Sprintf("%s:%s:", m.GroupID, m.ArtifactID))
}

func (m *MavenModule) CoursierSyntax() string {
	return fmt.Sprintf("%s:%s", m.GroupID, m.ArtifactID)
}

func (m *MavenModule) PackageSyntax() string {
	return m.CoursierSyntax()
}

func (m *MavenModule) SortText() string {
	return m.CoursierSyntax()
}

func (m *MavenModule) LsifJavaKind() string {
	if m.IsJDK() {
		return "jdk"
	}
	return "maven"
}

func (m *MavenModule) Description() string { return "" }

type MavenMetadata struct {
	Module *MavenModule
}

func (m *MavenModule) RepoName() api.RepoName {
	if m.IsJDK() {
		return "jdk"
	}
	return api.RepoName(fmt.Sprintf("maven/%s/%s", m.GroupID, m.ArtifactID))
}

func (m *MavenModule) CloneURL() string {
	cloneURL := url.URL{Path: string(m.RepoName())}
	return cloneURL.String()
}

// See [NOTE: Dependency-terminology]
type MavenPackageVersion struct {
	*MavenModule
	Version string
}

func (d *MavenPackageVersion) Equal(o *MavenPackageVersion) bool {
	return d == o || (d != nil && o != nil &&
		d.MavenModule.Equal(o.MavenModule) &&
		d.Version == o.Version)
}

func (d *MavenPackageVersion) Less(other PackageVersion) bool {
	o := other.(*MavenPackageVersion)

	if d.MavenModule.Equal(o.MavenModule) {
		return versionGreaterThan(d.Version, o.Version)
	}

	// TODO: This SortText method is quite inefficient and allocates.
	return d.SortText() > o.SortText()
}

func (d *MavenPackageVersion) PackageVersionSyntax() string {
	return fmt.Sprintf("%s:%s", d.PackageSyntax(), d.Version)
}

func (d *MavenPackageVersion) String() string {
	return d.PackageVersionSyntax()
}

func (d *MavenPackageVersion) PackageVersion() string {
	return d.Version
}

func (d *MavenPackageVersion) Scheme() string {
	return "semanticdb"
}

func (d *MavenPackageVersion) GitTagFromVersion() string {
	return "v" + d.Version
}

func (d *MavenPackageVersion) LsifJavaDependencies() []string {
	if d.IsJDK() {
		return []string{}
	}
	return []string{d.PackageVersionSyntax()}
}

// ParseMavenPackageVersion parses a dependency string in the Coursier format
// (colon seperated group ID, artifact ID and an optional version) into a MavenPackageVersion.
func ParseMavenPackageVersion(dependency string) (*MavenPackageVersion, error) {
	dep := &MavenPackageVersion{MavenModule: &MavenModule{}}

	switch ps := strings.Split(dependency, ":"); len(ps) {
	case 3:
		dep.Version = ps[2]
		fallthrough
	case 2:
		dep.MavenModule.GroupID = ps[0]
		dep.MavenModule.ArtifactID = ps[1]
	default:
		return nil, errors.Newf("dependency %q must contain at least one colon ':' character", dependency)
	}

	return dep, nil
}

// ParseMavenPackageFromRepoName is a convenience function to parse a repo name in a
// 'maven/<name>' format into a MavenPackageVersion.
func ParseMavenPackageFromRepoName(name string) (*MavenPackageVersion, error) {
	if name == "jdk" {
		return &MavenPackageVersion{MavenModule: jdkModule()}, nil
	}

	dep := strings.ReplaceAll(strings.TrimPrefix(name, "maven/"), "/", ":")
	if len(dep) == len(name) {
		return nil, errors.New("invalid maven dependency repo name, missing maven/ prefix")
	}

	return ParseMavenPackageVersion(dep)
}

// jdkModule returns the module for the Java standard library (JDK). This module
// is technically not a "maven module" because the JDK is not published as a
// Maven library. The only difference that's relevant for Sourcegraph is that we
// use a different coursier command to download JDK sources compared to normal
// maven modules:
// - JDK sources: `coursier java-home --jvm VERSION`
// - Maven sources: `coursier fetch MAVEN_MODULE:VERSION --classifier=sources`
// Since the difference is so small, the code is easier to read/maintain if we
// model the JDK as a Maven module.
func jdkModule() *MavenModule {
	return &MavenModule{
		GroupID:    "jdk",
		ArtifactID: "jdk",
	}
}
