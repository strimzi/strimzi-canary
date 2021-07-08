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
	controller.StartKafkaZookeeperContainers()

	// teardown
	// shutting down of test container is implicit
	defer controller.StopKafkaZookeeperContainers()

	// starting canary as go routine.
	controller.StartCanary()
	// exercise, verify: running all tests
	log.Println("Starting tests")
	code := m.Run()

	// returning exit code of testing
	os.Exit(code)
}

