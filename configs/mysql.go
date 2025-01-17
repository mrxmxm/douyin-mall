package configs

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Host:     "localhost",
		Port:     3308,
		User:     "root",
		Password: "123456", // 替换为你的密码
		DBName:   "douyin_mall",
	}
}
