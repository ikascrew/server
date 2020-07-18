package config

type Option func(*Config) error

func Port(p int) Option {
	return func(conf *Config) error {
		conf.Port = p
		return nil
	}
}

func Database(ip string, port int) Option {
	return func(conf *Config) error {
		conf.BoxIP = ip
		conf.BoxPort = port
		return nil
	}
}
