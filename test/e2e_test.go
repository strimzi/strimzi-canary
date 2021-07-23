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
	httpUrlPrefix   = "http://localhost:8080"
	metricsEndpoint = "/metrics"
)

/* test checks for following:
*  the presence of canary topic,
*  liveliness of topic (messages being produced),
*/
func TestCanaryTopicLiveliness(t *testing.T) {
	log.Println("TestCanaryTopic test starts")

	// setting up timeout and handler for consumer group, there is no need to wait much longer then is the retention time.
	timeout := time.After(time.Second * 10)
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
		brokers := []string{serviceManager.KafkaBrokerAddress}
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
		if !IsTopicPresent(serviceManager.TopicTestName, topics) {
			t.Fatalf("%s is not present amongst existing topic %v", serviceManager.TopicTestName, topics)
		}
		// consume single message
		group, err := sarama.NewConsumerGroup([]string{serviceManager.KafkaBrokerAddress}, "faq-g9", config)
		if err != nil {
			panic(err)
		}

		// set up client for getting partition count on canary topic
		client, _ := sarama.NewClient([]string{serviceManager.KafkaBrokerAddress},config )
		listOfPartitions, _ := client.Partitions(serviceManager.TopicTestName)
		var partitionsCount =  len(listOfPartitions)
		handler.partitionsConsumptionSlice = make([]bool, partitionsCount)

		// set up consumer group's handler for Strimzi canary topic
		topicsToConsume := []string{serviceManager.TopicTestName}

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
	log.Println("TestEndpointsAvailability test startss")

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
		t.Errorf("Prometheus metrics are not updated correctly on endpoint  %s", metricsEndpoint)
	}

}

// Test verifies correctness of canary's metric (produced records)
func TestMetricServerCanaryContent(t *testing.T) {
	log.Println("TestMetricServerCanaryContent test starts")
	// first record is created only after retention time, before that there is no record in metrics
	waitTimeMilliseconds, _ := strconv.Atoi(serviceManager.RetentionTime)
	time.Sleep( time.Duration(waitTimeMilliseconds * 2) * time.Millisecond )

	resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ := ioutil.ReadAll(resp.Body)
	totalProducedRecordsCount, err := strconv.Atoi(parseCanaryRecordsProducedFromMetrics(string(body)))
	if err != nil {
		t.Fatalf("Content of metric server is not updated as expected")
	}

	// for update of this data we have to wait with another request for at least reconcile time.
	time.Sleep( time.Duration(waitTimeMilliseconds * 3) * time.Millisecond )

	resp2, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body2, _ := ioutil.ReadAll(resp2.Body)

	// totalProducedRecordsCount2 stores value produced after defined number of seconds from obtaining totalProducedRecordsCount
	totalProducedRecordsCount2, _ :=  strconv.Atoi(parseCanaryRecordsProducedFromMetrics(string(body2)))
	log.Println("records produced before first request: ", totalProducedRecordsCount)
	log.Println("records produced before second request: ", totalProducedRecordsCount2)
	if totalProducedRecordsCount2 <= totalProducedRecordsCount {

		t.Errorf("Data are not updated within requested time period on endpoint %s", metricsEndpoint)
	}

}
