package test

import (
	"github.com/strimzi/strimzi-canary/test/service_manager"
	"log"
	"os"
	"testing"
)

var (
	controller service_manager.ServiceManager
)

func TestMain(m *testing.M) {
	controller = service_manager.CreateManager()
	controller.StartCanary()
	// starting of network for default kafka and zookeeper ports 9092, 2182
	controller.StartKafkaZookeeperContainers()

	// exercise, verify: running all tests
	log.Println("Starting tests")
	code := m.Run()

	// defer has no usage here.
	controller.StopKafkaZookeeperContainers()
	// returning exit code of testing
	os.Exit(code)
}

