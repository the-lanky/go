package lanky_rabbitmq

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	lcp "github.com/the-lanky/go/cryptography"
	llg "github.com/the-lanky/go/log"
	llt "github.com/the-lanky/go/types"
)

// Retries represents the number of retries for a specific operation.
type Retries uint

// NewRetries creates a new Retries object with the specified number of retries.
// The retries parameter determines the number of times a task will be retried
// in case of failure. It should be a non-negative integer.
//
// Example usage:
//
//	retries := NewRetries(3)
//	// Perform some task with retries
//	for i := 0; i < retries; i++ {
//	    // Retry logic here
//	}
func NewRetries(retries int) Retries {
	return Retries(retries)
}

// Consumer represents an interface for consuming messages from RabbitMQ.
type Consumer interface {
	Consume(msg amqp091.Delivery) error
}

// LankyConsumer represents a consumer for RabbitMQ.
type LankyConsumer struct {
	Consumer Consumer
}

// LankyPublisherOption represents the options for configuring a LankyPublisher.
type LankyPublisherOption struct {
	Retries      Retries       // The number of retries for publishing a message.
	DelayRetries time.Duration // The delay between retries for publishing a message.
}

// LankyRMQ is an interface that represents a RabbitMQ client for publishing and consuming messages.
type LankyRMQ interface {
	// Publish publishes a message to the specified topic.
	// It takes a context, topic string, message byte slice, and an optional LankyPublisherOption.
	Publish(ctx context.Context, topic string, message []byte, option *LankyPublisherOption)

	// Listen starts listening for messages on the specified consumers.
	// It takes a map of consumer names to LankyConsumer instances.
	Listen(consumers map[string]LankyConsumer)

	// Close closes the connection to the RabbitMQ server.
	Close()
}

type lrmq struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	config     llt.LankyRabbitConf
	log        *logrus.Logger
	crp        lcp.LankyCrypto
}

// Publish publishes a message to a RabbitMQ topic.
//
// Parameters:
//   - ctx: The context.Context for the operation.
//   - topic: The topic to publish the message to.
//   - message: The message to be published.
//   - option: The optional LankyPublisherOption for configuring retries and delays.
//
// Description:
//
//	This function publishes a message to a RabbitMQ topic. It takes a context.Context, a topic string, a message byte slice, and an optional LankyPublisherOption as parameters. The LankyPublisherOption can be used to configure the number of retries and the delay between retries. If the LankyPublisherOption is not provided, default values will be used.
//
//	The function uses a loop to attempt publishing the message multiple times until it succeeds or reaches the maximum number of retries. Each attempt is logged with the try number and a unique identifier. If message encryption fails, the function logs an error and waits for the specified delay before retrying. If publishing to the RabbitMQ channel fails, the function logs an error and waits for the specified delay before retrying. If the message is successfully published, the function logs a success message.
//
//	Note: This function assumes that the RabbitMQ channel and configuration have been properly set up before calling this function.
func (c *lrmq) Publish(
	ctx context.Context,
	topic string,
	message []byte,
	option *LankyPublisherOption,
) {
	var (
		retries = NewRetries(1)
		delay   = time.Second * 1

		try = NewRetries(1)
		uid = uuid.New().String()

		mu      sync.Mutex
		success bool
	)

	if option != nil {
		if rtr := option.Retries; rtr > 0 {
			retries = rtr
		}
		if dl := option.DelayRetries; dl > 0 {
			delay = dl
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for ok := true; ok; ok = try <= retries && !success {
		mu.Lock()

		c.log.Infof("🔼 [%d] [%s] Publish topic %s", try, uid, topic)

		if c.config.EnableDebugMessage {
			c.log.Debugf("🚀 Body: %s", string(message))
		}

		body, err := c.crp.EncryptToBytes(message)
		if err != nil {
			c.log.Infof("❌ [%d] [%s] Failed publish topic %s. Error message encryption!", try, uid, topic)
			c.log.Error(err)
			try++
			time.Sleep(delay)
			mu.Unlock()
			continue
		}

		if err := c.channel.PublishWithContext(
			ctx,
			c.config.ExchangeName,
			topic,
			false,
			false,
			amqp091.Publishing{
				ContentType: "text/plain",
				MessageId:   uid,
				Body:        body,
			},
		); err != nil {
			c.log.Infof("❌ [%d] [%s] Failed publish topic %s", try, uid, topic)
			c.log.Error(err)
			try++
			time.Sleep(delay)
		} else {
			success = true
			c.log.Infof("✅ [%d] [%s] Success publish topic %s", try, uid, topic)
		}

		mu.Unlock()
	}
}

// Listen starts consuming messages from RabbitMQ for the specified consumers.
// It declares the exchange and queue, binds the queue to the specified topics,
// and starts consuming messages from the queue. It invokes the Consume method
// of the consumer for each consumed message. If a panic occurs during message
// consumption, it logs the error, waits for the specified rejoin delay, and
// then restarts the consumer.
//
// Parameters:
//   - consumers: A map of topics and corresponding LankyConsumer instances.
//
// Example:
//
//	c := &lrmq{}
//	consumers := map[string]LankyConsumer{
//	    "topic1": LankyConsumer{Consumer: MyConsumer{}},
//	    "topic2": LankyConsumer{Consumer: MyOtherConsumer{}},
//	}
//	c.Listen(consumers)
//
// Note:
//
//	The LankyConsumer interface should have a Consume method that accepts a
//	*amqp.Delivery parameter and returns an error.
func (c *lrmq) Listen(consumers map[string]LankyConsumer) {
	var mu sync.Mutex

	if err := c.channel.ExchangeDeclare(
		c.config.ExchangeName,
		c.config.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		c.log.Fatalf(
			"❌ [E: %s] [Q: %s] Consumer failed to declare an exchange: %+v",
			c.config.ExchangeName,
			c.config.ExchangeQueue,
			err,
		)
	}

	q, err := c.channel.QueueDeclare(
		c.config.ExchangeQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.log.Fatalf(
			"❌ [E: %s] [Q: %s] Consumer failed to declare a queue: %+v",
			c.config.ExchangeName,
			c.config.ExchangeQueue,
			err,
		)
	}

	for topic := range consumers {
		if err = c.channel.QueueBind(
			q.Name,
			topic,
			c.config.ExchangeName,
			false,
			nil,
		); err != nil {
			c.log.Errorf(
				"❌ [E: %s] [Q: %s] Consumer failed to listening topic %s",
				c.config.ExchangeName,
				c.config.ExchangeQueue,
				topic,
			)
			c.log.Error(err)
		} else {
			c.log.Infof(
				"✨ [E: %s] [Q: %s] Consumer listening to topic: %s",
				c.config.ExchangeName,
				c.config.ExchangeQueue,
				topic,
			)
		}
	}

	messages, err := c.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.log.Fatalf(
			"❌ [E: %s] [Q: %s] Consumer failed to consume message: %+v",
			c.config.ExchangeName,
			c.config.ExchangeQueue,
			err,
		)
	}

	consumerFn := func() {
		var (
			topic     string
			messageId string
			delay     = time.Second * 5
		)

		if c.config.RejoinDelay > 0 {
			delay = c.config.RejoinDelay
		}

		defer func(topic *string, id *string) {
			if r := recover(); r != nil {
				c.log.Errorf(
					"❌ [%s] [%s] Got panic!!!",
					*id,
					*topic,
				)
				c.log.Info("🛠️ Rejoin rabbitmq service...")
				time.Sleep(delay)
				c.Listen(consumers)
			}
		}(&topic, &messageId)

		for msg := range messages {
			mu.Lock()
			topic = msg.RoutingKey
			messageId = msg.MessageId

			c.log.Infof(
				"🔽 [E: %s] [Q: %s] [%s] Consume topic %s",
				c.config.ExchangeName,
				c.config.ExchangeQueue,
				messageId,
				topic,
			)

			if _, ok := consumers[topic]; !ok {
				c.log.Errorf(`❌ [%s] Not found consumer`, topic)
				continue
			}

			decrypted, err := c.crp.DecryptFromBytes(msg.Body)
			if err != nil {
				c.log.Errorf(`❌ [%s] Failed to decrypt message`, topic)
				continue
			}

			if c.config.EnableDebugMessage {
				c.log.Debug(string(decrypted))
			}

			msg.Body = decrypted

			err = consumers[topic].Consumer.Consume(msg)
			if err != nil {
				c.log.Infof("❌ [%s] Failed...", topic)
				c.log.Error(err)
				continue
			}

			c.log.Infof("✅ [%s] [%s] Success...", messageId, topic)

			mu.Unlock()
		}
	}

	go consumerFn()

	c.log.Infof(
		"✅ [E: %s] [Q: %s] Rabbit consumer started...",
		c.config.ExchangeName,
		q.Name,
	)
}

// Close closes the RabbitMQ channel and connection.
// It first attempts to close the channel and logs the result.
// If the channel closing fails, it logs an error message and exits.
// If the channel closing succeeds, it logs a success message.
// Then, it attempts to close the connection and logs the result.
// If the connection closing fails, it logs an error message and exits.
// If the connection closing succeeds, it logs a success message.
func (c *lrmq) Close() {

	if err := c.channel.Close(); err != nil {
		c.log.Info("❌ Failed close channel rabbitmq...")
		c.log.Fatal(err)
	} else {
		c.log.Info("✅ Channel successfully closed")
	}

	if err := c.connection.Close(); err != nil {
		c.log.Info("❌ Failed close connection rabbitmq...")
		c.log.Fatal(err)
	} else {
		c.log.Info("✅ Connection successfully closed")
	}
}

// NewLankyRMQ creates a new instance of LankyRMQ with the given configuration and logger.
// If the logger is nil, a new instance of logrus.Logger will be created with the service name set to "Lanky RabbitMQ".
// It validates the configuration parameters and logs fatal errors if any of the required parameters are empty or invalid.
// It establishes a connection to RabbitMQ using the provided DSN and creates a channel.
// It also initializes a LankyCrypto instance with the provided secret key.
// Returns a pointer to the created LankyRMQ instance.
func NewLankyRMQ(
	conf llt.LankyRabbitConf,
	log *logrus.Logger,
) LankyRMQ {
	if log == nil {
		log = llg.NewInstance(llg.SetServiceName("Lanky RabbitMQ"))
	}

	if len(strings.TrimSpace(conf.Dsn)) == 0 {
		log.Fatal("Dsn should not be empty!")
	}

	if len(strings.TrimSpace(conf.Secret)) == 0 {
		log.Fatal("Secret key should not be empty")
	}

	if len(strings.TrimSpace(conf.Secret)) != 24 {
		log.Fatal("Secret key should be 24 character long")
	}

	if len(strings.TrimSpace(conf.ExchangeName)) == 0 {
		log.Fatal("Exchange name should not be empty")
	}

	if len(strings.TrimSpace(conf.ExchangeQueue)) == 0 {
		log.Fatal("Exchange queue should not be empty")
	}

	if len(strings.TrimSpace(conf.ExchangeType)) == 0 {
		log.Fatal("Exchange type should not be empty")
	}

	con, er := amqp091.Dial(conf.Dsn)
	if er != nil {
		log.Fatalf("❌ Failed to connect rabbitmq: %+v", er)
	}

	chn, er := con.Channel()
	if er != nil {
		log.Fatalf("❌ Failed to create channel rabbitmq: %+v", er)
	}

	crp := lcp.NewLankyCrypto(conf.Secret)

	return &lrmq{
		connection: con,
		channel:    chn,
		config:     conf,
		log:        log,
		crp:        crp,
	}
}
