//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package services defines some canary related services
package services

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
)

// TopicReconcileResult contains the result of a topic reconcile
type TopicReconcileResult struct {
	// new current number of brokers
	BrokersNumber int
	// new partitions assignments across brokers
	Assignments map[int32][]int32
	// if a refresh metadata is needed
	RefreshMetadata bool
}

// TopicService defines the service for canary topic management
type TopicService struct {
	canaryConfig *config.CanaryConfig
	client       sarama.Client
	admin        sarama.ClusterAdmin
	initialized  bool
}

// ErrExpectedClusterSize defines the error raised when the expected cluster size is not met
type ErrExpectedClusterSize struct{}

func (e *ErrExpectedClusterSize) Error() string {
	return fmt.Sprintf("Current cluster size differs from the expected size")
}

// NewTopicService returns an instance of TopicService
func NewTopicService(canaryConfig *config.CanaryConfig, client sarama.Client) *TopicService {
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		log.Printf("Error creating the Sarama cluster admin: %v", err)
		panic(err)
	}
	ts := TopicService{
		canaryConfig: canaryConfig,
		client:       client,
		admin:        admin,
	}
	return &ts
}

// Reconcile does a reconcile on the canary topic
//
// It first checks the number of brokers and gets the topic metadata
// If topic doesn't exist it's created with a partitions assignments having
// one leader partition for each broker
//
// It topic already exists, it checks the number of partitions compared to the current brokers
// 1. if cluster scaled up, it adds partitions processing a reassignment
// 2. if cluster scaled down, it just does a reassignment.
// In case of cluster scaled down, the partitions above the number of brokers are considered orphans
// and the producer will not send messages to him
//
// If a scale up, scale down, scale up happens, it forces a leader election for having preferred leaders
func (ts *TopicService) Reconcile() (TopicReconcileResult, error) {
	result := TopicReconcileResult{0, nil, false}
	// getting brokers for assigning canary topic replicas accordingly
	// on creation or cluster scale up/down when topic already exists
	brokers, _, err := ts.admin.DescribeCluster()
	if err != nil {
		log.Printf("Error describing cluster: %v", err)
		return result, err
	}
	// the brokers number that reflects the expected partitions for sending messages can be dynamic or fixed/expected
	// depending if the "dynamic" reassignment is enabled or not
	if ts.isDynamicReassignmentEnabled() {
		result.BrokersNumber = len(brokers)
	} else {
		result.BrokersNumber = ts.canaryConfig.ExpectedClusterSize
	}

	metadata, err := ts.admin.DescribeTopics([]string{ts.canaryConfig.Topic})
	if err != nil {
		log.Printf("Error retrieving metadata for topic %s: %v", ts.canaryConfig.Topic, err)
		return result, err
	}
	topicMetadata := metadata[0]

	if topicMetadata.Err == sarama.ErrUnknownTopicOrPartition {

		// canary topic doesn't exist, going to create it
		log.Printf("The canary topic %s doesn't exist\n", topicMetadata.Name)
		// topic is created if "dynamic" reassignment is enabled or the expected brokers are provided by the describe cluster
		if ts.isDynamicReassignmentEnabled() || ts.canaryConfig.ExpectedClusterSize == len(brokers) {

			if result.Assignments, err = ts.createTopic(len(brokers)); err != nil {
				log.Printf("Error creating topic %s: %v", topicMetadata.Name, err)
				return result, err
			}
			log.Printf("The canary topic %s was created\n", topicMetadata.Name)
		} else {
			log.Printf("The canary topic wasn't created. Expected brokers %d, Actual brokers %d",
				ts.canaryConfig.ExpectedClusterSize, len(brokers))
			// not creating the topic and returning error to avoid starting producer/consumer
			return result, &ErrExpectedClusterSize{}
		}

	} else {
		// canary topic already exists, check replicas assignments
		log.Printf("The canary topic %s already exists\n", topicMetadata.Name)
		logTopicMetadata(topicMetadata)

		// topic partitions reassignment happens if "dynamic" reassignment is enabled
		// or the topic service is just starting up
		if ts.isDynamicReassignmentEnabled() || !ts.initialized {

			log.Printf("Going to alter topic and reassigning partitions if needed")
			result.RefreshMetadata = len(brokers) != len(topicMetadata.Partitions)
			if result.Assignments, err = ts.alterTopic(len(topicMetadata.Partitions), len(brokers)); err != nil {
				log.Printf("Error altering topic %s: %v", topicMetadata.Name, err)
				return result, err
			}
			ts.checkTopic(len(brokers), topicMetadata)
			// TODO force a leader election. The feature is missing in Sarama library right now.
		}

	}
	ts.initialized = true
	return result, err
}

// Close closes the underneath Sarama admin instance
func (ts *TopicService) Close() {
	log.Printf("Closing topic service")

	err := ts.admin.Close()
	if err != nil {
		log.Printf("Error closing the Sarama cluster admin: %v", err)
		os.Exit(1)
	}
	log.Printf("Topic service closed")
}

func (ts *TopicService) createTopic(brokersNumber int) (map[int32][]int32, error) {
	assignments, minISR := ts.assignments(0, brokersNumber)

	v := strconv.Itoa(int(minISR))
	topicConfig := map[string]*string{
		"min.insync.replicas": &v,
	}

	topicDetail := sarama.TopicDetail{
		NumPartitions:     -1,
		ReplicationFactor: -1,
		ReplicaAssignment: assignments,
		ConfigEntries:     topicConfig,
	}
	err := ts.admin.CreateTopic(ts.canaryConfig.Topic, &topicDetail, false)
	return assignments, err
}

func (ts *TopicService) alterTopic(currentPartitions int, brokersNumber int) (map[int32][]int32, error) {
	assignmentsMap, _ := ts.assignments(currentPartitions, brokersNumber)

	assignments := make([][]int32, len(assignmentsMap))
	for i := 0; i < len(assignments); i++ {
		assignments[i] = make([]int32, len(assignmentsMap[int32(i)]))
		copy(assignments[i], assignmentsMap[int32(i)])
	}
	log.Printf("%v", assignments)

	var err error
	// less partitions than brokers (scale up)
	if currentPartitions < brokersNumber {
		// when replication factor is less than 3 because brokers are not 3 yet (see replicationFactor := min(brokersNumber, 3)),
		// it's not possible to create the new partitions directly with a replication factor higher than the current ones.
		// So first alter the assignment of current partitions with new replicas (higher replication factor)
		err = ts.alterAssignments(assignments[:currentPartitions])
		if err == nil {
			// passing the assigments just for the partitions that needs to be created
			err = ts.admin.CreatePartitions(ts.canaryConfig.Topic, int32(brokersNumber), assignments[currentPartitions:], false)
		}
	} else {
		// more or equals partitions than brokers, just need reassignment
		err = ts.alterAssignments(assignments[:currentPartitions])
	}
	return assignmentsMap, err
}

func (ts *TopicService) checkTopic(brokersNumber int, metadata *sarama.TopicMetadata) {
	electLeader := false
	if len(metadata.Partitions) == brokersNumber {
		for _, p := range metadata.Partitions {
			if p.ID != p.Leader {
				electLeader = true
			}
		}
	}
	log.Printf("Elect leader = %t\n", electLeader)
}

func (ts *TopicService) assignments(currentPartitions int, brokersNumber int) (map[int32][]int32, int) {
	partitions := max(currentPartitions, brokersNumber)
	replicationFactor := min(brokersNumber, 3)
	minISR := max(1, replicationFactor-1)

	assignments := make(map[int32][]int32, int(partitions))
	for p := 0; p < int(partitions); p++ {
		assignments[int32(p)] = make([]int32, int(replicationFactor))
		k := p
		for r := 0; r < int(replicationFactor); r++ {
			assignments[int32(p)][r] = int32(k % int(brokersNumber))
			k++
		}
	}
	log.Printf("assignments = %v, minISR = %d", assignments, int(minISR))
	return assignments, int(minISR)
}

// Alter the replica assignment for the partitions
//
// After the request for the replica assignement, it run a loop for checking if the reassignment is still ongoing
// It returns when the reassignment is done or there is an error
func (ts *TopicService) alterAssignments(assignments [][]int32) error {
	err := ts.admin.AlterPartitionReassignments(ts.canaryConfig.Topic, assignments)
	if err != nil {
		return err
	}

	partitions := make([]int32, 0, len(assignments))
	for partitionID := range assignments {
		partitions = append(partitions, int32(partitionID))
	}
	// loop for checking that there is no ongoing reassignments
	for {
		ongoing := false
		reassignments, err := ts.admin.ListPartitionReassignments(ts.canaryConfig.Topic, partitions)
		if err != nil {
			return nil
		}
		// on each partition of the topic shouldn't be adding or removing replicas ongoing
		for _, reassignmentStatus := range reassignments[ts.canaryConfig.Topic] {
			log.Printf("List reassignments = %+v\n", reassignmentStatus)
			ongoing = ongoing || (len(reassignmentStatus.AddingReplicas) != 0 || len(reassignmentStatus.RemovingReplicas) != 0)
		}
		if !ongoing {
			break
		}
		time.Sleep(2000 * time.Millisecond)
	}
	return nil
}

// If the "dynamic" topic partitions reassignment is enabled
func (ts *TopicService) isDynamicReassignmentEnabled() bool {
	return ts.canaryConfig.ExpectedClusterSize == config.ExpectedClusterSizeDefault
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func logTopicMetadata(topicMetadata *sarama.TopicMetadata) {
	log.Printf("Metadata for %s topic\n", topicMetadata.Name)
	for _, p := range topicMetadata.Partitions {
		log.Printf("\t{ID:%d Leader:%d Replicas:%v Isr:%v OfflineReplicas:%v}\n", p.ID, p.Leader, p.Replicas, p.Isr, p.OfflineReplicas)
	}
}
