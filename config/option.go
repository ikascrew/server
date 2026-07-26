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

// Ikasbox は ikasbox(HTTP :5555)を同一プロセス内で同居起動する
// モードを有効にする。db は ikasbox.db のパス
func Ikasbox(db string) Option {
	return func(conf *Config) error {
		conf.Ikasbox = true
		conf.IkasboxDB = db
		return nil
	}
}
