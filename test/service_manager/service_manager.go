//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

package service_manager

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
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
	pathToCanaryMain                string
}

const (
	canaryTestTopicName         = "__strimzi_canary_test_topicv123"
	kafkaBrokerAddress          = "127.0.0.1:9092"
	canaryReconcileIntervalTime = "1000"

	pathToDockerComposeImage = "compose-kafka-zookeeper.yaml"
	pathToMainMethod         = "../cmd/main.go"
)

func (c *ServiceManager) StartKafkaZookeeperContainers() {
	log.Println("Starting kafka & Zookeeper")

	errComposingContainers := c.executeCmdWithLogging(
		"start kafka and Zookeeper containers using docker-compose",
		"docker-compose",
		"-f", c.pathDockerComposeKafkaZookeeper, "up", "-d",
		)

	if  errComposingContainers != nil {
		log.Fatal(errComposingContainers.Error())
	}
	log.Println("Zookeeper and Kafka containers created")
	// after creation of containers we still have to wait for some time before successful communication with Kafka & Zookeper
	c.waitForBroker()
}

func (c *ServiceManager) StopKafkaZookeeperContainers() {
	log.Println("Stopping kafka & Zookeeper")
	errStoppingContainers := c.executeCmdWithLogging(
		"stop kafka and Zookeeper containers using docker-compose",
		"docker-compose",
		"-f", c.pathDockerComposeKafkaZookeeper, "down",
	)
	if errStoppingContainers != nil {
		log.Fatal(errStoppingContainers.Error())
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

func (c *ServiceManager) StartCanary() {
	log.Println("Starting Canary")
	c.setUpCanaryParamsViaEnv()
	myCmd := exec.Command("go", "run", c.pathToCanaryMain)

	if err := myCmd.Start(); err != nil {
		log.Fatal(err.Error())
	}
}

// per se it means waiting for container's broker to communicate correctly
func (c *ServiceManager) waitForBroker() {
	log.Println("start waiting for broker")
	timeout := time.After(30 * time.Second)
	brokerIsReadyChannel := make(chan bool)

	go func() {
		configuration := sarama.NewConfig()
		brokers := []string{c.KafkaBrokerAddress}

		for {
			// if we can create Cluster Admin, broker can communicate
			if _, err := sarama.NewClusterAdmin(brokers, configuration); err != nil {
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
		log.Println("Broker isn't ready within expected timeout")
		errObtainingLogs := c.executeCmdWithLogging(
			"obtain logs from zookeeper and kafka containers",
			"docker-compose",
			"-f", pathToDockerComposeImage, "logs",
			)
		if errObtainingLogs != nil {
			log.Println("Problem obtaining logs from kafka and zookeeper containers")
			log.Fatal(errObtainingLogs.Error())
		}
		log.Fatal("containers are not in suitable state")
	case <-brokerIsReadyChannel:
		log.Println("Container (Broker) is ready")
	}
}

func (c *ServiceManager) executeCmdWithLogging( commandDescription ,commandName string, commandArgs ...string) error{

	cmd := exec.Command(commandName, commandArgs...)
	var stdout, stderr bytes.Buffer
	// redirect Stdout and Stderr of command to buffers
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// execute command
	err := cmd.Run()
	// log Stdout Stderr from command (stored within buffer)
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	log.Printf("cmd description: %s\n", commandDescription)
	log.Printf("execute cmd: %s %s\n", commandName, strings.Join( commandArgs ," "))
	log.Printf("cmd stdout:\n%s", outStr)
	log.Printf("cmd stderr:\n%s", errStr)
	return err
}

func (c *ServiceManager) setUpCanaryParamsViaEnv() {
	log.Println("Setting up environment variables")
	os.Setenv(config.ReconcileIntervalEnvVar, c.ReconcileIntervalTime)
	os.Setenv(config.TopicEnvVar, c.TopicTestName)
	os.Setenv(config.BootstrapServersEnvVar, c.KafkaBrokerAddress)
}
