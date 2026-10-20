package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	Database			string
	ConnectionString	string
	SecretCode			string

	Port				string

	UploadPath			string
)

func init() {
	UploadPath = "webdata"
	Database = "mysql"
	Port = "9003"


	// err := godotenv.Load()

  	// if err != nil {
    // 	log.Fatal("Error loading .env file")
  	// }

	// Database = os.Getenv("DATABASE")
	// ConnectionString = os.Getenv("DATABASE_URL")
	// SecretCode = os.Getenv("SECRET_CODE")


	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath(".")
	// config.json 은 시크릿을 담으므로 커밋하지 않는다(config.json.example 참고).
	// 파일이 없으면 환경변수(DATABASE_URL, SECRET_CODE)만으로도 기동할 수 있다.
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("config file not loaded: %s \n", err)
	}

	if value := viper.Get("connectionString"); value != nil {
		ConnectionString = value.(string)
	}

	if value := viper.Get("Database"); value != nil {
		Database = value.(string)
	}

	if value := viper.Get("SecretCode"); value != nil {
		SecretCode = value.(string)
	}

	if value := viper.Get("port"); value != nil {
		Port = value.(string)
	}

	if value := viper.Get("uploadPath"); value != nil {
		UploadPath = value.(string)
	}

	// 환경변수가 있으면 설정 파일 값보다 우선한다.
	if value := os.Getenv("DATABASE"); value != "" {
		Database = value
	}
	if value := os.Getenv("DATABASE_URL"); value != "" {
		ConnectionString = value
	}
	if value := os.Getenv("SECRET_CODE"); value != "" {
		SecretCode = value
	}

	if ConnectionString == "" {
		panic("connectionString is not set (config/config.json or DATABASE_URL)")
	}
	// JWT 서명키가 비어 있으면 누구나 토큰을 위조할 수 있으므로 기동하지 않는다.
	if SecretCode == "" {
		panic("secretCode is not set (config/config.json or SECRET_CODE)")
	}
}