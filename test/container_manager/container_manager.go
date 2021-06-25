package container_manager

type ContainerManager interface {
	StartContainer()
	StartDefaultZookeeperKafkaNetwork()
	StopDefaultZookeeperKafkaNetwork()
	GetZookeeperHostPort() string
	GetKafkaHostPort() string
	PrintContainerInfo()
	CopyFile(string, string) bool
	ExecuteScript( string)
	StartCanary()
	Execute(s []string ) (int,error)
}