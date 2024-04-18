package rabbitmq

import (
	"context"
	"github.com/GabiHert/pesquisai-rabbitmq-lib/connection"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	connection                            connection.Connection
	channel                               *amqp.Channel
	queue                                 *amqp.Queue
	name, contentType, exchange, consumer string
	createIfNotExists                     bool
}

func (q *Queue) Connect(user, pwd, host, port, queueName string) (err error) {
	err = q.connection.Connect(user, pwd, host, port)
	if err != nil {
		return
	}

	q.name = queueName
	if q.channel == nil {
		q.channel, err = q.connection.Channel()
		if err != nil {
			return
		}
	}

	if q.queue == nil && q.createIfNotExists {
		var queue amqp.Queue
		queue, err = q.channel.QueueDeclare(
			q.name,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return
		}

		q.queue = &queue
	}
	return
}

func (q *Queue) Publish(ctx context.Context, content []byte) (err error) {
	err = q.channel.PublishWithContext(
		ctx,
		q.exchange,
		q.name,
		false,
		false,
		amqp.Publishing{
			ContentType: q.contentType,
			Body:        content,
		},
	)
	return err
}

func (q *Queue) Close() (err error) {
	err = q.channel.Close()
	if err != nil {
		return
	}

	err = q.connection.Close()
	if err != nil {
		return
	}

	return
}

func (q *Queue) Consume(ctx context.Context, handler func(delivery amqp.Delivery)) (err error) {
	messagesCh, err := q.channel.ConsumeWithContext(
		ctx,
		q.name,
		q.consumer,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	forever := make(chan bool)
	go func() {
		for msg := range messagesCh {
			handler(msg)
		}
	}()

	<-forever
	return
}

func NewQueue(name, contentType string, createIfNotExists bool) *Queue {
	return &Queue{
		channel:           nil,
		queue:             nil,
		name:              name,
		contentType:       contentType,
		createIfNotExists: createIfNotExists,
	}
}