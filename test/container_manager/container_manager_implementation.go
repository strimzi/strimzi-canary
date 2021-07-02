package container_manager

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"os"
	"time"
)

type controller struct {
	ctx       context.Context
	container testcontainers.Container
	kafkaContainer *testcontainers.LocalDockerCompose
	isDefaultNetworkRunning bool
	host string
}

// constants for evaluating occupying of default ports for Kafka and Zookeeper
func getTakenPorts() []string {
	return []string{"2181", "9092"}
}

func (c *controller) setContext() {
	c.ctx = context.Background()
}


func (c *controller) StartDefaultZookeeperKafkaNetwork() {
	var err error
	checkPortAvailability(c.host,  getTakenPorts())

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := dir + "/compose-kafka-zookeeper.yaml"
	c.kafkaContainer = testcontainers.NewLocalDockerCompose([]string{path}, "someName")
	c.kafkaContainer.WithCommand([]string{"up", "-d"}).Invoke()

	log.Println("docker compose network created")
}


func (c *controller) StopDefaultZookeeperKafkaNetwork() {
	port, err := checkPortAvailability(c.host, getTakenPorts())
	if err != nil {
		log.Println("port already closed: ", port)
	}
	c.kafkaContainer.Down()
}

func CreateManager() ContainerManager {
	var cManipulation ContainerManager = &controller{}
	return cManipulation

}

func (c *controller) StartCanary() {
	log.Println("starting canary")
	// zookeeper may take extra seconds to start before canary
	time.Sleep(time.Second * 5)
	go createCanary()
	// once canary starts it may take extra time tu spin up server
	time.Sleep(time.Second * 5)
}

