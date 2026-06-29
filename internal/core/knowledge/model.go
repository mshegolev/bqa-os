package knowledge

type Artifact struct {
	Filename string
	Content  string
}

type Finding struct {
	Name       string
	Domain     string
	Evidence   string
	SourcePath string
}

type Result struct {
	SessionsProcessed int
	ArtifactsCreated  int
	OutputDir         string
}

type ProjectProfile struct {
	Sessions       int
	ETLSignals     int
	GraphQLSignals int
	APISignals     int
	DQSignals      int
}
