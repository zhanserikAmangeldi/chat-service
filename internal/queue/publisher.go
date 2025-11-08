package queue

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zhanserikAmangeldi/chat-service/internal/models"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(rabbitURL string) (*Publisher, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		"chat_exchange",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

func (p *Publisher) PublishMessage(msg *models.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.channel.Publish(
		"chat_exchange",
		"message."+msg.RoomID,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
