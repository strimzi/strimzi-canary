package container_manager

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"os"
	"time"
)


type ContainerManager interface {
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


func (c *controller) setContext() {
	c.ctx = context.Background()
}


func (c *controller) StartKafkaZookeeperContainers() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := dir + "/compose-kafka-zookeeper.yaml"
	c.kafkaContainer = testcontainers.NewLocalDockerCompose([]string{path}, "someName")
	c.kafkaContainer.WithCommand([]string{"up", "-d"}).Invoke()

	log.Println("docker compose network created")
}


func (c *controller) StopKafkaZookeeperContainers() {
	c.kafkaContainer.Down()
}


func CreateManager() ContainerManager {
	var cManipulation ContainerManager = &controller{}
	return cManipulation
}


func (c *controller) StartCanary() {
	log.Println("starting canary")
	go createCanary()
	// once canary starts it may take extra time to open endpoints
	time.Sleep(time.Second * 1)
}
