package container_manager

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/servers"
	"github.com/strimzi/strimzi-canary/internal/services"
	"github.com/strimzi/strimzi-canary/internal/workers"
)

func createCanary( containerKafkaPort string ) {
	// get canary configuration
	canaryConfig := config.NewCanaryConfig()
	canaryConfig.BootstrapServers = containerKafkaPort

	log.Printf("Starting Strimzi canary tool for testing with config: %+v\n", canaryConfig)
	//metricsServer := servers.NewMetricsServer()
	//metricsServer.Start()

	metricsServer := servers.NewHttpServer()
	metricsServer.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	client := newClient(canaryConfig)

	topicService := services.NewTopicService(canaryConfig, client)
	producerService := services.NewProducerService(canaryConfig, client)
	consumerService := services.NewConsumerService(canaryConfig, client)

	canaryManager := workers.NewCanaryManager(canaryConfig, topicService, producerService, consumerService)
	canaryManager.Start()

	select {
	case sig := <-signals:
		log.Printf("Got signal: %v\n", sig)
	}
	canaryManager.Stop()
	metricsServer.Stop()

	log.Printf("Strimzi canary stopped")
}

func newClient(canaryConfig *config.CanaryConfig) sarama.Client {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.ClientID = canaryConfig.ClientID
	// set manual partitioner in order to specify the destination partition on sending
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	// TODO: handling error
	client, _ := sarama.NewClient([]string{canaryConfig.BootstrapServers}, config)
	return client
}

