package rabbitmq

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type RabbitMQ struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	QueueName     string
	Exchange      string
	Key           string
	MqUrl         string
	Queue         amqp.Queue
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	prefetchCount int
}

func GetServerAck(ServerName string) bool {
	return config.Viper.GetBool(fmt.Sprintf("consumer.%v.autoAck", ServerName))
}

func DefaultRabbitMQInstance(ServerName string) (*RabbitMQ, error) {
	MqUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/%v",
		config.Viper.GetString("server.username"),
		config.Viper.GetString("server.password"),
		config.Viper.GetString("server.host"),
		config.Viper.GetInt("server.port"),
		config.Viper.GetString("server.vhost"),
	)
	prefetchCount := config.Viper.GetInt(fmt.Sprintf("consumer.%v.prefetchCount", ServerName))
	autoAck := config.Viper.GetBool(fmt.Sprintf("consumer.%v.autoAck", ServerName))
	return NewRabbitMQInstance(ServerName, "", "", MqUrl, prefetchCount, autoAck)
}

func NewRabbitMQStruct(queueName string, exchange string, key string, MqUrl string, prefetchCount int) *RabbitMQ {
	return &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, MqUrl: MqUrl, prefetchCount: prefetchCount}
}

func NewRabbitMQInstance(queueName, exchange, key, MqUrl string, prefetchCount int, autoAck bool) (*RabbitMQ, error) {
	rabbitmq := NewRabbitMQStruct(queueName, exchange, key, MqUrl, prefetchCount)
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	if err != nil {
		return nil, err
	}
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	if err != nil {
		return nil, err
	}
	if !autoAck {
		err = rabbitmq.channel.Qos(rabbitmq.prefetchCount, 0, false)
		if err != nil {
			return nil, err
		}
	}
	rabbitmq.channel.NotifyClose(rabbitmq.notifyClose)
	rabbitmq.channel.NotifyPublish(rabbitmq.notifyConfirm)
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
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		if r.conn.IsClosed() || r.channel.IsClosed() {
			return errors.Wrap(err, "PublishSimple")
		}
		return err
	}
	err = r.channel.PublishWithContext(
		ctx,
		r.Exchange,
		r.QueueName,
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
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
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
