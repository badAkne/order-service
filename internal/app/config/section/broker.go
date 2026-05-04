package section

type (
	Broker struct {
		Kafka BrokerKafka
	}

	BrokerKafka struct {
		Addresses     []string `required:"true"`
		ConsumerGroup string   `split_words:"true"`
		ClientID      string   `split_words:"true" default:"order-service"`

		ModelOrder BrokerKafkaModelOrder `split_words:"true"`
	}

	BrokerKafkaModelOrder struct {
		Created BrokerKafkaModelOrderCreated `split_words:"true"`
	}

	BrokerKafkaModelOrderCreated struct {
		Topic         string `required:"true" default:"order.created"`
		ConsumerGroup string `split_words:"true"` // only for sub, has priority
	}
)
