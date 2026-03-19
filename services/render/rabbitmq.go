package main

type RabbitMQ struct {
}

func connectRabbitMQ(url string) (*RabbitMQ, error) {
	return &RabbitMQ{}, nil
}

func (r *RabbitMQ) consume(handler func([]byte) error) error {
	return nil
}

func (r *RabbitMQ) close() {
}
