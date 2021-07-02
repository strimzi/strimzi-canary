package test

import (
	"github.com/segmentio/kafka-go"
	"regexp"
)

// We are only interested in counter that is produced
func parseCountFromMetrics( input string) string  {
	regex, _ := regexp.Compile("(?m)^promhttp_metric_handler_requests_total.*(\\d+)$")
	data := regex.FindStringSubmatch(input)
	if len(data) > 1 {
		return data[1]
	}
	return ""
}

func isTopicPresent(topicName string) (bool, error){
	// get map of topics from metadata
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		return false, err
	}

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false, err
	}

	m := map[string]struct{}{}

	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	_, isPresent := m[topicName]
	return isPresent, nil
}

