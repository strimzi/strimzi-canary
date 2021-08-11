package service_manager

import (
	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
	"log"
	"os"
	"os/exec"
	"time"
)

// Implementation of Service Manager
type ServiceManager struct {
	CanaryConfig
	Paths
}

// Configurations of canary that is manipulated in e2e tests
type CanaryConfig struct {
	ReconcileIntervalTime string
	TopicTestName         string
	KafkaBrokerAddress    string

}

// paths to services ( kafka zookeeper docker compose, and Canary application main method)
type Paths struct {
	pathDockerComposeKafkaZookeeper string
	pathToCanaryMain string
}

const (
	canaryTestTopicName         = "__strimzi_canary_test_topicv123"
	kafkaBrokerAddress          = "127.0.0.1:9092"
	canaryReconcileIntervalTime = "1000"

	pathToDockerComposeImage    = "compose-kafka-zookeeper.yaml"
	pathToMainMethod            = "../cmd/main.go"
)

func (c *ServiceManager) StartKafkaZookeeperContainers() {
	log.Println("Starting kafka & Zookeeper")

	var cmd = exec.Command("docker-compose", "-f", c.pathDockerComposeKafkaZookeeper, "up", "-d" )
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	log.Println("Zookeeper and Kafka containers created")
	// after creation of containers we still have to wait for some time before successful communication with Kafka & Zookeper
	c.waitForBroker()
}

func (c *ServiceManager) StopKafkaZookeeperContainers() {
	var cmd = exec.Command("docker-compose", "-f", c.pathDockerComposeKafkaZookeeper, "down")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
}

func CreateManager() *ServiceManager {
	manager := &ServiceManager{}

	manager.ReconcileIntervalTime = canaryReconcileIntervalTime
	manager.TopicTestName = canaryTestTopicName
	manager.pathToCanaryMain = pathToMainMethod
	manager.pathDockerComposeKafkaZookeeper = pathToDockerComposeImage
	manager.KafkaBrokerAddress = kafkaBrokerAddress
	return manager
}

// Canary topic will not be immediately created, therefore it is up to tests to wait for its creation.
func (c *ServiceManager) StartCanary() {
	log.Println("Starting Canary")
	c.setUpCanaryParamsViaEnv()
	myCmd := exec.Command("go", "run",  c.pathToCanaryMain )

	if err := myCmd.Start(); err != nil {
		log.Fatal(err.Error())
	}
}

// per se it means waiting for container's broker to communicate correctly
func (c *ServiceManager) waitForBroker(){
	log.Println("start waiting for broker")
	timeout := time.After(10 * time.Second)
	brokerIsReadyChannel := make(chan bool)

	go func() {
		configuration := sarama.NewConfig()
		brokers := []string{c.KafkaBrokerAddress}

		for ;; {
			// if we can create Cluster Admin, broker can communicate
			_, err := sarama.NewClusterAdmin(brokers, configuration)
			if err != nil {
				log.Println("waiting for broker's start")
				time.Sleep(time.Millisecond * 500)
				continue
			}
			break
		}

		brokerIsReadyChannel <- true
	}()

	select {
	case <-timeout:
		log.Fatal("Broker isn't ready within expected timeout")
	case <-brokerIsReadyChannel:
		log.Println("Container (Broker) is ready")
	}
}

func (c *ServiceManager) setUpCanaryParamsViaEnv(){
	log.Println("Setting up environment variables")
	os.Setenv(config.ReconcileIntervalEnvVar, c.ReconcileIntervalTime)
	os.Setenv(config.TopicEnvVar, c.TopicTestName)
	os.Setenv(config.BootstrapServersEnvVar, c.KafkaBrokerAddress)

}