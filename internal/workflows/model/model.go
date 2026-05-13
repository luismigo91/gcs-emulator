package model

type Workflow struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	SourceContents  string `json:"sourceContents,omitempty"`
	State           string `json:"state,omitempty"`
}

type Execution struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Result   string `json:"result,omitempty"`
	Error    string `json:"error,omitempty"`
}
