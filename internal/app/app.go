package app

import (
	"context"
	"fmt"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/cache"
	"github.com/alexjoedt/echosight/internal/crypt"
	"github.com/alexjoedt/echosight/internal/eventflow"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/http"
	"github.com/alexjoedt/echosight/internal/influx"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/alexjoedt/echosight/internal/mail"
	"github.com/alexjoedt/echosight/internal/notify"
	engine "github.com/alexjoedt/echosight/internal/observer"
	"github.com/alexjoedt/echosight/internal/postgres"
	"github.com/redis/go-redis/v9"
)

func Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// init logger
	logger.Init("debug", true)
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	crypter := crypt.New([]byte(config.Secret))

	// set loglevel after load config
	logger.Init(config.LogLevel, true)

	// Init PostgreSQL Database
	logger.Debugf("Connect to Database...")
	db, err := postgres.New(config.postgresDSN())
	if err != nil {
		return err
	}
	defer db.Close()

	go db.Sessions.StartCleanup(time.Second * 60)
	defer db.Sessions.StopCleanup()

	// Init InfluxDB
	logger.Debugf("Connect to InlfuxDB...")
	influxClient, err := influx.New(config.InfluxURL(), config.InfluxDB.Token)
	if err != nil {
		return err
	}

	// Init EventHandler
	logger.Debugf("Initialize Event-Engine...")
	eventHandler := eventflow.NewEngine()
	defer eventHandler.Stop()

	// Init Notification-Service
	// TODO: read all enabled notification services from the DB and initialize them
	// TODO: create a sql table for notification services
	// TODO: add API route to initalize and enable notification services
	notifier := notify.NewNotifier()

	// Init MailService
	logger.Infof("Initialize Mail-Service")
	mailer, err := mail.New(mail.Opts{
		AppPreferences: &db.Preferences,
		Recipients:     &db.Recipients,
		Crypter:        crypter,
	})
	if err != nil {
		logger.Warnf("failed to initialize mailer: %v", err)
	} else {
		notifier.AddSender("mail", mailer)
	}

	// Init Telegram Bot
	tele, err := notify.NewTelegramBot(&db.Preferences)
	if err != nil {
		logger.Warnf("failed to init telegram bot: %v", err)
	} else {
		notifier.AddSender("telegram", tele)
	}

	// Init observer engine and starts
	logger.Debugf("Initialize Observer-Engine...")
	scheduler := engine.NewScheduler(&db.Detectors, influxClient, eventHandler, notifier) // the eventhandler is the server and the server gets the Observer-Engine as dependency

	// load all active detectors
	dFilter := filter.NewDefaultDetectorFilter()
	active := true
	dFilter.Active = &active
	detectors, err := db.Detectors.List(ctx, dFilter)
	if err != nil {
		return fmt.Errorf("failed to load detectors at startup: %v", err)
	}
	scheduler.AddDetectors(detectors...)

	// TODO: read state of observer/scheduler from database
	scheduler.Start()

	// Init redis
	logger.Debugf("Connect to Redis...")
	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       0,
	})
	_, err = rc.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Init Cache
	logger.Debugf("Initialize Cache...")
	var cacheStore echosight.Cache
	switch config.Cache.CacheType {
	case CacheRedis:
		cacheStore = cache.NewRedisCache(rc, time.Duration(config.Cache.TTL))
	default:
		cacheStore = cache.NewMemoryCache()
	}

	logger.Debugf("Initialize Web-Server...")
	server, err := http.NewServer(http.ServerOpts{
		Addr:           config.HTTP.Port,
		TrustedOrigins: config.HTTP.TrustedOrigins,
	})
	if err != nil {
		return err
	}

	server.EventHandler = eventHandler
	server.Scheduler = scheduler
	server.UserService = cache.NewUserCache(cacheStore, time.Minute*15, &db.Users)
	server.DetectorService = &db.Detectors
	server.HostService = &db.Hosts
	server.RecipientService = &db.Recipients
	server.PreferenceService = &db.Preferences
	server.SessionService = &db.Sessions
	server.MetricReader = influxClient
	server.Crypter = crypter

	server.IsDev = config.isDev()
	server.RateLimiter = http.RateLimiter{
		Enabled: config.HTTP.Limiter.Enabled,
		Burst:   config.HTTP.Limiter.Burst,
		Limit:   int(config.HTTP.Limiter.RateLimit),
	}

	logger.Infof("Starting EchoSight Server")
	logger.Infof("Environment: %s", config.Environment)
	logger.Infof("Port: %s", config.HTTP.Port)
	logger.Infof("Version: %s", echosight.Version)
	logger.Infof("Revision: %s", echosight.Revision)

	err = server.Run()
	if err != nil {
		return err
	}

	logger.Infof("Server stopped")
	logger.Infof("Waiting for background jobs")
	scheduler.Stop()
	logger.Infof("Shutdown")
	return nil
}
