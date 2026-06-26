package knowledge

type Domain string

const (
	DomainETL         Domain = "etl"
	DomainGraphQL     Domain = "graphql"
	DomainAPI         Domain = "api"
	DomainDataQuality Domain = "data_quality"
	DomainBugs        Domain = "bugs"
	DomainPrompts     Domain = "prompts"
)

var DefaultArtifactFilenames = []string{
	"etl_patterns.yaml",
	"graphql_patterns.yaml",
	"api_patterns.yaml",
	"data_quality_patterns.yaml",
	"common_bugs.yaml",
	"successful_prompts.yaml",
	"project_profile.yaml",
}

type Artifact struct {
	Filename string
	Content  string
}

type Finding struct {
	Name       string
	Domain     Domain
	Keywords   []string
	Evidence   string
	SourcePath string
}

type Result struct {
	SessionsProcessed int
	ArtifactsCreated  int
	OutputDir         string
}

type ProjectProfile struct {
	Sessions           int
	ETLFindings        int
	GraphQLFindings    int
	APIFindings        int
	DataQualityFindings int
	BugFindings        int
	PromptFindings     int
}
