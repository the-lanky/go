package lanky_postgre

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	llog "github.com/the-lanky/go/log"
	llt "github.com/the-lanky/go/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

// LankyPostgreDb is an interface that represents a connection to a PostgreSQL database.
type LankyPostgreDb interface {
	// Db returns the underlying *gorm.DB instance.
	Db() *gorm.DB

	// Sql returns the underlying *sql.DB instance.
	Sql() *sql.DB

	// Close closes the database connection.
	Close()
}

// postgre represents a PostgreSQL database connection.
type postgre struct {
	db    *gorm.DB       // The GORM database connection.
	sqlDb *sql.DB        // The SQL database connection.
	log   *logrus.Logger // The logger instance for logging.
}

// NewLankyPostgre creates a new instance of LankyPostgreDb with the given configuration.
// It establishes a connection to the PostgreSQL database using the provided configuration parameters.
// If the logger parameter is nil, a default logger instance will be created.
// The isProduction parameter determines the log level for the connection.
// The function returns a pointer to the LankyPostgreDb interface.
func NewLankyPostgre(conf llt.LankyPostgreConf, isProduction bool, logger *logrus.Logger) LankyPostgreDb {
	if logger == nil {
		logger = llog.NewInstance(
			llog.SetServiceName("Lanky PostgreDB"),
		)
	}

	logger.Info("🆕 Creating database connection...")

	logLevel := glog.Info
	if isProduction {
		logLevel = glog.Warn
	}

	gormLogger := glog.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		glog.Config{
			SlowThreshold:             conf.SlowSqlThreshold,
			Colorful:                  true,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logLevel,
		},
	)

	tmpDsn := make([]string, 0)

	if conf.Host == "" {
		conf.Host = "localhost"
	}

	if conf.User == "" {
		conf.User = "postgres"
	}

	if conf.Port == "" {
		conf.Port = "5432"
	}

	if conf.DbName == "" {
		conf.DbName = "postgres"
	}

	tmpDsn = append(tmpDsn, fmt.Sprintf("host=%s", conf.Host))
	tmpDsn = append(tmpDsn, fmt.Sprintf("user=%s", conf.User))
	tmpDsn = append(tmpDsn, fmt.Sprintf("port=%s", conf.Port))
	tmpDsn = append(tmpDsn, fmt.Sprintf("dbname=%s", conf.DbName))

	if conf.Password != "" {
		tmpDsn = append(tmpDsn, fmt.Sprintf("password=%s", conf.Password))
	}

	if conf.SslMode != "" {
		tmpDsn = append(tmpDsn, fmt.Sprintf("sslmode=%s", conf.SslMode))
	}

	if conf.TimeZone != "" {
		tmpDsn = append(tmpDsn, fmt.Sprintf("TimeZone=%s", conf.TimeZone))
	}

	dsn := strings.Join(tmpDsn, " ")

	db, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN: dsn,
		}),
		&gorm.Config{
			Logger:                 gormLogger,
			SkipDefaultTransaction: conf.SkipDefaultTransaction,
		},
	)
	if err != nil {
		logger.Info("❌ Failed connecting to the database")
		logger.Fatal(err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		logger.Info("❌ Failed get the database")
		logger.Fatal(err)
	}

	err = sqlDb.Ping()
	if err != nil {
		logger.Info("❌ Connection lost...")
		logger.Fatal(err)
	}

	var (
		maxIdleConnection = 5
		maxOpenConnection = 10
		connMaxLifeTime   = time.Hour
	)

	if conf.MaximumIdleConnection > 0 {
		maxIdleConnection = conf.MaximumIdleConnection
	}

	if conf.MaximumOpenConnection > 0 {
		maxOpenConnection = conf.MaximumOpenConnection
	}

	if conf.ConnectionMaxLifeTime > 0 {
		connMaxLifeTime = conf.ConnectionMaxLifeTime
	}

	sqlDb.SetMaxIdleConns(maxIdleConnection)
	sqlDb.SetMaxOpenConns(maxOpenConnection)
	sqlDb.SetConnMaxLifetime(connMaxLifeTime)

	logger.Infof(
		"✅ Successfully connect to the database %s@%s:%s/%s",
		conf.User,
		conf.Host,
		conf.Port,
		conf.DbName,
	)

	return &postgre{
		db:    db,
		sqlDb: sqlDb,
		log:   logger,
	}
}

func (p *postgre) Db() *gorm.DB {
	return p.db
}

func (p *postgre) Sql() *sql.DB {
	return p.sqlDb
}

func (p *postgre) Close() {
	if err := p.Sql().Close(); err != nil {
		p.log.Info("❌ Failed to close connection database!")
		p.log.Fatal(err)
	} else {
		p.log.Info("✅ Success closing database connection...")
	}
}
