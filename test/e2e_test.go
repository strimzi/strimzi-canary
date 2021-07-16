// +build e2e

package test

import (
	"github.com/Shopify/sarama"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)


const (
	httpUrlPrefix                      = "http://localhost:8080"
	metricsEndpoint                    = "/metrics"
	canaryTopicName                    = "__strimzi_canary"
	metricEndpointRequestTimeout       = 3
	kafkaMainBroker                    = "127.0.0.1:9092"
)

/* test checks for following:
*  the presence of canary topic,
*  liveliness of topic (messages being produced),
*/
func TestCanaryTopicLiveliness(t *testing.T) {
	log.Println("TestCanaryTopic test starts")

	// setting up timeout
	timeout := time.After(40 * time.Second)
	testDone := make(chan bool)
	// test itself.
	go func() {
		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true

		//kafka end point
		brokers := []string{kafkaMainBroker}
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
		partitionConsumer, _ := consumer.ConsumePartition(canaryTopicName, 0, 0)
		msg := <-partitionConsumer.Messages()
		log.Printf("Consumed: offset:%d  value:%v", msg.Offset, string(msg.Value))
		testDone <- true

	}()

	select {
	case <-timeout:
		t.Fatalf("Test didn't finish in time due to message not being read in time")
	case <-testDone:
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

func TestMetricServerContentUpdating(t *testing.T) {
	log.Println("TestMetricServerContentUpdating test starts")
	time.Sleep(time.Second * 1)

	resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ := ioutil.ReadAll(resp.Body)
	totalRequestCountT1, _ := strconv.Atoi(ParseCountFromMetrics(string(body)))
	if totalRequestCountT1 < 1 {
		t.Fatalf("Content of metric server is not updated as expected")
	}

	// test wait for period of time before sending next request
	time.Sleep(time.Second * metricEndpointRequestTimeout)
	resp2, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body2, _ := ioutil.ReadAll(resp2.Body)


	// totalRequestCountT2 stores value produced after defined number of seconds from obtaining totalRequestCountT1
	totalRequestCountT2, _ :=  strconv.Atoi(ParseCountFromMetrics(string(body2)))
	if totalRequestCountT2 <= totalRequestCountT1{
		t.Fatalf("tcount1:  %d tcount2: %d ", totalRequestCountT1, totalRequestCountT2)
		//t.Errorf("Data are not updated within requested time period %d on endpoint %s", metricEndpointRequestTimeout, metricsEndpoint)

	}

}





