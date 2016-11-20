package models

type Vote struct {
	Name  string `json:"name" gorethink:"name"`
	Count int    `json:"count" gorethink:"count"`
}

type DockerContainer struct {
	ImageName   string `json:"imageName"`
	ContainerID string `json:"containerId"`
}
