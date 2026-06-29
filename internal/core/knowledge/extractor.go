package knowledge

import "strings"

type Extractor struct{}

func (Extractor) Extract(sessions []NormalizedSession) ExtractionResult {
	result := ExtractionResult{
		Profile: ProjectProfile{Sessions: len(sessions)},
	}

	for _, session := range sessions {
		lower := strings.ToLower(session.NormalizedMarkdown)
		sourcePath := strings.ToLower(session.SourcePath + " " + session.NormalizedPath)
		text := cleanEvidenceText(session.NormalizedMarkdown)

		if isETLSignal(lower, sourcePath) {
			result.Profile.ETLSignals++
			result.ETLPatterns = append(result.ETLPatterns, Finding{Name: "etl_validation", Domain: "etl", Evidence: evidence(text, etlNeedle(lower)), SourcePath: session.NormalizedPath})
		}
		if isGraphQLSignal(lower, sourcePath) {
			result.Profile.GraphQLSignals++
			result.GraphQLPatterns = append(result.GraphQLPatterns, Finding{Name: "graphql_functional_testing", Domain: "graphql", Evidence: evidence(text, graphqlNeedle(lower)), SourcePath: session.NormalizedPath})
		}
		if isAPISignal(lower) {
			result.Profile.APISignals++
			result.APIPatterns = append(result.APIPatterns, Finding{Name: "api_contract_testing", Domain: "api", Evidence: evidence(text, apiNeedle(lower)), SourcePath: session.NormalizedPath})
		}
		if isDataQualitySignal(lower) {
			result.Profile.DQSignals++
			result.DataQualityPatterns = append(result.DataQualityPatterns, Finding{Name: "data_quality_validation", Domain: "data_quality", Evidence: evidence(text, dqNeedle(lower)), SourcePath: session.NormalizedPath})
		}
		if isFailureSignal(lower) {
			result.CommonBugs = append(result.CommonBugs, Finding{Name: "common_failure_signal", Domain: "bugs", Evidence: evidence(text, failureNeedle(lower)), SourcePath: session.NormalizedPath})
		}
		if isPromptSignal(lower) {
			result.SuccessfulPrompts = append(result.SuccessfulPrompts, Finding{Name: "successful_prompt_candidate", Domain: "prompts", Evidence: evidence(text, promptNeedle(lower)), SourcePath: session.NormalizedPath})
		}
	}

	return result
}
