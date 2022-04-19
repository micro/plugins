module github.com/go-micro/plugins/v3/broker/segmentio

go 1.16

require (
	github.com/asim/go-micro/v3 v3.7.1
	github.com/go-micro/plugins/v3/broker/kafka v1.0.0
	github.com/go-micro/plugins/v3/codec/segmentio v1.0.0
	github.com/google/uuid v1.2.0
	github.com/segmentio/kafka-go v0.4.16
)

replace (
	github.com/go-micro/plugins/v3/broker/kafka => ../kafka
	github.com/go-micro/plugins/v3/codec/segmentio => ../../codec/segmentio
)
