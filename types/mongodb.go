package lanky_types

import "time"

// LankyMongoConf represents the configuration options for connecting to a MongoDB database.
type LankyMongoConf struct {
	Protocol          string        // The protocol to use for the connection (e.g., "mongodb").
	Host              string        // The hostname or IP address of the MongoDB server.
	User              string        // The username for authentication.
	Password          string        // The password for authentication.
	Database          string        // The name of the database to connect to.
	Port              string        // The port number for the MongoDB server.
	OptionParameter   string        // Additional options for the connection.
	ReadPreferrence   string        // The read preference for the connection.
	ConnectionTimeout time.Duration // The timeout for establishing a connection.
	MaxConnIdleTime   time.Duration // The maximum time a connection can remain idle.
	HeartbeatInterval time.Duration // The interval for sending heartbeat messages.
	MaxPoolSize       uint          // The maximum number of connections in the connection pool.
	MinPoolSize       uint          // The minimum number of connections in the connection pool.
	EnabledMonitor    bool          // Whether to enable monitoring of the connection.
}
