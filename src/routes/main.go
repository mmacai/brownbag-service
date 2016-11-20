package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"../db"
	"../models"
)

func getDockerClientInfo() []models.DockerContainer {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	dockerContainers := []models.DockerContainer{}

	var serviceImageName string = "brownbag-service"

	if os.Getenv("IMAGE_NAME_SERVICE") != "" {
		serviceImageName = os.Getenv("IMAGE_NAME_SERVICE")
	}

	fmt.Println("Available containers", containers)

	for _, container := range containers {
		if strings.Contains(container.Image, serviceImageName) {
			dockerContainers = append(dockerContainers, models.DockerContainer{ImageName: container.Image, ContainerID: container.ID[:10]})
		}
	}

	return dockerContainers
}

func GetVotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	votes := db.Read()
	json.NewEncoder(w).Encode(votes)
}

func UpdateVotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get body from request
	body, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		panic(bodyErr)
	}

	// Parse json into Votes struct
	var response models.Vote
	responseErr := json.Unmarshal(body, &response)

	if responseErr != nil {
		panic(responseErr)
	}

	db.Update(response)
	dockerContainers := getDockerClientInfo()

	// Response to client with newly defined Votes instance
	json.NewEncoder(w).Encode(dockerContainers)
}

func GetContainerInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	dockerContainers := getDockerClientInfo()

	// Response to client with newly defined Votes instance
	json.NewEncoder(w).Encode(dockerContainers)
}
