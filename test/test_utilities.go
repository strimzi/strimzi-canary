package test

import (
	"regexp"
)


// We are only interested in counter that is produced
func ParseCountFromMetrics( input string) string  {
	regex, _ := regexp.Compile("(?m)^promhttp_metric_handler_requests_total.*200...(\\d+)$")
	data := regex.FindStringSubmatch(input)
	if len(data) > 1 {
		return data[1]
	}
	return ""
}


func IsTopicPresent(topicName string, topics []string  ) bool{
	topicsAsMap := map[string]struct{}{}
	for _, topic := range topics {
		topicsAsMap[topic] = struct{}{}
	}
	_, isTopicPresent := topicsAsMap[topicName]
	return isTopicPresent
}
