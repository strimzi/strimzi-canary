package container_manager

type ContainerManager interface {
	StartContainer()
	GetZookeperHostPort() string
	GetKafkaHostPort() string
	PrintContainerInfo()
	CopyFile(string, string) bool
	ExecuteScript( string)
	StartCanary()
}