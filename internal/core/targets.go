package core

// Target describes a single cleanable filesystem entry.
type Target struct {
	Name      string   // exact name or wildcard (e.g. "*.pyc")
	IsDir     bool     // whether this target is expected to be a directory
	HintFiles []string // if set, only match when one of these files exists in an ancestor
}

// Category groups related targets and carries metadata.
type Category struct {
	Name        string
	Description string
	Targets     []Target
	Risky       bool // risky categories are off by default and warned about
}

// DefaultCategories is the full set of categories the cleaner knows about.
//
// Risky targets (Cargo.lock, Go vendor/bin/pkg, .vscode/.idea) are flagged
// because deleting them can break a project. They are opt-in.
var DefaultCategories = []Category{
	{
		Name:        "Node.js",
		Description: "Node dependencies and package manager caches",
		Targets: []Target{
			{Name: "node_modules", IsDir: true},
			{Name: ".npm", IsDir: true},
			{Name: ".yarn", IsDir: true},
			{Name: "npm-debug.log"},
			{Name: "yarn-error.log"},
		},
	},
	{
		Name:        "Java",
		Description: "Maven/Gradle build output and caches",
		Targets: []Target{
			{Name: "target", IsDir: true, HintFiles: []string{"pom.xml", "build.gradle", "build.gradle.kts"}},
			{Name: "build", IsDir: true, HintFiles: []string{"pom.xml", "build.gradle", "build.gradle.kts", "CMakeLists.txt"}},
			{Name: ".gradle", IsDir: true},
		},
	},
	{
		Name:        "Python",
		Description: "Bytecode caches and virtual environments",
		Targets: []Target{
			{Name: "__pycache__", IsDir: true},
			{Name: ".pytest_cache", IsDir: true},
			{Name: "*.pyc"},
			{Name: "*.pyo"},
			{Name: ".tox", IsDir: true},
			{Name: "venv", IsDir: true},
			{Name: ".venv", IsDir: true},
			{Name: "env", IsDir: true},
		},
	},
	{
		Name:        "Rust",
		Description: "Cargo build output (target directory)",
		Targets: []Target{
			{Name: "target", IsDir: true, HintFiles: []string{"Cargo.toml"}},
		},
	},
	{
		Name:        "Go",
		Description: "Go build output",
		Targets: []Target{
			{Name: "bin", IsDir: true, HintFiles: []string{"go.mod"}},
			{Name: "pkg", IsDir: true, HintFiles: []string{"go.mod"}},
		},
	},
	{
		Name:        "C/C++",
		Description: "C/C++ build artifacts and object files",
		Targets: []Target{
			{Name: "cmake-build-debug", IsDir: true},
			{Name: "cmake-build-release", IsDir: true},
			{Name: "*.o"},
			{Name: "*.obj"},
		},
	},
	{
		Name:        "Docker",
		Description: "Docker local data and overrides",
		Targets: []Target{
			{Name: ".docker", IsDir: true},
			{Name: "docker-compose.override.yml"},
		},
	},
	{
		Name:        "IDE",
		Description: "Editor and OS metadata files",
		Targets: []Target{
			{Name: "*.swp"},
			{Name: "*.swo"},
			{Name: ".DS_Store"},
			{Name: "Thumbs.db"},
		},
	},
	{
		Name:        "Build",
		Description: "Generic build output and cache directories",
		Targets: []Target{
			{Name: "dist", IsDir: true},
			{Name: "out", IsDir: true},
			{Name: "output", IsDir: true},
			{Name: "release", IsDir: true},
			{Name: "debug", IsDir: true},
			{Name: ".cache", IsDir: true},
			{Name: "tmp", IsDir: true},
			{Name: "temp", IsDir: true},
		},
	},
	// Note: "dist", "out", "tmp", "temp" are common enough to match without
	// hints — they're almost always build/temp artifacts. "build" and
	// "target" are restricted via HintFiles because they're very generic
	// names that could be legitimate directories in non-project contexts.
	// Risky categories: deleting these can break projects. Opt-in only.
	{
		Name:        "Risky",
		Description: "Targets that may break projects (Cargo.lock, vendor, IDE configs)",
		Risky:       true,
		Targets: []Target{
			{Name: "Cargo.lock"},
			{Name: "vendor", IsDir: true},
			{Name: ".vscode", IsDir: true},
			{Name: ".idea", IsDir: true},
		},
	},
}

// CategoryByName returns a category lookup map.
func CategoryByName() map[string]Category {
	m := make(map[string]Category, len(DefaultCategories))
	for _, c := range DefaultCategories {
		m[c.Name] = c
	}
	return m
}
