package container_manager

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
)

type controller struct {
	ctx       context.Context
	container testcontainers.Container
}



// Path is assumed from you working directory
// rights to file which will be there are set to 775(oct) = 509 decimal
// true if passed, false if failed to do so.
func (c *controller) CopyFile(sourceFilePath string, destinationFilePath string) bool{
	workingDirPath, _ := os.Getwd()
	copyFile1 := c.container.CopyFileToContainer(c.ctx, workingDirPath+ sourceFilePath, destinationFilePath, 509)
	//copyFileToContainer returns nil if everything goes well
	if copyFile1 != nil {
		return false
	}
	return true
}
// param is path to script inside container such as "/sleep.sh"
// if it is such function it is good to call it as go function
//TODO return type
func (c *controller) ExecuteScript(scriptPath string) {
	fmt.Println("executing script")
	executeScript, _ := c.container.Exec(c.ctx, []string{scriptPath})
	fmt.Println("output of script execution", executeScript)
}

func (c *controller) SetContext() {
	c.ctx = context.Background()
}

func (c *controller) GetKafkaHostPort() string {
	hostPortKafka, _ := c.container.MappedPort(c.ctx, "9092")
	return parsePortNumber(hostPortKafka)
}

func (c *controller) GetZookeperHostPort() string {
	hostPortZoo, _ := c.container.MappedPort(c.ctx, "2181")
	return parsePortNumber(hostPortZoo)

}

func (c *controller) PrintContainerInfo() {
	containerID := c.container.GetContainerID()
	host, _ := c.container.Host(c.ctx)
	fmt.Printf("Info\n")
	fmt.Printf("containerId: %+v\n", containerID)
	fmt.Printf("host: %+v\n", host)

}

func (c *controller) StartContainer() {
	c.SetContext()
	req := testcontainers.ContainerRequest{
		Image:        "johnnypark/kafka-zookeeper",
		ExposedPorts: []string{"2181/tcp", "9092/tcp"},
		Env: map[string]string{
			"ADVERTISED_HOST": "127.0.0.1",
			// it is important to set this value to 1 otherwise checksum in Canary won't match
			"NUM_PARTITIONS":  "1",
		},
		AutoRemove: true,
		//we must wait some time for container to setup kafka before trying to connect to kafka.
		WaitingFor: wait.ForLog("INFO [KafkaServer id=0] started (kafka.server.KafkaServer)"),
	}
	var err error
	c.container, err = testcontainers.GenericContainer(c.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Errorf("error")
	}
	fmt.Println("container created")

	workingDirPath, err := os.Getwd()
	fmt.Printf("you working dir: %v\n", workingDirPath)
	// Passing of scripts sleep and stop
}
func CreateManager(  ) ContainerManager {
	var cManipulation ContainerManager = &controller{}
	return cManipulation

}

func (c *controller) StartCanary() {
	kafkaAddress := "localhost:" + c.GetKafkaHostPort()
	go createCanary(kafkaAddress)
}

