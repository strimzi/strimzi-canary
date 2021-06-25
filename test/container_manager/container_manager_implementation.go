package container_manager

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
)

type controller struct {
	ctx       context.Context
	container testcontainers.Container
	kafkaContainer *testcontainers.LocalDockerCompose
	isDefaultNetworkRunning bool
	host string
}

// constants for evaluating occupying of default ports for Kafka and Zookeeper
func getTakenPorts() []string {
	return []string{"2181", "9092"}
}

func (c* controller) Execute( s []string ) (int, error){
	return c.container.Exec(c.ctx, s)
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

func (c *controller) setContext() {
	c.ctx = context.Background()
}

func (c *controller) GetKafkaHostPort() string {
	hostPortKafka, _ := c.container.MappedPort(c.ctx, "9092")
	return parsePortNumber(hostPortKafka)
}

func (c *controller) GetZookeeperHostPort() string {
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
func (c *controller) StartDefaultZookeeperKafkaNetwork() {
	var err error
	checkPortAvailability(c.host,  []string{"9092", "2181"})

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := dir + "/compose-kafka-zookeeper.yaml"
	c.kafkaContainer = testcontainers.NewLocalDockerCompose([]string{path}, "someName")
	c.kafkaContainer.WithCommand([]string{"up", "-d"}).Invoke()



	//workingDirPath, err := os.Getwd()
	//fmt.Printf("you working dir: %v\n", workingDirPath)
	// Passing of scripts sleep and stop
	// waiting before container is running
	//time.Sleep(6 * time.Second)
	log.Println("docker compose network created")
}


func (c *controller) StopDefaultZookeeperKafkaNetwork() {
	port, err := checkPortAvailability(c.host, getTakenPorts())
	if err != nil {
		log.Println("port already closed: ", port)
	}
	c.kafkaContainer.Down()
}


func (c *controller) StartContainer() {
	c.setContext()
	fmt.Println("1.) som before req")
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
	fmt.Println("2.) som after req")
	var err error
	c.container, err = testcontainers.GenericContainer(c.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	fmt.Println("3.) som after genericContainer")
	if err != nil {
		fmt.Errorf("error")
	}
	fmt.Println("container created")

	workingDirPath, err := os.Getwd()
	fmt.Printf("you working dir: %v\n", workingDirPath)
	// Passing of scripts sleep and stop
}
func CreateManager() ContainerManager {
	var cManipulation ContainerManager = &controller{}
	return cManipulation

}

func (c *controller) StartCanary() {
	log.Println("inside function(): StartCanary1")
	kafkaAddress := "localhost:" + c.GetKafkaHostPort()
	go createCanary(kafkaAddress)
}

