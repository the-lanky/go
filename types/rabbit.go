package lanky_types

import "time"

// LankyRabbitConf represents the configuration for RabbitMQ.
type LankyRabbitConf struct {
	Dsn                string        // The RabbitMQ DSN.
	ExchangeName       string        // The name of the exchange.
	ExchangeType       string        // The type of the exchange.
	ExchangeQueue      string        // The name of the exchange queue.
	Secret             string        // Secret represents the secret value used for authentication or encryption. Should be 24 character long
	EnableDebugMessage bool          // EnableDebugMessage indicates whether debug messages should be enabled.
	RejoinDelay        time.Duration // RejoinDelay represents the duration to wait before attempting to rejoin a connection.
}
