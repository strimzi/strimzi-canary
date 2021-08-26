//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

package test

import (
	"log"
	"os"
	"testing"

	"github.com/strimzi/strimzi-canary/test/service_manager"
)

var (
	serviceManager *service_manager.ServiceManager
)

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
