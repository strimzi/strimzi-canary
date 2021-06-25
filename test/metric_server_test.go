package test

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"testing"
	"time"
)

const (
	httpUrlPrefix = "http://localhost:8080"
	metricsEndpoint = "/metrics"
	canaryTopicName = "__strimzi_canary"
	kafkaVersion = "2.6.0"
)

func TestCanaryTopicLiveliness(t *testing.T) {
	fmt.Println("TestCanaryTopicLiveliness")
	topic := canaryTopicName
	partition := 0
	// connect to kafka broker
	//conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:" + controller.GetKafkaHostPort(), topic, partition)
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil { log.Fatal("failed to dial leader:", err) }
	// read single message
	fmt.Println("waiting for message")
	messsage, err := conn.ReadMessage(10)
	if err != nil {
		t.Errorf("some error: %s", err.Error())
	}

	fmt.Println(string(messsage.Value))


}


func TestCanaryTopicPresence(t *testing.T) {
	t.Skip()
	fmt.Println("TestCanaryTopicPresence")
	var command = []string{
		"./opt/kafka_2.13-"+ kafkaVersion +"/bin/kafka-topics.sh",
		"--describe", "--zookeeper",
		"localhost:2181",
		"--topic",
		canaryTopicName,
	}
	var res, err = controller.Execute(command)
	if err != nil  {
		t.Errorf("failed to find topic %s, due to %s", canaryTopicName, err.Error())
	}
	if res != 0 {
		t.Errorf("topic %s doesn't exist", canaryTopicName)
	}
}


func TestEndpointsAvailability(t *testing.T) {
	fmt.Println("TestEndpointsAvailability")
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
		log.Println("endpoint:", input.endpoint   )
		var completeUrl string = httpUrlPrefix + input.endpoint
		resp, err := http.Get(completeUrl)
		if err != nil {
			t.Errorf("Http server unreachable for url: %s",completeUrl  )
		}
		wantResponseStatus := input.responseCode
		gotResponseStatus := resp.StatusCode
		if wantResponseStatus != gotResponseStatus {
			t.Errorf("expected response code: %d obtained: %d",wantResponseStatus,gotResponseStatus  )
		}
	}
}

func TestMetricServerContentUpdating(t *testing.T) {
	fmt.Println("TestMetricServerContentUpdating")
	resp, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body, _ := ioutil.ReadAll(resp.Body)
	sb := string(body)
	// we test whether number of result is increased  (got1 stores first produced value)
	got1 := parseCountFromMetrics(sb)
	if len(got1) < 1 {
		t.Errorf("No correct data produced")
	}
	// test  has to wait for 30 second before next round of data producing is finished.
	time.Sleep(time.Second * 31)
	resp2, _ := http.Get(httpUrlPrefix + metricsEndpoint)
	body2, _ := ioutil.ReadAll(resp2.Body)
	time.Sleep(time.Second * 250)
	sb2 := string(body2)
	// got2 stores value produced after 30 second from recording first value
	got2 := parseCountFromMetrics(sb2)
	if got2 <= got1{
		log.Println(got2)
		log.Println(got1)
		t.Errorf("Data are not updated correctly on service")
	}
	controller.StopDefaultZookeeperKafkaNetwork()
}





// We are only interested in counter that is produced
func parseCountFromMetrics( input string) string  {
	regex, _ := regexp.Compile("(?m)^promhttp_metric_handler_requests_total.*(\\d+)$")
	data := regex.FindStringSubmatch(input)
	if len(data) > 1 {
		return data[1]
	}
	return ""

}