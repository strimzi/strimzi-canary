package service_manager

func IsTopicPresent(topicName string, topics []string  ) bool{
	topicsAsMap := map[string]struct{}{}
	for _, topic := range topics {
		topicsAsMap[topic] = struct{}{}
	}
	_, isTopicPresent := topicsAsMap[topicName]
	return isTopicPresent
}