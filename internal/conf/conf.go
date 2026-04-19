package conf

// Bootstrap 是应用聚合配置，按层划分常用配置。
type Bootstrap struct {
	Server Server `mapstructure:"server"`
	Data   Data   `mapstructure:"data"`
	Biz    Biz    `mapstructure:"biz"`
	Worker Worker `mapstructure:"worker"`
}

type Server struct {
	HTTP HTTP `mapstructure:"http"`
	CORS CORS `mapstructure:"cors"`
}

type HTTP struct {
	Addr string `mapstructure:"addr"`
}

type CORS struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type Data struct {
	Database Database `mapstructure:"database"`
	Redis    Redis    `mapstructure:"redis"`
}

type Database struct {
	DSN             string `mapstructure:"dsn"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime_minutes"`
}

type Redis struct {
	Addr string `mapstructure:"addr"`
}

type Biz struct {
	Task     Task     `mapstructure:"task"`
	Features Features `mapstructure:"features"`
}

type Task struct {
	DefaultScope string `mapstructure:"default_scope"`
}

type Features struct {
	EmailService EmailServiceFeature `mapstructure:"email_service"`
}

type EmailServiceFeature struct {
	Enabled bool `mapstructure:"enabled"`
}

type Worker struct {
	RedisAddr string `mapstructure:"redis_addr"`
	SMTPEmail string `mapstructure:"smtp_email"`
	SMTPToken string `mapstructure:"smtp_token"`
}

func (b *Bootstrap) ApplyDefaults() {
	if b == nil {
		return
	}
	if b.Server.HTTP.Addr == "" {
		b.Server.HTTP.Addr = ":8080"
	}
	if len(b.Server.CORS.AllowOrigins) == 0 {
		b.Server.CORS.AllowOrigins = []string{"http://localhost:5173"}
	}
	if b.Data.Database.DSN == "" {
		b.Data.Database.DSN = "postgres://root:123@localhost:5432/root?sslmode=disable&TimeZone=Asia/Shanghai"
	}
	if b.Data.Database.MaxIdleConns <= 0 {
		b.Data.Database.MaxIdleConns = 10
	}
	if b.Data.Database.MaxOpenConns <= 0 {
		b.Data.Database.MaxOpenConns = 100
	}
	if b.Data.Database.ConnMaxLifetime <= 0 {
		b.Data.Database.ConnMaxLifetime = 30
	}
	if b.Data.Redis.Addr == "" {
		b.Data.Redis.Addr = "127.0.0.1:16379"
	}
	if b.Worker.RedisAddr == "" {
		b.Worker.RedisAddr = b.Data.Redis.Addr
	}
	if b.Worker.SMTPEmail == "" {
		b.Worker.SMTPEmail = "1056652209@qq.com"
	}
	if b.Biz.Task.DefaultScope == "" {
		b.Biz.Task.DefaultScope = "my_todo"
	}
	if !b.Biz.Features.EmailService.Enabled {
		b.Biz.Features.EmailService.Enabled = false
	}
}
