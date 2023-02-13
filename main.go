package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Edbeer/paymentapi/api"
	"github.com/Edbeer/paymentapi/config"
	"github.com/Edbeer/paymentapi/storage/psql"
	"github.com/Edbeer/paymentapi/storage/redis"
	"github.com/sirupsen/logrus"

	"github.com/Edbeer/paymentapi/pkg/db/psql"
	"github.com/Edbeer/paymentapi/pkg/db/redis"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jConfig "github.com/uber/jaeger-client-go/config"
	jLog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

// @title           Payment Application
// @version         1.0
// @description     Simple payment system

// @securitydefinitions.apikey
// @in header
// @name Authorization
func main() {
	// init config
	config := config.GetConfig()
	// init logrus
	log := logrus.New()
	// init postgres
	db, err := psql.NewPostgresDB(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("init postgres")

	// init redis
	redisClient := red.NewRedisClient()
	defer redisClient.Close()
	log.Println("init redis")

	psql := postgres.NewPostgresStorage(db)
	redisStore := redisrepo.NewRedisStorage(redisClient)

	jaegerCfgInstance := jConfig.Configuration{
		ServiceName: "PAYMENT_API",
		Sampler: &jConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 10,
		},
		Reporter: &jConfig.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: "jaeger:6831",
		},
	}

	tracer, closer, err := jaegerCfgInstance.NewTracer(
		jConfig.Logger(jLog.StdLogger),
		jConfig.Metrics(metrics.NullFactory),
	)
	if err != nil {
		log.Fatal("cannot create tracer", err)
	}
	log.Println("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	log.Println("Opentracing connected")

	// init server
	log.Println("init server")
	s := api.NewJSONApiServer(config, db, redisClient, psql, redisStore, log)
	go func() {
		s.Run()
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<- quitCh

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.Server.Shutdown(ctx)
}