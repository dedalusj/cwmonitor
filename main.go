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
	"github.com/urfave/cli"

	log "github.com/sirupsen/logrus"
)

var version string

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
		Interval:  time.Duration(c.Int("interval")) * time.Minute,
		HostId:    c.String("hostid"),
		Metrics:   c.String("metrics"),
		Once:      c.Bool("once"),
		Version:   c.App.Version,
		Client:    client,
	}
}

func setupCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	defer func() {
		signal.Stop(sigCh)
		cancel()
	}()
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
			Usage:  "Comma separated list of metrics",
			Value:  "cpu,memory",
			EnvVar: "CWMONITOR_METRICS",
		},
		cli.IntFlag{
			Name:   "interval",
			Usage:  "Time interval",
			Value:  1,
			EnvVar: "CWMONITOR_INTERVAL",
		},
		cli.BoolFlag{
			Name:  "once",
			Usage: "Run once (i.e. not on an interval)",
		},
		cli.StringFlag{
			Name:   "namespace",
			Usage:  "Namespace for the metric data",
			Value:  "CWMonitor",
			EnvVar: "CWMONITOR_NAMESPACE",
		},
		cli.StringFlag{
			Name:   "hostid",
			Usage:  "ID of the current host",
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
