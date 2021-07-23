package test

import (
	"github.com/strimzi/strimzi-canary/test/service_manager"
	"log"
	"os"
	"path"
	"runtime"
	"testing"
)

var (
	serviceManager *service_manager.ServiceManager
)

// returning into the root from /test
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {

	serviceManager = service_manager.CreateManager()
	// starting of network for default kafka and zookeeper ports 9092, 2182
	serviceManager.StartKafkaZookeeperContainers()
	serviceManager.StartCanary()

	log.Println("Starting tests")
	code := m.Run()

	// defer has no usage here.
	serviceManager.StopKafkaZookeeperContainers()
	// returning exit code of testing
	os.Exit(code)
}

