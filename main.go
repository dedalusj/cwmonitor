package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/dedalusj/cwmonitor/monitor"
	"github.com/dedalusj/cwmonitor/util"
	"gopkg.in/urfave/cli.v1"

	log "github.com/sirupsen/logrus"
)

var version = "dev"
var buildTime = time.Now().Format("20060102T150405Z")
var buildNumber = "local"

func initLogger(c *cli.Context) {
	log.SetFormatter(&util.Formatter{})
	log.SetLevel(log.InfoLevel)
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}
}

func getConfig(c *cli.Context) monitor.Config {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := cloudwatch.New(sess)

	return monitor.Config{
		Namespace: c.String("namespace"),
		Interval:  time.Duration(c.Int("interval")) * time.Second,
		HostId:    c.String("hostid"),
		Metrics:   c.String("metrics"),
		Once:      c.Bool("once"),
		Metadata:  util.Metadata{Version: c.App.Version, BuildTime: buildTime, BuildNumber: buildNumber},
		Client:    client,
	}
}

func setupCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx
}

func main() {
	app := cli.NewApp()
	app.Name = "cwmonitor"
	app.Usage = "Publish Custom Metrics to CloudWatch"
	app.Version = version
	app.Author = "Jacopo Sabbatini"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "metrics",
			Usage:  "Comma separated list of metrics. Available: cpu, memory, swap, disk, docker-stats, docker-health",
			Value:  "cpu,memory",
			EnvVar: "CWMONITOR_METRICS",
		},
		cli.IntFlag{
			Name:   "interval",
			Usage:  "Time interval between data collection (seconds)",
			Value:  60,
			EnvVar: "CWMONITOR_INTERVAL",
		},
		cli.BoolFlag{
			Name:  "once",
			Usage: "Run once (i.e. not on an interval)",
		},
		cli.StringFlag{
			Name:   "namespace",
			Usage:  "CloudWatch namespace",
			Value:  "CWMonitor",
			EnvVar: "CWMONITOR_NAMESPACE",
		},
		cli.StringFlag{
			Name:   "hostid",
			Usage:  "ID of the current host used as dimension for the upload",
			EnvVar: "CWMONITOR_ID",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug logging",
		},
	}
	app.Action = func(c *cli.Context) error {
		initLogger(c)
		config := getConfig(c)
		ctx := setupCtx()
		err := monitor.Run(config, ctx)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	}
	app.Run(os.Args)
}
