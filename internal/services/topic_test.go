//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// +build unit_test

// Package services defines an interface for canary services and related implementations

package services

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/strimzi/strimzi-canary/internal/config"
	"math/rand"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestRequestedAssignments(t *testing.T) {
	var tests = []struct {
		name                string
		numPartitions       int
		numBrokers          int
		useRack             bool
		brokersWithMultipleLeaders []int32
		expectedMinISR      int
	}{
		{"one broker", 1, 1, false, []int32{}, 1},
		{"three brokers without rack info", 3, 3, false, []int32{}, 2},
		{"fewer brokers than partitions", 3, 2, false, []int32{0}, 1},
		{"six brokers with rack info",  6, 6, true, []int32{}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.CanaryConfig{
				Topic: "test",
			}

			brokers, brokerMap := createBrokers(t, tt.numBrokers, tt.useRack)

			ts := NewTopicService(cfg, nil)

			assignments, minISR := ts.requestedAssignments(tt.numPartitions, brokers)

			if tt.expectedMinISR != minISR {
				t.Errorf("unexpected minISR, got = %d, want = %d", minISR, tt.expectedMinISR)
			}

			for _, brokerIds := range assignments {
				duplicatedBrokers := make(map[int32]int)

				for _, brokerId := range brokerIds {
					if _, ok := duplicatedBrokers[brokerId]; !ok {
						duplicatedBrokers[brokerId] = 1
					} else {
						duplicatedBrokers[brokerId] = duplicatedBrokers[brokerId] + 1
					}

				}

				for brokerId, count := range duplicatedBrokers {
					if count > 1 {
						t.Errorf("partition is replicated on same broker (%d) more than once (%d)", brokerId, count)
					}
				}
			}

			leaderBrokers := make(map[int32]int)
			for _, brokerIds := range assignments {

				leaderBrokerId := brokerIds[0]
				if _, ok := leaderBrokers[leaderBrokerId]; !ok {
					leaderBrokers[leaderBrokerId] = 1
				} else {
					leaderBrokers[leaderBrokerId] = leaderBrokers[leaderBrokerId] + 1
				}
			}

			for brokerId, count := range leaderBrokers {
				if count > 1 {
					found := false
					for _, expectedBrokerId  := range tt.brokersWithMultipleLeaders {
						if expectedBrokerId == brokerId {
							found = true
							break
						}
					}

					if !found {
						t.Errorf("broker %d is leader for more than one partition (%d)", brokerId, count)
					}
				}
			}

			if tt.useRack {
				for i, brokerIds := range assignments {
					rackBrokerId := make(map[string][]int32)
					for _, brokerId := range brokerIds {
						broker := brokerMap[brokerId]
						_, ok := rackBrokerId[broker.Rack()]
						if !ok {
							rackBrokerId[broker.Rack()] = make([]int32, 0)
						}
						rackBrokerId[broker.Rack()] = append(rackBrokerId[broker.Rack()], broker.ID())
					}

					for rack, brokerIds := range rackBrokerId {
						if len(brokerIds) > 1 {
							t.Errorf("partition %d has been assigned to %d brokers %v in rackBrokerId %s", i, len(brokerIds), brokerIds, rack)
						}
					}
				}
			}

			ts.Close()
		})
	}

}

func createBrokers(t *testing.T, num int, rack bool) ([]*sarama.Broker, map[int32]*sarama.Broker) {
	brokers := make([]*sarama.Broker, 0)
	brokerMap := make(map[int32]*sarama.Broker)
	for i := 0; i < num ; i++ {
		broker := &sarama.Broker{}

		setBrokerID(t, broker, i)

		brokers = append(brokers, broker)
		brokerMap[broker.ID()] = broker
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(brokers), func(i, j int) { brokers[i], brokers[j] = brokers[j], brokers[i] })

	if rack {
		rackNames := make([]string, 3)
		for i, _ := range rackNames {
			rackNames[i] = fmt.Sprintf("useRack%d", i)
		}

		for i, broker := range brokers {
			rackName := rackNames[i%3]
			setBrokerRack(t, broker, rackName)
		}
	}

	return brokers, brokerMap
}

func setBrokerID(t *testing.T, broker *sarama.Broker, i int) {
	idV := reflect.ValueOf(broker).Elem().FieldByName("id")
	setUnexportedField(idV, int32(i))
	if int32(i) != broker.ID() {
		t.Errorf("failed to set id by reflection")
	}
}

func setBrokerRack(t *testing.T, broker *sarama.Broker, rackName string) {
	rackV := reflect.ValueOf(broker).Elem().FieldByName("rack")
	setUnexportedField(rackV, &rackName)
	if rackName != broker.Rack() {
		t.Errorf("failed to set useRack by reflection")
	}
}

func setUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}
