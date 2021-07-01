package container_manager

import (
	"flag"
	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strimzi/strimzi-canary/internal/config"
	"github.com/strimzi/strimzi-canary/internal/security"
	"github.com/strimzi/strimzi-canary/internal/servers"
	"github.com/strimzi/strimzi-canary/internal/services"
	"github.com/strimzi/strimzi-canary/internal/workers"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	clientCreationFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "client_creation_error_total",
		Namespace: "strimzi_canary",
		Help:      "Total number of errors while creating Sarama client",
	}, nil)
)


func createCanary() {
	canaryConfig := config.NewCanaryConfig()

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Errorf("Error on setting logtostderr to true")
	}
	flag.Set("v", strconv.Itoa(canaryConfig.VerbosityLogLevel))
	flag.Parse()

	glog.Infof("Starting Strimzi canary tool with config: %+v", canaryConfig)

	httpServer := servers.NewHttpServer()
	httpServer.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	client, err := newClient(canaryConfig)
	if err != nil {
		glog.Fatalf("Error creating new Sarama client: %v", err)
	}

	topicService := services.NewTopicService(canaryConfig, client)
	producerService := services.NewProducerService(canaryConfig, client)
	consumerService := services.NewConsumerService(canaryConfig, client)

	canaryManager := workers.NewCanaryManager(canaryConfig, topicService, producerService, consumerService)
	canaryManager.Start()

	sig := <-signals
	glog.Infof("Got signal: %v", sig)
	canaryManager.Stop()
	httpServer.Stop()

	glog.Infof("Strimzi canary stopped")
}

func newClient(canaryConfig *config.CanaryConfig) (sarama.Client, error) {
	config := sarama.NewConfig()
	kafkaVersion, err := sarama.ParseKafkaVersion(canaryConfig.KafkaVersion)
	if err != nil {
		return nil, err
	}
	config.Version = kafkaVersion
	config.ClientID = canaryConfig.ClientID
	// set manual partitioner in order to specify the destination partition on sending
	config.Producer.Partitioner = sarama.NewManualPartitioner
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	if canaryConfig.SaramaLogEnabled {
		sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	}

	if canaryConfig.TLSEnabled {
		config.Net.TLS.Enable = true
		if config.Net.TLS.Config, err = security.NewTLSConfig(canaryConfig); err != nil {
			glog.Fatalf("Error configuring TLS: %v", err)
		}
	}

	backoff := services.NewBackoff(canaryConfig.BootstrapBackoffMaxAttempts, canaryConfig.BootstrapBackoffScale*time.Millisecond, services.MaxDefault)
	for {
		client, clientErr := sarama.NewClient([]string{canaryConfig.BootstrapServers}, config)
		if clientErr == nil {
			return client, nil
		}
		delay, backoffErr := backoff.Delay()
		if backoffErr != nil {
			glog.Errorf("Error connecting to the Kafka cluster after %d retries: %v", canaryConfig.BootstrapBackoffMaxAttempts, backoffErr)
			return nil, backoffErr
		}
		clientCreationFailed.With(nil).Inc()
		glog.Warningf("Error creating new Sarama client, retrying in %d ms: %v", delay.Milliseconds(), clientErr)
		time.Sleep(delay)
	}
}

