package lanky_types

import (
	"time"

	"github.com/sirupsen/logrus"
)

// LankyPostgreConf represents the configuration options for connecting to a PostgreSQL database.
type LankyPostgreConf struct {
	Host                   string         // The hostname or IP address of the PostgreSQL server.
	Port                   string         // The port number of the PostgreSQL server.
	User                   string         // The username for authenticating with the PostgreSQL server.
	Password               string         // The password for authenticating with the PostgreSQL server.
	DbName                 string         // The name of the PostgreSQL database.
	SslMode                string         // The SSL mode for the PostgreSQL connection.
	TimeZone               string         // The timezone to use for the PostgreSQL connection.
	EnableDebug            bool           // Whether to enable debug mode for the PostgreSQL connection.
	MaximumIdleConnection  int            // The maximum number of idle connections in the connection pool.
	MaximumOpenConnection  int            // The maximum number of open connections in the connection pool.
	ConnectionMaxLifeTime  time.Duration  // The maximum lifetime of a connection in the connection pool.
	SkipDefaultTransaction bool           // Whether to skip the default transaction for each connection.
	SlowSqlThreshold       time.Duration  // The threshold duration for logging slow SQL queries.
	Logger                 *logrus.Logger // The logger instance for logging PostgreSQL-related messages.
}
