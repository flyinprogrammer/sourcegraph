package catalog

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Component struct {
	Kind          string
	Name          string
	Description   string
	Lifecycle     string
	SourceRepo    api.RepoName
	SourceCommit  api.CommitID
	SourcePaths   []string
	UsagePatterns []UsagePattern
	APIDefPath    string

	// Edges
	DependsOn []string
	OwnedBy   string
}

type UsagePattern struct {
	Query string
}

type Group struct {
	Name        string
	Title       string
	ParentGroup string
	Members     []string
}

type Edge struct {
	Type    EntityRelationType
	Out, In string // entity names
}

type EntityRelationType string

const (
	DependsOnRelation    = "DEPENDS_ON"
	DependencyOfRelation = "DEPENDENCY_OF"

	PartOfRelation  = "PART_OF"
	HasPartRelation = "HAS_PART"

	MemberOfRelation  = "MEMBER_OF"
	HasMemberRelation = "HAS_MEMBER"

	//

	LifecycleProduction   = "PRODUCTION"
	LifecycleExperimental = "EXPERIMENTAL"
)

func newQueryUsagePattern(query string) UsagePattern {
	return UsagePattern{
		Query: `count:4 repo:^github\.com/sourcegraph/sourcegraph$ ` + query,
	}
}

func Data() ([]Component, []Group, []Edge) {
	var edges []Edge

	const (
		sourceRepo   = "github.com/sourcegraph/sourcegraph"
		sourceCommit = "2ada4911722e2c812cc4f1bbfb6d5d1756891392"
	)
	components := []Component{
		{
			Kind:         "SERVICE",
			Name:         "frontend",
			Description:  "Serves the web app, public APIs, and internal APIs.",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/frontend", "enterprise/cmd/frontend"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`\.api/graphql patterntype:literal`),
				newQueryUsagePattern(`lang:typescript requestGraphQL\(|useConnection\(|useQuery\(|gql` + "`" + ` patterntype:regexp`),
			},
			APIDefPath: "cmd/frontend/graphqlbackend/schema.graphql",
			DependsOn:  []string{"gitserver", "client-web", "repo-updater", "executor", "github-proxy", "precise-code-intel-worker", "query-runner", "searcher", "sitemap", "symbols"},
			OwnedBy:    "search-core",
		},
		{
			Kind:         "SERVICE",
			Name:         "gitserver",
			Description:  "Mirrors repositories from their code host.",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/gitserver"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bgitserver\.Client\b patterntype:regexp`),
				newQueryUsagePattern(`lang:go \bgit\.[A-Z]\w+\(ctx, patterntype:regexp`),
				newQueryUsagePattern(`lang:go "github.com/sourcegraph/sourcegraph/internal/vcs/git" patterntype:literal`),
			},
			APIDefPath: "internal/gitserver/protocol/gitserver.go",
			DependsOn:  []string{"repo-updater"},
			OwnedBy:    "repo-mgmt",
		},
		{
			Kind:         "SERVICE",
			Name:         "repo-updater",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/repo-updater", "enterprise/cmd/repo-updater"},
			DependsOn:    []string{"gitserver", "github-proxy"},
			OwnedBy:      "repo-mgmt",
		},
		{
			Kind:         "SERVICE",
			Name:         "searcher",
			Description:  "Provides on-demand unindexed search for repositories",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/searcher"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bsearcher\.Search\( patterntype:regexp`),
			},
			DependsOn: []string{"gitserver"},
			OwnedBy:   "search-core",
		},
		{
			Kind:         "SERVICE",
			Name:         "executor",
			Lifecycle:    LifecycleExperimental,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"enterprise/cmd/executor"},
			DependsOn:    []string{"frontend"},
			OwnedBy:      "code-intel",
		},
		{
			Kind:         "SERVICE",
			Name:         "precise-code-intel-worker",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"enterprise/cmd/precise-code-intel-worker"},
			DependsOn:    []string{"frontend", "worker"},
			OwnedBy:      "code-intel",
		},
		{
			Kind:         "SERVICE",
			Name:         "github-proxy",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/github-proxy"},
			OwnedBy:      "repo-mgmt",
		},
		{
			Kind:         "SERVICE",
			Name:         "query-runner",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/query-runner"},
			DependsOn:    []string{"frontend"},
			OwnedBy:      "search-product",
		},
		{
			Kind:         "SERVICE",
			Name:         "worker",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/worker", "enterprise/cmd/worker"},
		},
		{
			Kind:         "SERVICE",
			Name:         "server",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/server", "enterprise/cmd/server"},
			DependsOn:    []string{"frontend", "repo-updater", "symbols", "query-runner", "gitserver"},
		},
		{
			Kind:         "SERVICE",
			Name:         "symbols",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/symbols"},
			DependsOn:    []string{"gitserver", "frontend"},
			OwnedBy:      "code-intel",
		},
		{
			Kind:         "SERVICE",
			Name:         "sitemap",
			Lifecycle:    LifecycleExperimental,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/sitemap"},
			DependsOn:    []string{"frontend"},
		},
		{
			Kind:         "TOOL",
			Name:         "sg",
			Description:  "The Sourcegraph developer tool",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"dev/sg"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:markdown ` + "`" + `sg[` + "`" + `\s] patterntype:regexp`),
				newQueryUsagePattern(`lang:markdown (^|\s*\$ )sg\s patterntype:regexp`),
			},
		},
		{
			Kind:        "TOOL",
			Name:        "src-cli",
			Description: "Sourcegraph CLI",
			Lifecycle:   LifecycleProduction,
			// Only the gitlab mirror of this repo is loaded by the default dev-private config.
			SourceRepo:    "gitlab.sgdev.org/sourcegraph/src-cli",
			SourceCommit:  "4a4341bc1c53fc5306f09bdcb31e8892ee40e6c7",
			SourcePaths:   []string{"."},
			UsagePatterns: []UsagePattern{
				// newQueryUsagePattern(`lang:markdown ` + "`" + `src[` + "`" + `\s] patterntype:regexp`),
				// newQueryUsagePattern(`lang:markdown (^|\s*\$ )src\s patterntype:regexp`),
			},
			DependsOn: []string{"frontend"},
		},
		{
			Kind:         "LIBRARY",
			Name:         "client-web",
			Description:  "Main web app UI",
			Lifecycle:    LifecycleExperimental,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/web"},
			DependsOn:    []string{"extension-api", "client-shared", "wildcard", "frontend"},
			OwnedBy:      "frontend-platform",
		},
		{
			Kind:         "LIBRARY",
			Name:         "client-browser",
			Description:  "Browser extension and native code host extension",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/browser"},
			DependsOn:    []string{"extension-api", "client-shared", "wildcard", "frontend"},
		},
		{
			Kind:         "LIBRARY",
			Name:         "client-shared",
			Description:  "Frontend code shared by the web app and browser extension",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/shared"},
			DependsOn:    []string{"extension-api", "wildcard", "frontend"},
			OwnedBy:      "frontend-platform",
		},
		{
			Kind:         "LIBRARY",
			Name:         "wildcard",
			Description:  "The Wildcard component library is a collection of design-approved reusable components that are suitable for use within the Sourcegraph codebase.",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/wildcard"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:typescript import @sourcegraph/wildcard patterntype:regexp`),
			},
			OwnedBy: "frontend-platform",
		},
		{
			Kind:         "LIBRARY",
			Name:         "extension-api",
			Description:  "Public TypeScript API for Sourcegraph extensions",
			Lifecycle:    LifecycleProduction,
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/extension-api"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:typescript import from ['"]sourcegraph['"] patterntype:regexp`),
			},
		},
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })
	for _, c := range components {
		for _, dependsOn := range c.DependsOn {
			edges = append(edges, Edge{
				Type: DependsOnRelation,
				Out:  c.Name,
				In:   dependsOn,
			})
			edges = append(edges, Edge{
				Type: DependencyOfRelation,
				Out:  dependsOn,
				In:   c.Name,
			})
		}
	}

	groups := []Group{
		{
			Name:    "code-graph",
			Title:   "Code graph",
			Members: []string{"yink-teo"},
		},
		{
			Name:        "search-core",
			Title:       "Search core",
			ParentGroup: "code-graph",
			Members:     []string{"jeffwarner", "ryanhitchman", "stefanhengl", "tomas", "keegancs"},
		},
		{
			Name:        "search-product",
			Title:       "Search product",
			ParentGroup: "code-graph",
			Members:     []string{"lguychard", "fkling", "ccheek", "rok", "juliana", "rijnard"},
		},
		{
			Name:        "code-intel",
			Title:       "Code intelligence",
			ParentGroup: "code-graph",
			Members:     []string{"oconvey", "vgandhi", "cesarj", "chrismwendt", "teej", "olaf", "noahsc", "efritz"},
		},
		{
			Name:        "batch-changes",
			Title:       "Batch Changes",
			ParentGroup: "code-graph",
			Members:     []string{"chris-pine", "kelli-rockwell", "adeola-akinsiku", "adam-harvey", "erik-seliger", "thorsten-ball"},
		},
		{
			Name:        "code-insights",
			Title:       "Code insights",
			ParentGroup: "code-graph",
			Members:     []string{"felix-becker", "cristina-birkel", "justin-boyson", "coury-clark", "vova-kulikov"},
		},

		{
			Name:    "enablement",
			Title:   "Enablement",
			Members: []string{"jean-du-plessis"},
		},
		{
			Name:        "repo-mgmt",
			Title:       "Repo management",
			ParentGroup: "enablement",
			Members:     []string{"jplahn", "indrag", "rslade", "mweitzel", "alex-ostrikov"},
		},
		{
			Name:        "delivery",
			Title:       "Delivery",
			ParentGroup: "enablement",
			Members:     []string{"jean-du-plessis", "kevin-wojkovich", "crystal-augustus"},
		},
		{
			Name:        "dev-experience",
			Title:       "Dev experience",
			ParentGroup: "enablement",
			Members:     []string{"kristen-stretch", "jh-chabran", "robert-lin", "dave-try"},
		},
		{
			Name:        "frontend-platform",
			Title:       "Frontend platform",
			ParentGroup: "enablement",
			Members:     []string{"patrick-dubroy", "valeryb", "tomross"},
		},

		{
			Name:    "cloud",
			Title:   "Cloud",
			Members: []string{"billcreager"},
		},
		{
			Name:        "growth",
			Title:       "Growth",
			ParentGroup: "cloud",
			Members:     []string{"stephengutekanst"},
		},
		{
			Name:        "extensibility",
			Title:       "Extensibility",
			ParentGroup: "cloud",
			Members:     []string{"murat-sutunc", "erzhan-torokulov", "beatrix-woo", "tharuntej-kandala"},
		},
		{
			Name:        "security",
			Title:       "Security",
			ParentGroup: "cloud",
			Members:     []string{"diego-comas", "lauren-chapman", "david-sandy", "mohammad-alam", "andre-eleuterio"},
		},
		{
			Name:        "devops",
			Title:       "DevOps",
			ParentGroup: "cloud",
			Members:     []string{"jennifer-mitchell", "dax-mcdonald"},
		},
		{
			Name:        "cloud-saas",
			Title:       "Cloud SaaS",
			ParentGroup: "cloud",
			Members:     []string{"rafal-leszczynski", "rafal-gajdulewicz", "milan-freml", "artem-ruts", "joe-cheng"},
		},
	}
	for _, g := range groups {
		if g.ParentGroup != "" {
			edges = append(edges, Edge{
				Type: PartOfRelation,
				Out:  g.Name,
				In:   g.ParentGroup,
			})
			edges = append(edges, Edge{
				Type: HasPartRelation,
				Out:  g.ParentGroup,
				In:   g.Name,
			})
		}
	}

	return components, groups, edges
}

func Components() []Component {
	components, _, _ := Data()
	return components
}

func ComponentByName(name string) *Component {
	for _, c := range Components() {
		if c.Name == name {
			return &c
		}
	}
	return nil
}

func Groups() []Group {
	_, groups, _ := Data()
	return groups
}

func GroupByName(name string) *Group {
	for _, g := range Groups() {
		if g.Name == name {
			return &g
		}
	}
	return nil
}
