package service_manager

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"os"
	"os/exec"
	"time"
)


type ServiceManager interface {
	StartKafkaZookeeperContainers()
	StopKafkaZookeeperContainers()
	StartCanary()
}


type controller struct {
	ctx       context.Context
	container testcontainers.Container
	kafkaContainer *testcontainers.LocalDockerCompose
	isDefaultNetworkRunning bool
	host string
}


const (
	pathToDockerComposeImage        = "/compose-kafka-zookeeper.yaml"
	pathToMainMethod                = "../../cmd/main.go"
)


func (c *controller) setContext() {
	c.ctx = context.Background()
}


func (c *controller) StartKafkaZookeeperContainers() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := dir + pathToDockerComposeImage
	c.kafkaContainer = testcontainers.NewLocalDockerCompose([]string{path}, "someName")
	c.kafkaContainer.WithCommand([]string{"up", "-d"}).Invoke()

	log.Println("docker compose network created")
}


func (c *controller) StopKafkaZookeeperContainers() {
	c.kafkaContainer.Down()
}


func CreateManager() ServiceManager {
	var cManipulation ServiceManager = &controller{}
	return cManipulation
}


func (c *controller) StartCanary() {
	log.Println("starting canary")
	// Output is not necessary but
	go exec.Command("go", "run",  pathToMainMethod ).Run()

	// once canary starts it may take extra time to open endpoints
	time.Sleep(time.Second * 10)
}
