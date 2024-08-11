package lanky_types

import (
	"time"

	"github.com/sirupsen/logrus"
)

type LankyPostgreConf struct {
	Host                   string
	Port                   string
	User                   string
	Password               string
	DbName                 string
	SslMode                string
	TimeZone               string
	EnableDebug            bool
	MaximumIdleConnection  int
	MaximumOpenConnection  int
	ConnectionMaxLifeTime  time.Duration
	SkipDefaultTransaction bool
	SlowSqlThreshold       time.Duration
	Logger                 *logrus.Logger
}
