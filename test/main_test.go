package test

import (
	"github.com/strimzi/strimzi-canary/test/container_manager"
	"log"
	"os"
	"testing"
)

// global variables for accessing container information
var (
	controller container_manager.ContainerManager
)

func TestMain(m *testing.M) {
	// setup
	// setting up manager
	controller = container_manager.CreateManager()
	// starting of network for default kafka and zookeeper ports 9092, 2182
	controller.StartDefaultZookeeperKafkaNetwork()
	// starting kafka zookeeper on testContainer with Dynamic ports
	// staring canary as go routine
	log.Println("startujem canary")
	controller.StartCanary()
	// exercise, verify: running all tests
	log.Println("startujem testy")
	code := m.Run()

	// teardown
	// shutting down of test container is implicit
	controller.StopDefaultZookeeperKafkaNetwork()

	// returning exit code of testing
	os.Exit(code)
}

