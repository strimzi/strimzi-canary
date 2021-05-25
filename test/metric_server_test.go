package test

import (
	"github.com/strimzi/strimzi-canary/test/container_manager"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
	"time"
)

func TestMetricServerAvailability(t *testing.T) {
	var controller = container_manager.CreateManager()
	controller.StartContainer()
	controller.StartCanary()

	time.Sleep(8 * time.Second)
	resp, err := http.Get("http://127.0.0.1:8080/metrics")
	if err != nil {
		t.Errorf("Canary unable to create it's metric server: " )
	}

	want := 200
	got := resp.StatusCode
	if got != want {
		t.Errorf("http://localhost:8080 runs but doesn't response OK , for get http://localhost:8081/metrics")
	}
}

func TestMetricServerContentUpdating(t *testing.T) {
	var controller = container_manager.CreateManager()
	controller.StartContainer()
	controller.StartCanary()

	time.Sleep(8 * time.Second)

	resp, _ := http.Get("http://localhost:8080/metrics")
	body, _ := ioutil.ReadAll(resp.Body)
	sb := string(body)
	// we test whether number of result is increased  (got1 stores first produced value)
	got1 := correctDataRegex(sb)
	if len(got1) < 1 {
		t.Errorf("No correct data produced")
	}
	// test  has to wait for 30 second before next round of data producing is finished.

	time.Sleep(time.Second * 31)
	resp2, _ := http.Get("http://localhost:8080/metrics")
	body2, _ := ioutil.ReadAll(resp2.Body)

	sb2 := string(body2)

	// got2 stores value produced after 30 second from recording first value
	got2 := correctDataRegex(sb2)

	want1, want2 := "0" , "1"
	if want1 != got1 || want2 != got2{
		t.Errorf("Data are not updated correctly on service")
	}


}

// We are only interested in counter that is produced
func correctDataRegex( input string) string  {
	regex, _ := regexp.Compile("(?m)^promhttp_metric_handler_requests_total.*(\\d+)$")
	data := regex.FindStringSubmatch(string(input))
	if len(data) > 1 {
		return data[1]
	}
	return ""

}