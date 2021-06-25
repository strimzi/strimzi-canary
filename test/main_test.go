package test

import (
	"github.com/strimzi/strimzi-canary/test/container_manager"
	"testing"
	"time"
	"os"
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
	controller.StartContainer()
	// staring canary as go routine
	controller.StartCanary()

	// exercise, verify: running all tests
	time.Sleep(time.Second * 5)
	code := m.Run()

	// teardown
	// shutting down of test container is implicit
	controller.StopDefaultZookeeperKafkaNetwork()

	// returning exit code of testing
	os.Exit(code)
}

