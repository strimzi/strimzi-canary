package container_manager

type ContainerManager interface {
	StartDefaultZookeeperKafkaNetwork()
	StopDefaultZookeeperKafkaNetwork()
	StartCanary()
}