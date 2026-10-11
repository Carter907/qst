package graph

type SubGuideRelation struct {
	Guide     string `yaml:"guide"`
	Adherence string `yaml:"adherence"`
	Segment   string `yaml:"segment"`
}

type GuideMetadata struct {
	Prerequisites []string           `yaml:"prerequisites"`
	SubGuides     []SubGuideRelation `yaml:"sub_guides"`
	Scope         string             `yaml:"scope"`
	Tags          []string           `yaml:"tags"`
}

type Guide struct {
	ID         string
	Path       string
	Metadata   GuideMetadata
	LineCount  int
	HasContent bool
	Body       string
}

type Manifest struct {
	Title            string       `yaml:"title"`
	Description      string       `yaml:"description"`
	Scopes           []string     `yaml:"scopes"`
	Adherences       []string     `yaml:"adherences"`
	RelaxedSubguides bool         `yaml:"relaxed_subguides"`
	StrictCoverage   bool         `yaml:"strict_coverage"`
	RequireSubguides bool         `yaml:"require_subguides"`
	Tours            []TourConfig `yaml:"tours"`
}

type TourConfig struct {
	Name   string   `yaml:"name"`
	Guides []string `yaml:"guides"`
}
