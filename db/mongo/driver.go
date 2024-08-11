package lanky_mongo

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	llg "github.com/the-lanky/go/log"
	llt "github.com/the-lanky/go/types"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// LankyMongo represents an interface for interacting with a MongoDB database.
type LankyMongo interface {
	// Database returns the MongoDB database instance.
	Database() *mongo.Database

	// Client returns the MongoDB client instance.
	Client() *mongo.Client

	// Close closes the connection to the MongoDB server.
	Close()
}

// libPrefix is the prefix used for MongoDB related constants in the library.
const libPrefix = "MONGODB"

// fatal is a helper function that logs a fatal error message and exits the program.
// It takes a logrus.Logger instance, a message string, and an error as input.
// If the error is nil, it logs the message with a fatal level and exits the program.
// If the error is not nil, it logs the message with a fatal level, logs the error, and exits the program.
func fatal(log *logrus.Logger, message string, err error) {
	if err == nil {
		log.Fatalf("❌ [%s] %s", libPrefix, message)
	} else {
		log.Infof("❌ [%s] %s", libPrefix, message)
		log.Fatal(err)
	}
}

// success logs a success message with the provided logger.
// It formats the message with a checkmark symbol, the library prefix, and the given message.
// The formatted message is then logged at the Info level.
//
// Parameters:
//   - log: the logger to use for logging the success message
//   - message: the success message to log
//
// Example:
//
//	success(logger, "Operation completed successfully")
func success(log *logrus.Logger, message string) {
	log.Infof("✅ [%s] %s", libPrefix, message)
}

// databaseValidation validates the configuration for connecting to a MongoDB database.
// It checks if the configuration is provided and if the required fields are present.
// If any validation fails, it logs a fatal error using the provided logger.
//
// Parameters:
//   - conf: A pointer to the LankyMongoConf struct containing the configuration details.
//   - logger: A pointer to the logrus.Logger used for logging fatal errors.
func databaseValidation(conf *llt.LankyMongoConf, logger *logrus.Logger) {
	if conf == nil {
		fatal(logger, "instance required a configuration", nil)
	}

	if conf.Protocol == "" {
		fatal(logger, "Protocol is required", nil)
	}

	if conf.Protocol != "mongodb" && conf.Protocol != "mongodb+srv" {
		fatal(logger, "Protocol should be mongodb or mongodb+srv", nil)
	}

	if conf.User == "" {
		fatal(logger, "User is required", nil)
	}

	if conf.Host == "" {
		conf.Host = "localhost"
	}

	if conf.Port == "" && conf.Protocol == "mongodb" {
		fatal(logger, "Port is required", nil)
	}
}

// buildDsn constructs a MongoDB connection string based on the provided configuration.
// It takes a llt.LankyMongoConf object as input and returns a string representing the connection string.
// The connection string is built using the protocol, user, password, host, port, and database information from the configuration.
// If the protocol is "mongo+srv", the connection string is constructed without specifying the port.
// If the configuration has any additional option parameters, they are appended to the connection string.
// The constructed connection string is then returned.
func buildDsn(conf llt.LankyMongoConf) string {
	dsn := fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s",
		conf.Protocol,
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Database,
	)

	if conf.Protocol == "mongo+srv" {
		dsn = fmt.Sprintf(
			"%s://%s:%s@%s/%s",
			conf.Protocol,
			conf.User,
			conf.Password,
			conf.Host,
			conf.Database,
		)
	}

	if len(conf.OptionParameter) > 0 {
		dsn += conf.OptionParameter
	}

	return dsn
}

// buildReadPreference is a function that builds and returns a modified options.ClientOptions based on the provided readPreference.
// It takes in a pointer to options.ClientOptions and a string representing the readPreference.
// The readPreference can be one of the following values: "primary", "primaryPreferred", "secondary", "secondaryPreferred", or "nearest".
// If the readPreference is not one of the valid options, it defaults to "primary".
// The function sets the readPreference on the options.ClientOptions based on the provided readPreference value.
// It returns the modified options.ClientOptions.
func buildReadPreference(opt *options.ClientOptions, readPreference string) *options.ClientOptions {
	switch readPreference {
	case "primary":
		opt.SetReadPreference(readpref.Primary())
	case "primaryPreferred":
		opt.SetReadPreference(readpref.PrimaryPreferred())
	case "secondary":
		opt.SetReadPreference(readpref.Secondary())
	case "secondaryPreferred":
		opt.SetReadPreference(readpref.SecondaryPreferred())
	case "nearest":
		opt.SetReadPreference(readpref.Nearest())
	default:
		opt.SetReadPreference(readpref.Primary())
	}

	return opt
}

// buildMonitor is a function that creates and configures a command monitor for MongoDB client options.
// It takes in a pointer to a ClientOptions struct and a logger from the logrus package.
// The monitor is responsible for logging information about MongoDB commands.
// It sets the Started, Succeeded, and Failed event handlers to log the relevant information.
// The Started event handler logs the database name, command name, and command string.
// The Succeeded event handler logs the database name, command name, duration, and reply string.
// The Failed event handler logs the database name, command name, duration, and failure message.
// Finally, it sets the monitor on the ClientOptions and returns the modified options.
func buildMonitor(opt *options.ClientOptions, logger *logrus.Logger) *options.ClientOptions {
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, e *event.CommandStartedEvent) {
			logger.Infof(
				"[%s] [%s] %s",
				e.DatabaseName,
				e.CommandName,
				e.Command.String(),
			)
		},
		Succeeded: func(ctx context.Context, e *event.CommandSucceededEvent) {
			logger.Infof(
				"[%s] [%s] [%s] %s",
				e.DatabaseName,
				e.CommandName,
				e.Duration.String(),
				e.Reply.String(),
			)
		},
		Failed: func(ctx context.Context, e *event.CommandFailedEvent) {
			logger.Errorf(
				"[%s] [%s] [%s] %s",
				e.DatabaseName,
				e.CommandName,
				e.Duration.String(),
				e.Failure,
			)
		},
	}
	opt = opt.SetMonitor(monitor)
	return opt
}

// buildSuccessMessage generates a success message indicating a successful connection to a MongoDB database.
// It takes a llt.LankyMongoConf object as input and returns a string.
//
// The success message is constructed based on the provided configuration parameters:
// - conf.Protocol: The protocol used to connect to the MongoDB server.
// - conf.User: The username used for authentication.
// - conf.Host: The host address of the MongoDB server.
// - conf.Port: The port number of the MongoDB server.
// - conf.Database: The name of the database connected to.
//
// If the protocol is "mongodb+srv", the success message will be formatted without including the port number.
// Otherwise, the success message will include the port number.
//
// Example usage:
//
//	conf := llt.LankyMongoConf{
//	    Protocol: "mongodb",
//	    User: "admin",
//	    Host: "localhost",
//	    Port: "27017",
//	    Database: "mydb",
//	}
//	successMsg := buildSuccessMessage(conf)
//	fmt.Println(successMsg)
//
// Output:
//
//	Connected to mongo mongodb://admin@localhost:27017/mydb
func buildSuccessMessage(conf llt.LankyMongoConf) string {
	msg := fmt.Sprintf(
		"Connected to mongo %s://%s@%s:%s/%s",
		conf.Protocol,
		conf.User,
		conf.Host,
		conf.Port,
		conf.Database,
	)

	if conf.Protocol == "mongodb+srv" {
		msg = fmt.Sprintf(
			"Connected to mongo %s://%s@%s/%s",
			conf.Protocol,
			conf.User,
			conf.Host,
			conf.Database,
		)
	}

	return msg
}

type mg struct {
	ctx    context.Context
	client *mongo.Client
	db     *mongo.Database
	log    *logrus.Logger
}

// NewLankyMongo creates a new instance of LankyMongo, which is a MongoDB driver for the Lanky library.
// It takes the following parameters:
// - ctx: The context.Context object for managing the lifecycle of the MongoDB connection.
// - conf: The LankyMongoConf object containing the configuration for the MongoDB connection.
// - logger: A pointer to a logrus.Logger object for logging purposes. If nil, a new instance of logrus.Logger will be created.
//
// It returns an instance of LankyMongo.
//
// Example usage:
//
//	ctx := context.Background()
//	conf := llt.LankyMongoConf{
//	  // set configuration values
//	}
//	logger := logrus.New()
//	mongo := NewLankyMongo(ctx, conf, logger)
func NewLankyMongo(
	ctx context.Context,
	conf llt.LankyMongoConf,
	logger *logrus.Logger,
) LankyMongo {
	if logger == nil {
		logger = llg.NewInstance(llg.SetServiceName("Lanky Mongodb"))
	}

	databaseValidation(&conf, logger)

	dsn := buildDsn(conf)
	opt := options.Client()

	opt = opt.ApplyURI(dsn)
	opt = buildReadPreference(opt, conf.ReadPreferrence)
	opt = opt.SetConnectTimeout(conf.ConnectionTimeout)
	opt = opt.SetMaxConnIdleTime(conf.MaxConnIdleTime)
	opt = opt.SetHeartbeatInterval(conf.HeartbeatInterval)
	opt = opt.SetMaxPoolSize(uint64(conf.MaxPoolSize))
	opt = opt.SetMinPoolSize(uint64(conf.MinPoolSize))

	if conf.EnabledMonitor {
		opt = buildMonitor(opt, logger)
	}

	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		fatal(logger, "Failed to create mongodb connection", err)
	}

	err = client.Ping(ctx, opt.ReadPreference)
	if err != nil {
		fatal(logger, "Failed to ping mongodb server", err)
	}

	success(logger, buildSuccessMessage(conf))

	return &mg{
		ctx:    ctx,
		db:     nil,
		client: client,
		log:    logger,
	}
}

func (c *mg) Database() *mongo.Database {
	return c.db
}

func (c *mg) Client() *mongo.Client {
	return c.client
}

func (c *mg) Close() {
	if err := c.client.Disconnect(c.ctx); err != nil {
		fatal(c.log, "Failed disconnecting mongodb", err)
	} else {
		success(c.log, "Connection successully closed")
	}
}
