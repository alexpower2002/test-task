package utils

import (
	"os"
	"strconv"
	"time"

	"mkk-luna-test-task/internal/repository"

	"github.com/joho/godotenv"
)

type Envs struct {
	Port           int
	JwtSecret      string
	JwtExpiration  time.Duration
	MigrationsPath string

	MySQL    repository.MySQLConfig
	Redis    repository.RedisConfig
	RedisTTL time.Duration
}

const envFilePath = ".env"

func LoadEnvs() (Envs, error) {
	_ = godotenv.Load(envFilePath)

	var envs Envs
	var err error

	portStr := os.Getenv("PORT")

	envs.Port, err = strconv.Atoi(portStr)

	if err != nil {
		return Envs{}, err
	}

	envs.JwtSecret = os.Getenv("JWT_SECRET")

	jwtExpStr := os.Getenv("JWT_EXPIRATION_HOURS")

	expHours, err := strconv.Atoi(jwtExpStr)

	if err != nil {
		return Envs{}, err
	}

	envs.JwtExpiration = time.Duration(expHours) * time.Hour

	envs.MigrationsPath = os.Getenv("MIGRATIONS_PATH")

	mysqlPortStr := os.Getenv("MYSQL_PORT")

	mysqlPort, err := strconv.Atoi(mysqlPortStr)

	if err != nil {
		return Envs{}, err
	}

	maxOpenConnsStr := os.Getenv("MYSQL_MAX_OPEN_CONNS")

	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)

	if err != nil {
		return Envs{}, err
	}

	maxIdleConnsStr := os.Getenv("MYSQL_MAX_IDLE_CONNS")

	maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)

	if err != nil {
		return Envs{}, err
	}

	connMaxLifetimeMinsStr := os.Getenv("MYSQL_CONN_MAX_LIFETIME_MINS")

	connMaxLifetimeMins, err := strconv.Atoi(connMaxLifetimeMinsStr)

	if err != nil {
		return Envs{}, err
	}

	envs.MySQL = repository.MySQLConfig{
		User:                os.Getenv("MYSQL_USER"),
		Password:            os.Getenv("MYSQL_PASSWORD"),
		Host:                os.Getenv("MYSQL_HOST"),
		Port:                mysqlPort,
		DBName:              os.Getenv("MYSQL_DBNAME"),
		MaxOpenConns:        maxOpenConns,
		MaxIdleConns:        maxIdleConns,
		ConnMaxLifetimeMins: connMaxLifetimeMins,
	}

	redisDBStr := os.Getenv("REDIS_DB")

	redisDB, err := strconv.Atoi(redisDBStr)

	if err != nil {
		return Envs{}, err
	}

	redisDialTimeoutStr := os.Getenv("REDIS_DIAL_TIMEOUT_SECONDS")

	redisDialTimeoutSec, err := strconv.Atoi(redisDialTimeoutStr)

	if err != nil {
		return Envs{}, err
	}

	redisPoolSizeStr := os.Getenv("REDIS_POOL_SIZE")

	redisPoolSize, err := strconv.Atoi(redisPoolSizeStr)

	if err != nil {
		return Envs{}, err
	}

	redisMinIdleConnsStr := os.Getenv("REDIS_MIN_IDLE_CONNS")

	redisMinIdleConns, err := strconv.Atoi(redisMinIdleConnsStr)

	if err != nil {
		return Envs{}, err
	}

	envs.Redis = repository.RedisConfig{
		Addr:         os.Getenv("REDIS_ADDR"),
		Password:     os.Getenv("REDIS_PASSWORD"),
		DB:           redisDB,
		DialTimeout:  time.Duration(redisDialTimeoutSec) * time.Second,
		PoolSize:     redisPoolSize,
		MinIdleConns: redisMinIdleConns,
	}

	redisTTLStr := os.Getenv("REDIS_TTL_MINUTES")

	redisTTLMinutes, err := strconv.Atoi(redisTTLStr)

	if err != nil {
		return Envs{}, err
	}

	envs.RedisTTL = time.Duration(redisTTLMinutes) * time.Minute

	return envs, nil
}
