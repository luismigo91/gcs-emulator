package model

type BuildTrigger struct {
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	Filename      string         `json:"filename,omitempty"`
	Substitutions map[string]string `json:"substitutions,omitempty"`
	TriggerTemplate *TriggerTemplate `json:"triggerTemplate,omitempty"`
	Github        *GitHubEventConfig `json:"github,omitempty"`
}

type TriggerTemplate struct {
	ProjectID  string `json:"projectId"`
	RepoName   string `json:"repoName"`
	BranchName string `json:"branchName"`
}

type GitHubEventConfig struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	PullRequest *PullRequestFilter `json:"pullRequest,omitempty"`
	Push        *PushFilter        `json:"push,omitempty"`
}

type PullRequestFilter struct {
	Branch string `json:"branch"`
}

type PushFilter struct {
	Branch string `json:"branch"`
}

type Build struct {
	ID         string   `json:"id"`
	Status     string   `json:"status"`
	TriggerID  string   `json:"triggerId,omitempty"`
	StartTime  string   `json:"startTime,omitempty"`
	Logs       []string `json:"logs,omitempty"`
}
