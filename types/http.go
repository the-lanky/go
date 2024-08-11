package lanky_types

import "time"

// LankyServerConf represents the configuration for a Lanky server.
type LankyServerConf struct {
	Host          string        // Host specifies the hostname or IP address on which the server should listen.
	Addr          string        // Addr specifies the network address on which the server should listen.
	ReadTimeout   time.Duration // ReadTimeout specifies the maximum duration for reading the entire request.
	WriteTimeout  time.Duration // WriteTimeout specifies the maximum duration before timing out writes of the response.
	IdleTimeout   time.Duration // IdleTimeout specifies the maximum amount of time to wait for the next request when keep-alives are enabled.
	ShutdownDelay time.Duration // ShutdownDelay specifies the delay before forcefully shutting down the server.
}
