package bus

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"analytics-svc/internal/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitConsumer struct {
	Service services.TrackService

	URL         string
	Exchange    string
	Queue       string
	Bindings    []string
	PrefetchCnt int
}

type BusEvent struct {
	EventId    string         `json:"event_id"`
	AppName    string         `json:"app_name"`
	UserID     *string        `json:"user_id"`
	EventName  string         `json:"event_name"`
	Properties map[string]any `json:"properties"`
	TS         *time.Time     `json:"ts"`
}

func NewRabbitConsumer(s services.TrackService) *RabbitConsumer {
	url := os.Getenv("RABBITMQ_URL")
	return &RabbitConsumer{
		Service:     s,
		URL:         url,
		Exchange:    "events",
		Queue:       "analytics.events",
		Bindings:    []string{"note.*", "task.*", "user.*"},
		PrefetchCnt: 500,
	}
}

func (c *RabbitConsumer) ConsumeForever(ctx context.Context) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := c.consumeOnce(ctx); err != nil {
			log.Printf("rabbit consume error: %v (retry in %s)", err, backoff)
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = time.Second
	}
}

func (c *RabbitConsumer) consumeOnce(ctx context.Context) error {
	conn, err := amqp.Dial(c.URL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(c.Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	q, err := ch.QueueDeclare(c.Queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	for _, key := range c.Bindings {
		if err := ch.QueueBind(q.Name, key, c.Exchange, false, nil); err != nil {
			return err
		}
	}

	if err := ch.Qos(c.PrefetchCnt, 0, false); err != nil {
		return err
	}

	deliveries, err := ch.Consume(q.Name, "analytics-consumer", false, false, false, false, nil)
	if err != nil {
		return err
	}

	connClosed := make(chan *amqp.Error, 1)
	ch.NotifyClose(connClosed)

	for {
		select {
		case <-ctx.Done():
			return nil

		case cerr := <-connClosed:
			if cerr == nil {
				return nil
			}
			return cerr

		case msg, ok := <-deliveries:
			if !ok {
				return nil
			}

			var e BusEvent
			if err := json.Unmarshal(msg.Body, &e); err != nil {
				_ = msg.Nack(false, false)
				continue
			}

			at := time.Now().UTC()
			if e.TS != nil {
				at = e.TS.UTC()
			}

			_, err := c.Service.Track(services.TrackInput{
				EventId:    e.EventId,
				AppName:    e.AppName,
				UserID:     e.UserID,
				EventName:  e.EventName,
				Properties: e.Properties,
				At:         at,
			})

			if errors.Is(err, services.ErrQueueFull) {
				_ = msg.Nack(false, true)
				continue
			}

			if err != nil {
				_ = msg.Nack(false, true)
				continue
			}

			_ = msg.Ack(false)
		}
	}
}
