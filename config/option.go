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
		conf.DBIP = ip
		conf.DBPort = port
		return nil
	}
}
