// +build e2e

package test

import (
	"context"
	"github.com/Shopify/sarama"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)


const (
	metricEndpointRequestTimeout = 3
	httpUrlPrefix   = "http://localhost:8080"
	metricsEndpoint = "/metrics"
	canaryTopicName = "__strimzi_canary"
	KafkaBroker     = "127.0.0.1:9092"
)

/* test checks for following:
*  the presence of canary topic,
*  liveliness of topic (messages being produced),
*/
func TestCanaryTopicLiveliness(t *testing.T) {
	log.Println("TestCanaryTopic test starts")

	// setting up timeout and handler for consumer group
	timeout := time.After(40 * time.Second)
	handler := exampleConsumerGroupHandler{}
	handler.mutexWritePartitionPresence = &sync.Mutex{}
	handler.consumingDone = make(chan bool)

	// test itself.
	go func() {
		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true
		config.Consumer.Offsets.Initial =  sarama.OffsetOldest
		ctx := context.Background()

		//kafka end point
		brokers := []string{KafkaBroker}
		//get broker
		consumer, err := sarama.NewConsumer(brokers, config)

		if err != nil {
			t.Fatalf(err.Error())
		}
		// get all topics
		topics, err := consumer.Topics()
		if err != nil {
			t.Fatalf(err.Error())
		}

		if !IsTopicPresent(canaryTopicName, topics) {
			t.Fatalf("%s is not present", canaryTopicName)
		}

		// consume single message
		group, err := sarama.NewConsumerGroup([]string{KafkaBroker}, "faq-g9", config)
		if err != nil {
			panic(err)
		}

		// set up client for getting partition count on canary topic
		client, _ := sarama.NewClient([]string{KafkaBroker},config )
		listOfPartitions, _ := client.Partitions(canaryTopicName)
		var partitionsCount =  len(listOfPartitions)
		handler.partitionsConsumptionSlice = make([]bool, partitionsCount)

		// set up consumer group's handler for Strimzi canary topic
		topicsToConsume := []string{canaryTopicName}

		// group.Consume is blocking
		go group.Consume(ctx, topicsToConsume, handler)
	}()

	select {
	case <-timeout:
		t.Fatalf("Test didn't finish in time due to message not being read in time")
	case <-handler.consumingDone:
		log.Println("message received")
	}

}

func TestEndpointsAvailability(t *testing.T) {
	log.Println("TestEndpointsAvailability test starts")

	var testInputs = [...]struct{
		endpoint           string
		expectedStatusCode int
	}{
		{metricsEndpoint, 200},
		{"/liveness", 200},
		{"/readiness", 200},
		{"/invalid", 404},
	}

	for _, testInput := range testInputs {
		var completeUrl = httpUrlPrefix + testInput.endpoint
		resp, err := http.Get(completeUrl)
		if err != nil {
			t.Fatalf("Http server unreachable for url: %s",completeUrl  )
		}

		wantResponseStatus := testInput.expectedStatusCode
		gotResponseStatus := resp.StatusCode
		if wantResponseStatus != gotResponseStatus {
			t.Fatalf("endpoint: %s expected response code: %d obtained: %d" ,completeUrl,wantResponseStatus,gotResponseStatus  )
		}
		log.Printf("endpoint:  %s, responded with expected status code %d\n", testInput.endpoint, testInput.expectedStatusCode)
	}
}

func TestMetricServerPrometheusContent(t *testing.T) {
	log.Println("TestMetricServerPrometheusContent test starts")

	resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ := ioutil.ReadAll(resp.Body)
	totalRequestCountT1, _ := strconv.Atoi(parseSucReqRateFromMetrics(string(body)))
	if totalRequestCountT1 < 1 {
		t.Fatalf("Content of metric server is not updated as expected")
	}

	resp2, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body2, _ := ioutil.ReadAll(resp2.Body)

	// totalRequestCountT2 stores value produced after defined number of seconds from obtaining totalRequestCountT1
	totalRequestCountT2, _ :=  strconv.Atoi(parseSucReqRateFromMetrics(string(body2)))
	if totalRequestCountT2 <= totalRequestCountT1{
		t.Errorf("Data are not updated within requested time period %d on endpoint %s", metricEndpointRequestTimeout, metricsEndpoint)
	}

}

// Test looks at any sort of published error such as client creation describe cluster etc...
func TestMetricServerCanaryContent(t *testing.T) {
	log.Println("TestMetricServerCanaryContent test starts")
	// request is repeated multiple times (after waiting 0.5 second) in case data have not yet been published.
	var totalErrorString string
	var repeat = 5
	for i := 0; i < repeat; i++ {
		resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
		body, _ := ioutil.ReadAll(resp.Body)
		totalErrorString = parseGeneratedErrorsFromMetrics(string(body))
		if totalErrorString == "" {
			time.Sleep(time.Microsecond * 500)
			continue
		}
		// data have been loaded
		break
	}

	totalErrorsCount, err := strconv.Atoi(totalErrorString)
	// Because canary starts before there is Any Broker available, there should be reports about whole sort of unsuccessful attempts.
	if totalErrorsCount == 0 {
		t.Fatalf("Content of metric server is not updated as expected")
	}
	if err != nil {
		t.Fatalf(err.Error())
	}
}
