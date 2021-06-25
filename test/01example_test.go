package test



//
//import (
//	"github.com/strimzi/strimzi-canary/test/container_manager"
//	"testing"
//)
//
//
//// test will start script inside container that will run script which will sleep for 10 seconds
//func TestM(t *testing.T) {
//	println("1")
//	var controller = container_manager.CreateManager()
//	controller.StartContainer()
//
//
//	println("2")
//	controller.StartContainer()
//	println("3")
//	// canary can be only started once your container is running (it must bind to it's kafka mapped port)
//	println("salala cakam nez zapnem canary")
//	controller.StartCanary()
//	println("TYPEO")
//
//	controller.CopyFile("/files_to_transfer/sleep.sh", "/sleep.sh")
//	controller.ExecuteScript("/sleep.sh")
//
//	// there is no need to turn down container as it goes down automatically with test
//}
