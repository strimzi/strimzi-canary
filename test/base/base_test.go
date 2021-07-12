// +build e2e

package base

import (
	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/test"
	"io/ioutil"
	"log"
	"net/http"
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
	log.Println("1TestCanaryTopic test startss")

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
		cluster, err := sarama.NewConsumer(brokers, config)

		if err != nil {
			t.Error(err.Error())
		}
		// get all topics
		topics, err := cluster.Topics()
		if err != nil {
			t.Error(err.Error())
		}


		if !test.IsTopicPresent(canaryTopicName, topics) {
			t.Errorf("%s is not present", canaryTopicName)
		}

		// consume single message
		consumer, _ := sarama.NewConsumer(brokers, nil)
		partitionConsumer, _ := consumer.ConsumePartition(canaryTopicName, 0, 0)
		msg := <-partitionConsumer.Messages()
		log.Printf("Consumed: offset:%d  value:%v", msg.Offset, string(msg.Value))
		testDone <- true

	}()

	select {
	case <-timeout:
		t.Error("Test didn't finish in time due to message not being read in time")
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
			t.Errorf("Http server unreachable for url: %s",completeUrl  )
		}

		wantResponseStatus := testInput.expectedStatusCode
		gotResponseStatus := resp.StatusCode
		if wantResponseStatus != gotResponseStatus {
			t.Errorf("endpoint: %s expected response code: %d obtained: %d" ,completeUrl,wantResponseStatus,gotResponseStatus  )
		}
		log.Printf("endpoint:  %s, responded with expected status code %d\n", testInput.endpoint, testInput.expectedStatusCode)
	}
}

func TestMetricServerContentUpdating(t *testing.T) {
	log.Println("TestMetricServerContentUpdating test starts")

	resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ := ioutil.ReadAll(resp.Body)
	totalRequestCountT1 := test.ParseCountFromMetrics(string(body))
	if len(totalRequestCountT1) < 1 {
		t.Errorf("Content of metric server is not updated as expected")
	}

	// test wait for period of time before sending next request
	time.Sleep(time.Second * metricEndpointRequestTimeout)
	resp2, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body2, _ := ioutil.ReadAll(resp2.Body)


	// totalRequestCountT2 stores value produced after defined number of seconds from obtaining totalRequestCountT1
	totalRequestCountT2 := test.ParseCountFromMetrics(string(body2))
	if totalRequestCountT2 <= totalRequestCountT1{
		t.Errorf("Data are not updated within requested time period %d on endpoint %s", metricEndpointRequestTimeout, metricsEndpoint)
	}

}





