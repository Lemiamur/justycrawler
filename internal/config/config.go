package config

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DefaultWorkerCount = 10
	DefaultMaxDepth    = 2
)

type Config struct {
	StartURL     string `mapstructure:"start_url"`
	SameHost     bool   `mapstructure:"same_host"`
	MaxDepth     int    `mapstructure:"max_depth"`
	WorkerCount  int    `mapstructure:"worker_count"`
	ForceRecrawl bool   `mapstructure:"force_recrawl"`
	HTTP         HTTP   `mapstructure:"http"`
	Mongo        Mongo  `mapstructure:"mongo"`
	Redis        Redis  `mapstructure:"redis"`
	Log          Log    `mapstructure:"log"`
}

type HTTP struct {
	Timeout time.Duration `mapstructure:"timeout"`
}

type Mongo struct {
	URI        string `mapstructure:"uri"`
	Database   string `mapstructure:"database"`
	Collection string `mapstructure:"collection"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	SetKey   string `mapstructure:"set_key"`
}

type Log struct {
	Level string `mapstructure:"level"`
}

// Порядок приоритетов: флаги > переменные окружения > файл config.yaml
func New() (*Config, error) {
	viper.SetDefault("http.timeout", "30s")
	viper.SetDefault("mongo.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "crawler_db")
	viper.SetDefault("mongo.collection", "links")

	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.set_key", "crawler:visited_urls")

	viper.SetDefault("worker_count", DefaultWorkerCount)
	viper.SetDefault("max_depth", DefaultMaxDepth)
	viper.SetDefault("same_host", true)
	viper.SetDefault("log.level", "info")

	pflag.String("start_url", "", "Стартовый URL для краулинга (обязательно)")
	pflag.Bool("same_host", viper.GetBool("same_host"), "Ограничить обход только стартовым хостом")
	pflag.Int("max_depth", viper.GetInt("max_depth"), "Максимальная глубина обхода")
	pflag.Int("worker_count", viper.GetInt("worker_count"), "Количество одновременных воркеров")
	pflag.Bool("force_recrawl", false, "Очистить состояние перед запуском для принудительного повторного обхода")
	pflag.Duration("http.timeout", viper.GetDuration("http.timeout"), "Таймаут для HTTP запросов")
	pflag.String("mongo.uri", viper.GetString("mongo.uri"), "URI для подключения к MongoDB")
	pflag.String("redis.addr", viper.GetString("redis.addr"), "Адрес для подключения к Redis (host:port)")
	pflag.String("log.level", viper.GetString("log.level"), "Уровень логирования (debug, info, warn, error)")

	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return nil, err
	}

	viper.SetConfigName("config")    // имя файла без расширения
	viper.SetConfigType("yaml")      // тип файла
	viper.AddConfigPath(".")         // искать в текущей директории
	viper.AddConfigPath("./configs") // и в директории configs

	_ = viper.ReadInConfig()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.StartURL == "" {
		return nil, errors.New("необходимо указать стартовый URL через флаг -start_url или в конфиге")
	}

	return &cfg, nil
}
