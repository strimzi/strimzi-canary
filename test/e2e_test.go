//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

//go:build e2e

package test

import (
	"context"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

const (
	httpUrlPrefix   = "http://localhost:8080"
	metricsEndpoint = "/metrics"
)

/*
*  test checks for following:
*  the presence of canary topic,
*  Liveness of topic (messages being produced),
 */
func TestCanaryTopicLiveness(t *testing.T) {
	log.Println("TestCanaryTopic test starts")
	consumingHandler := NewConsumerGroupHandler()
	timeout := time.After(time.Second * 10)
	errs := make(chan error, 1)
	// test itself.
	go func() {
		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
		ctx := context.Background()
		clusterAdmin, err := sarama.NewClusterAdmin([]string{serviceManager.KafkaBrokerAddress}, config)
		if err != nil {
			errs <- err
			return
		}

		var topicPartitionCount int
		// wait for topic creation
		for {
			topicMetadata, err := clusterAdmin.DescribeTopics([]string{serviceManager.TopicTestName})
			if err != nil {
				log.Printf("Problem communicating with kafka broker: %v", err)
				time.Sleep(time.Millisecond * 500)
				continue
			}

			if len(topicMetadata) == 0 || errors.Is(topicMetadata[0].Err, sarama.ErrUnknownTopicOrPartition) {
				// topic haven't been created yet.
				time.Sleep(time.Millisecond * 500)
				continue
			}
			canaryTopicMetadata := topicMetadata[0]
			topicPartitionCount = len(canaryTopicMetadata.Partitions)
			break
		}

		// consume single message
		group, err := sarama.NewConsumerGroup([]string{serviceManager.KafkaBrokerAddress}, "faq-g9", config)
		if err != nil {
			errs <- err
			return
		}

		// set up client for getting partition count on canary topic
		consumingHandler.partitionsConsumptionSlice = make([]bool, topicPartitionCount)

		// set up consumer group's consumingHandler for Strimzi canary topic
		topicsToConsume := []string{serviceManager.TopicTestName}

		// group.Consume is blocking
		go group.Consume(ctx, topicsToConsume, consumingHandler)
	}()

	select {
	case <-timeout:
		t.Fatalf("Test didn't finish in time due to message not being read in time")
	case err := <-errs:
		if err != nil {
			t.Fatal(err)
		}
	case <-consumingHandler.consumingDone:
		log.Println("message received")
	}
	close(errs)
}

func TestEndpointsAvailability(t *testing.T) {
	log.Println("TestEndpointsAvailability test starts")

	var testInputs = [...]struct {
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
			t.Fatalf("Http server unreachable for url: %s", completeUrl)
		}

		wantResponseStatus := testInput.expectedStatusCode
		gotResponseStatus := resp.StatusCode
		if wantResponseStatus != gotResponseStatus {
			t.Fatalf("endpoint: %s expected response code: %d obtained: %d", completeUrl, wantResponseStatus, gotResponseStatus)
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
	totalRequestCountT2, _ := strconv.Atoi(parseSucReqRateFromMetrics(string(body2)))
	if totalRequestCountT2 <= totalRequestCountT1 {
		t.Errorf("Prometheus metrics are not updated correctly on endpoint  %s", metricsEndpoint)
	}

}

// Test verifies correctness of canary's metric (produced records)
func TestMetricServerCanaryContent(t *testing.T) {
	log.Println("TestMetricServerCanaryContent test starts")
	// first record is created only after reconcile time, before that there is no record in metrics
	waitTimeMilliseconds, _ := strconv.Atoi(serviceManager.ReconcileIntervalTime)
	time.Sleep(time.Duration(waitTimeMilliseconds*2) * time.Millisecond)

	resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ := ioutil.ReadAll(resp.Body)
	totalProducedRecordsCount, err := strconv.Atoi(parseCanaryRecordsProducedFromMetrics(string(body)))
	if err != nil {
		t.Fatalf("Content of metric server is not updated as expected")
	}

	// for update of this data we have to wait with another request for at least reconcile time.
	time.Sleep(time.Duration(waitTimeMilliseconds*3) * time.Millisecond)

	resp2, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body2, _ := ioutil.ReadAll(resp2.Body)

	// totalProducedRecordsCount2 stores value produced after defined number of seconds from obtaining totalProducedRecordsCount
	totalProducedRecordsCount2, _ := strconv.Atoi(parseCanaryRecordsProducedFromMetrics(string(body2)))
	log.Println("records produced before first request: ", totalProducedRecordsCount)
	log.Println("records produced before second request: ", totalProducedRecordsCount2)
	if totalProducedRecordsCount2 <= totalProducedRecordsCount {
		t.Errorf("Data are not updated within requested time period on endpoint %s", metricsEndpoint)
	}
}
