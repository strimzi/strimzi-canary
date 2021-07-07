package test

import (
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"context"
	"log"
	"net/http"
	"testing"
	"time"
)

/* TODO: current version doesn't use dynamic port allocation due to Canary not running properly on Kafka mapped to any other port than default  9092
+ what is later propagated as series of unwanted behaviours, for example not having open port on 9092 ends in failed attempt to connect to this port despite other port being set in configuration
*/


const (
	httpUrlPrefix                   = "http://localhost:8080"
	metricsEndpoint                 = "/metrics"
	canaryTopicName                 = "__strimzi_canary"
	metricServerUpdateTimeInSeconds = 30
)

/* test checks for following:
*  the presence of canary topic,
*  liveliness of topic (messages being produced),
*/
func TestCanaryTopicLiveliness(t *testing.T) {

	log.Println("TestCanaryTopic test starts")

	// setting Up timeout
	timeout := time.After(40 * time.Second)
	done := make(chan bool)

	// test itself.
	go func() {
		// test topic presence
		isPresent, err := isTopicPresent(canaryTopicName)
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
		log.Println("Canary produces messages")
		// check expected format of topic
		done <- true

	}()

	select {
	case <-timeout:
		t.Error("Test didn't finish in time")
	case <-done:
		log.Println("message successfully received")
	}

}

func TestEndpointsAvailability(t *testing.T) {
	log.Println("TestEndpointsAvailability test starts")
	var inputs = [...]struct{
		endpoint string
		responseCode int
	}{
		{metricsEndpoint, 200},
		{"/liveness", 200},
		{"/readiness", 200},
		{"/invalid", 404},
	}

	for _, input := range inputs {

		var completeUrl string = httpUrlPrefix + input.endpoint
		resp, err := http.Get(completeUrl)
		if err != nil {
			t.Errorf("Http server unreachable for url: %s",completeUrl  )
		}
		wantResponseStatus := input.responseCode
		gotResponseStatus := resp.StatusCode
		if wantResponseStatus != gotResponseStatus {
			t.Errorf("endpoint: %s expected response code: %d obtained: %d" ,completeUrl,wantResponseStatus,gotResponseStatus  )
		}
		log.Printf("endpoint:  %s, responded with expected status code %d\n", input.endpoint, input.responseCode)
	}
}

func TestMetricServerContentUpdating(t *testing.T) {
	log.Println("TestMetricServerContentUpdating test starts")
	resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ := ioutil.ReadAll(resp.Body)
	sb := string(body)
	// we test whether number of result is increased  (got1 stores first produced value)
	got1 := parseCountFromMetrics(sb)
	if len(got1) < 1 {
		t.Errorf("No correct data produced")
	}
	// test  has to wait for 30 second before next round of data producing is finished.
	time.Sleep(time.Second * (metricServerUpdateTimeInSeconds + 1))
	resp2, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body2, _ := ioutil.ReadAll(resp2.Body)
	sb2 := string(body2)
	// got2 stores value produced after 30 second from recording first value
	got2 := parseCountFromMetrics(sb2)
	if got2 <= got1{
		log.Println(got2)
		log.Println(got1)
		t.Errorf("Data are not updated within requested time period %d on endpoint %s", metricServerUpdateTimeInSeconds, metricsEndpoint)
	}

}





