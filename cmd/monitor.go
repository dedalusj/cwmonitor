package cmd

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/pkg/errors"
	"mon-put-data/metrics"
	"mon-put-data/util"

	log "github.com/sirupsen/logrus"
)

func buildMetrics(s string) []metrics.Metric {
	metricsSet := map[string]bool{}
	for _, m := range strings.Split(s, ",") {
		metricsSet[m] = true
	}
	
	collectedMetrics := make([]metrics.Metric, 0, len(metricsSet))
	for m := range metricsSet {
		switch m {
		case "memory":
			collectedMetrics = append(collectedMetrics, metrics.Memory{})
		case "swap":
			collectedMetrics = append(collectedMetrics, metrics.Swap{})
		case "disk":
			collectedMetrics = append(collectedMetrics, metrics.Disk{})
		case "cpu":
			collectedMetrics = append(collectedMetrics, metrics.CPU{})
		case "docker-stats":
			collectedMetrics = append(collectedMetrics, metrics.DockerStat{})
		case "docker-health":
			collectedMetrics = append(collectedMetrics, metrics.DockerHealth{})
		case "":
			continue
		default:
			log.Warnf("unknown metric: %s", m)
		}
	}

	return collectedMetrics
}

func Gather(collectedMetrics []metrics.Metric) metrics.Data {
	data := metrics.Data{}
	for _, metric := range collectedMetrics {
		d, err := metric.Gather()
		if err != nil {
			log.Errorf("failed to Monitor data from metric [%s]: %s", metric.Name(), err)
			continue
		}

		data = append(data, d...)
	}

	return data
}

func convertPointToCloudWatch(point *metrics.Point) cloudwatch.MetricDatum {
	datum := cloudwatch.MetricDatum{
		MetricName: aws.String(point.Name),
		Unit:       aws.String(string(point.Unit)),
		Value:      aws.Float64(point.Value),
		Timestamp:  aws.Time(point.Timestamp),
		Dimensions: make([]*cloudwatch.Dimension, 0, len(point.Dimensions)),
	}

	for _, dim := range point.Dimensions {
		datum.Dimensions = append(datum.Dimensions, &cloudwatch.Dimension{
			Name: aws.String(dim.Name),
			Value: aws.String(dim.Value),
		})
	}
	return datum
}

func convertDataToCloudWatch(data metrics.Data) []*cloudwatch.MetricDatum {
	output := make([]*cloudwatch.MetricDatum, 0, len(data))
	for _, d := range data {
		datum := convertPointToCloudWatch(d)
		output = append(output, &datum)
	}
	return output
}

func Put(data metrics.Data, client cloudwatchiface.CloudWatchAPI, namespace string) error {
	multierror := util.MultiError{}
	for _, d := range data.Batch(20) {
		datum := convertDataToCloudWatch(d)
		_, err := client.PutMetricData(&cloudwatch.PutMetricDataInput{
			Namespace: aws.String(namespace),
			MetricData: datum,
		})
		if err != nil {
			multierror.Add(err)
		}
	}

	return errors.Wrap(multierror.ErrorOrNil(), "failed to put metric data")
}

func Monitor(metrics []metrics.Metric, extraDimensions []metrics.Dimension, client cloudwatchiface.CloudWatchAPI, namespace string) {
	data := Gather(metrics)
	data.AddDimensions(extraDimensions...)
	err := Put(data, client, namespace)
	if err != nil {
		log.Errorf("failed to Put metrics: %s", err)
	}
}

type Config struct {
	Namespace string
	Interval  time.Duration
	Id        string
	Metrics   string
	Once      bool
	Version   string
}

func (c Config) Validate() error {
	err := util.MultiError{}

	if c.Namespace == "" {
		err.Add(errors.New("namespace cannot be empty"))
	}
	if c.Interval == time.Duration(0) {
		err.Add(errors.New("interval cannot be zero"))
	}
	if c.Id == "" {
		err.Add(errors.New("id cannot be empty"))
	}
	if c.Metrics == "" {
		err.Add(errors.New("metrics cannot be empty"))
	}

	return err
}

func Exec(c Config) error {
	log.Infof("mon-Put-data - version: %s", c.Version)

	err := c.Validate()
	if err != nil {
		return errors.Wrap(err, "invalid inputs")
	}

	extraDimensions, _ := metrics.MapToDimensions(map[string]string{"machine": c.Id})
	collectedMetrics := buildMetrics(c.Metrics)

	sess := session.Must(session.NewSession())
	svc := cloudwatch.New(sess)

	if c.Once {
		Monitor(collectedMetrics, extraDimensions, svc, c.Namespace)
	} else {
		log.Infof("interval: %d minutes", c.Interval)
		for range time.Tick(c.Interval) {
			Monitor(collectedMetrics, extraDimensions, svc, c.Namespace)
		}
	}

	return nil
}
