package model

type Repository struct {
	Name       string `json:"name"`
	Format     string `json:"format"`
	Description string `json:"description,omitempty"`
}

type DockerImage struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
	URI  string   `json:"uri"`
}
