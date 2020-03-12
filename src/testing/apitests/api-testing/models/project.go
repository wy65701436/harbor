package models

// Project : For /api/projects
type Project struct {
	Name     string    `json:"project_name"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

// metadata : metadata for project
type Metadata struct {
	AccessLevel string `json:"public"`
}

// ExistingProject : For /api/projects?name=***
type ExistingProject struct {
	Name string `json:"name"`
	ID   int    `json:"project_id"`
}
