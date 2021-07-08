package test

import (
	"context"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)


const (
	httpUrlPrefix                   = "http://localhost:8080"
	metricsEndpoint                 = "/metrics"
	canaryTopicName                 = "__strimzi_canary"
	metricServerUpdateTimeInSeconds = 30
	kafkaServer = "localhost:9092"
)

/* test checks for following:
*  the presence of canary topic,
*  liveliness of topic (messages being produced),
*/
func TestCanaryTopicLiveliness(t *testing.T) {
	log.Println("TestCanaryTopic test starts")

	// setting up timeout
	timeout := time.After(40 * time.Second)
	done := make(chan bool)

	// test itself.
	go func() {
		// test topic presence
		isPresent, err := isTopicPresent(canaryTopicName, kafkaServer)
		if err != nil {
			t.Errorf("cannot connect to canary topic: %s due to error %s\n",canaryTopicName, err.Error()  )
		}
		if ! isPresent {
			t.Errorf("%s topic doesn't exist", canaryTopicName)
		}
		log.Printf("%s topic is present\n", canaryTopicName)

		// test topic liveliness
		partition := 0
		conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", canaryTopicName, partition)
		if err != nil {
			log.Fatal("failed to dial leader:", err)
		}
		// read single message
		log.Println("waiting for message from kafka")
		_, err = conn.ReadMessage(10)
		if err != nil {
			t.Errorf("error when waiting for message from %s: %s", canaryTopicName , err.Error())
		}
		log.Println("waiting for next message")
		done <- true

	}()

	select {
	case <-timeout:
		t.Error("Test didn't finish in time")
	case <-done:
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
	totalRequestCountT1 := parseCountFromMetrics(string(body))
	if len(totalRequestCountT1) < 1 {
		t.Errorf("Content of metric server is not updated as expected")
	}

	// test  has to wait for Defined time before next round of data producing is finished.
	time.Sleep(time.Second * (metricServerUpdateTimeInSeconds + 1))
	resp, _ = http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ = ioutil.ReadAll(resp.Body)

	// totalRequestCountT2 stores value produced after defined number of seconds from obtaining totalRequestCountT1
	totalRequestCountT2 := parseCountFromMetrics(string(body))
	if totalRequestCountT2 <= totalRequestCountT1{
		t.Errorf("Data are not updated within requested time period %d on endpoint %s", metricServerUpdateTimeInSeconds, metricsEndpoint)
	}

}





