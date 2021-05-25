package test

import (
	"github.com/strimzi/strimzi-canary/test/container_manager"
	"testing"
)


// test will start script inside container that will run script which will sleep for 10 seconds
func TestM(t *testing.T) {
	var controller = container_manager.CreateManager()
	controller.StartContainer()

	// canary can be only started once your container is running (it must bind to it's kafka mapped port)
	controller.StartCanary()


	controller.CopyFile("/files_to_transfer/sleep.sh", "/sleep.sh")
	controller.ExecuteScript("/sleep.sh")

	// there is no need to turn down container as it goes down automatically with test
}
