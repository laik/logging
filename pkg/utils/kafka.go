package utils

import (
	"github.com/segmentio/kafka-go"
	"net"
	"strconv"
)

func CheckTopicExist(uri, topic string) (bool, error) {
	conn, err := kafka.Dial("tcp", uri)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false, err
	}

	for _, partition := range partitions {
		if partition.Topic == topic {
			return true, nil
		}
	}

	return false, nil
}

func CreateTopic(uri, topic string, partition int) error {
	conn, err := kafka.Dial("tcp", uri)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     partition,
			ReplicationFactor: 1,
		},
	}
	return controllerConn.CreateTopics(topicConfigs...)
}
