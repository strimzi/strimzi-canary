package service_manager

import (
	"github.com/strimzi/strimzi-canary/internal/config"
	"log"
	"os"
	"os/exec"
	"time"
)


type ServiceManager interface {
	StartKafkaZookeeperContainers()
	StopKafkaZookeeperContainers()
	StartCanary()
}

type controller struct {
}

const (
	pathToDockerComposeImage        = "/compose-kafka-zookeeper.yaml"
	pathToMainMethod                = "../cmd/main.go"
)

func (c *controller) StartKafkaZookeeperContainers() {
	log.Println("starting Zookeeper and Kafka")
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fullComposePath := dir + pathToDockerComposeImage
	var cmd = exec.Command("docker-compose", "-f", fullComposePath, "up", "-d" )
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	// some extra time since container starts before we start to communicate with it.
	time.Sleep(time.Second * 30)
	log.Println("docker compose network created")
}

func (c *controller) StopKafkaZookeeperContainers() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fullComposePath := dir + pathToDockerComposeImage
	var cmd = exec.Command("docker-compose", "-f",  fullComposePath, "down")

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
}

func CreateManager() ServiceManager {
	var cManipulation ServiceManager = &controller{}
	return cManipulation
}

func (c *controller) StartCanary() {
	log.Println("starting canary")
	c.setUpCanaryParamsViaEnv()
	var cmd = exec.Command("go", "run",  pathToMainMethod )
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 7)

}

func (c *controller) setUpCanaryParamsViaEnv(){
	os.Setenv(config.ReconcileIntervalEnvVar, "1000")
	os.Setenv(config.BootstrapBackoffScaleEnvVar, "50")
}