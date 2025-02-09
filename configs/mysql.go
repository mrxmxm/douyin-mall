package configs

type MySQLConfig struct {
	Host     string // 数据库主机地址
	Port     int    // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名称
}

func NewMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Host:     "localhost",
		Port:     3308,
		User:     "root",
		Password: "123456", // 实际环境中应从配置文件或环境变量读取
		DBName:   "douyin_mall",
	}
}
