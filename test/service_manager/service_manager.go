package service_manager

import (
	"context"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/testcontainers/testcontainers-go"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)


type ServiceManager interface {
	StartKafkaZookeeperContainers()
	StopKafkaZookeeperContainers()
	StartCanary()
	StopCanary()
}


type controller struct {
	ctx       context.Context
	container testcontainers.Container
	kafkaContainer *testcontainers.LocalDockerCompose
	isDefaultNetworkRunning bool
	host string
	cmd *exec.Cmd
	id string
}


const (
	pathToDockerComposeImage        = "/compose-kafka-zookeeper.yaml"
	pathToMainMethod                = "../cmd/main.go"
)


func (c *controller) setContext() {
	c.ctx = context.Background()
}


func (c *controller) StartKafkaZookeeperContainers() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := dir + pathToDockerComposeImage
	c.kafkaContainer = testcontainers.NewLocalDockerCompose([]string{path}, "someName")
	c.kafkaContainer.WithCommand([]string{"up", "-d"}).Invoke()

	log.Println("docker compose network created")
}


func (c *controller) StopKafkaZookeeperContainers() {
	c.kafkaContainer.Down()
}


func CreateManager() ServiceManager {
	var cManipulation ServiceManager = &controller{}
	return cManipulation
}


func (c *controller) StartCanary() {
	log.Println("starting canary")
	c.setUpCanaryParamsViaEnv()
	//out, err :=  exec.Command("go", "run",  pathToMainMethod ).Output()

	c.cmd = exec.Command("go", "run",  pathToMainMethod )

	log.Println(c.cmd)
	if err := c.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println(c.cmd.Process.Pid)
	c.id = strconv.Itoa(c.cmd.Process.Pid)
	// once canary starts it may take extra time to open endpoints
	log.Println(c.cmd.ProcessState)
	log.Println("!!!!!!!!before canary wait ")
	time.Sleep(time.Second * 40)
	log.Println("!!!!!!!!!after canary wait")
}

func (c *controller) StopCanary(){
	log.Println("Stopping Canary")
	log.Println("Stopping Canary")
	log.Println("Stopping Canary")
	log.Println("Stopping Canary")
	log.Println("Stopping Canary")
	log.Println("Stopping Canary")



	log.Println("Stopping Canary")
	//var a os.Signal = os.Kill
	var f  = exec.Command("kill", "-9", c.id    )
	f.Run()
	//
	//if err := c.cmd.Process.Signal(a); err != nil {
	//	log.Fatal("failed to stop Canary: ", err)
	//}
	//if err := c.cmd.Process.Signal(a); err != nil {
	//	log.Fatal("failed to stop Canary: ", err)
	//}
	//if err := c.cmd.Process.Signal(a); err != nil {
	//	log.Fatal("failed to stop Canary: ", err)
	//}
	log.Println("Canary Stopped Successfully")
}

func (c *controller) setUpCanaryParamsViaEnv(){
	os.Setenv(config.ReconcileIntervalEnvVar, "500000")
}