package rabbitmq

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queue     amqp.Queue
	queueName string
}

func GetServerAck(ServerName string) bool {
	return config.Viper.GetBool(fmt.Sprintf("consumer.%v.autoAck", ServerName))
}

func GetMqUrl() string {
	MqUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/%v",
		config.Viper.GetString("server.username"),
		config.Viper.GetString("server.password"),
		config.Viper.GetString("server.host"),
		config.Viper.GetInt("server.port"),
		config.Viper.GetString("server.vhost"),
	)
	return MqUrl
}

func NewRabbitMQInstance(queueName string) (*RabbitMQ, error) {
	rabbitmq := &RabbitMQ{}
	var err error
	rabbitmq.conn, err = amqp.Dial(GetMqUrl())
	if err != nil {
		return nil, err
	}
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	if err != nil {
		return nil, err
	}
	rabbitmq.queueName = queueName
	return rabbitmq, nil
}

func (r *RabbitMQ) Destroy() error {
	err := r.channel.Close()
	if err != nil {
		return err
	}
	err = r.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) PublishSimple(ctx context.Context, message []byte) error {
	q, err := r.channel.QueueDeclare(
		r.queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = r.channel.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
			Timestamp:   time.Now(),
		})
	if err != nil {
		if r.conn.IsClosed() || r.channel.IsClosed() {
			return errors.Wrap(err, "PublishSimple:Conn or channel is Closed")
		}
		return err
	}
	return nil
}

func (r *RabbitMQ) ConsumeSimple() (<-chan amqp.Delivery, error) {
	q, _ := r.channel.QueueDeclare(
		r.queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	msgs, err := r.channel.Consume(
		q.Name,
		"",
		config.Viper.GetBool("consumer.favorite.autoAck"),
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
