package client

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/viper"
)

var SMTP_TOKEN string
var emailSender EmailSender

func TestMain(m *testing.M) {
	viper.SetConfigName("local")
	viper.AddConfigPath("..")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			panic(err)
		}
	}
	SMTP_TOKEN = viper.GetString("SMTP_TOKEN")
	if SMTP_TOKEN == "" {
		panic("SMTP_TOKEN not found in local.env or environment")
	}
	emailSender = NewEmailClient(SMTP_TOKEN, "1056652209@qq.com")
	os.Exit(m.Run())
}
