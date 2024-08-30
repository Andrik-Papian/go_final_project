package config

import (
    "os"

    "github.com/joho/godotenv"
    log "github.com/sirupsen/logrus"
)

type Config struct {
    Port     string
    DBFile   string
    Password string
    LogLevel log.Level
}

func New() (*Config, error) {
    err := godotenv.Load(".env")
    if err != nil {
        return nil, err
    }

    cfg := Config{
        Port:     os.Getenv("TODO_PORT"),
        DBFile:   os.Getenv("TODO_DBFILE"),
        Password: os.Getenv("TODO_PASSWORD"),
    }

    logLevel, err := log.ParseLevel(os.Getenv("TODO_LOGLEVEL"))
    if err != nil {
        return nil, err
    }

    cfg.LogLevel = logLevel

    return &cfg, nil
}

