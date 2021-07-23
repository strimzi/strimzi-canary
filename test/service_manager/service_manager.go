package service_manager

import (
	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
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

// Implementation of Service Manager
type Controller struct {
	CanaryConfig
	Paths
}

// Configurations of canary that is manipulated in e2e tests
type CanaryConfig struct {
	RetentionTime string
	kafkaBrokerAddress string
	TopicTestName string
}

// paths to services ( kafka zookeeper docker compose, and Canary application main method)
type Paths struct {
	pathDockerComposeKafkaZookeeper string
	pathToCanaryMain string
}

const (
	canaryTestTopicName      = "__strimzi_canary_test"
	kafkaBrokerAddress       = "127.0.0.1:9092"
	canaryRetentionTime      = "1000"

	pathToDockerComposeImage = "test/compose-kafka-zookeeper.yaml"
	pathToMainMethod         = "cmd/main.go"
)

func (c *Controller) StartKafkaZookeeperContainers() {
	log.Println("Starting kafka & Zookeeper")

	var cmd = exec.Command("docker-compose", "-f", c.pathDockerComposeKafkaZookeeper, "up", "-d" )
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	log.Println("Zookeeper and Kafka containers created")
	// after creation of containers we still have to wait for some time before successful communication with Kafka & Zookeper
	c.waitForBroker()
}

func (c *Controller) StopKafkaZookeeperContainers() {
	var cmd = exec.Command("docker-compose", "-f", c.pathDockerComposeKafkaZookeeper, "down")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
}

func CreateManager() ServiceManager {
	manager := &Controller{}
	manager.kafkaBrokerAddress = kafkaBrokerAddress
	manager.RetentionTime = canaryRetentionTime
	manager.TopicTestName = canaryTestTopicName
	manager.pathToCanaryMain = pathToMainMethod
	manager.pathDockerComposeKafkaZookeeper = pathToDockerComposeImage
	return manager
}

func (c *Controller) StartCanary() {
	log.Println("Starting Canary")
	c.setUpCanaryParamsViaEnv()
	myCmd := exec.Command("go", "run",  c.pathToCanaryMain )
	//myCmd.Stdout = os.Stdout
	//myCmd.Stderr = os.Stderr

	if err := myCmd.Start(); err != nil {
		log.Fatal(err.Error())
	}
	// before we start to consider canary prepared we wait for it to create expected topic
	c.waitForCanarySetUp()

}

// per se it means waiting for container's broker to communicate correctly
func (c *Controller) waitForBroker(){
	log.Println("start waiting for broker")
	timeout := time.After(10 * time.Second)
	brokerIsReadyChannel := make(chan bool)

	go func() {
		configuration := sarama.NewConfig()
		brokers := []string{c.kafkaBrokerAddress}

		for ;; {
			consumer, err := sarama.NewConsumer(brokers, configuration)
			if err != nil {
				log.Println("waiting for broker's start")
				time.Sleep(time.Millisecond * 500)
				continue
			}
			// if we can create consumer broker can communicate
			consumer.Close()
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

// waiting while canary create new topic on kafka broker.
func (c *Controller) waitForCanarySetUp(){
	log.Println("start waiting for canary setup")
	timeout := time.After(10 * time.Second)
	canaryIsReadyChannel := make(chan bool)

	go func() {
		configuration := sarama.NewConfig()
		brokers := []string{c.kafkaBrokerAddress}

		for ;;{
			consumer, err := sarama.NewConsumer(brokers, configuration)

			// because we should firstly wait for ready broker this clause should not happened at all
			if err != nil {
				log.Fatal(err.Error())
			}
			topics, err := consumer.Topics()
			if err != nil {
				log.Fatal(err.Error())
			}
			if IsTopicPresent(c.TopicTestName, topics){break}
			log.Println("waiting for topic creation")
			time.Sleep(time.Millisecond * 500)
			err = consumer.Close()
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		canaryIsReadyChannel <- true
	}()
	select {
	case <-timeout:
		log.Fatal("Canary didn't boot up properly")
	case <-canaryIsReadyChannel:
		log.Println("Canary is ready")
	}


}

func (c *Controller) setUpCanaryParamsViaEnv(){
	log.Println("Setting up environment variables")
	os.Setenv(config.ReconcileIntervalEnvVar, c.RetentionTime)
	os.Setenv(config.TopicEnvVar, c.TopicTestName)
	os.Setenv(config.BootstrapServersEnvVar, c.kafkaBrokerAddress)

}