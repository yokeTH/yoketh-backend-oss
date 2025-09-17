package db

type Config struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	DBName   string `env:"NAME,required"`
	SSLMode  string `env:"SSLMODE,required"`
}
