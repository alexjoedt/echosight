package mail

type SMTPConfig struct {
	Sender        string `mapstructure:"smtp_sender"`
	Host          string `mapstructure:"smtp_host"`
	Port          string `mapstructure:"smtp_port"`
	User          string `mapstructure:"smtp_user"`
	Enabled       bool   `mapstructure:"smtp_enabled"`
	PasswordCrypt string `mapstructure:"smtp_password_crypt"`
	password      string
}
