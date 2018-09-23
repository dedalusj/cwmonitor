package monitor

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatch/cloudwatchiface"
	"github.com/dedalusj/cwmonitor/metrics"
	"github.com/dedalusj/cwmonitor/util"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

// GatherData collects data for the given metrics
func GatherData(collectedMetrics []metrics.Metric) metrics.Data {
	log.Debugf("gathering data for %d metrics", len(collectedMetrics))
	data := metrics.Data{}
	for _, metric := range collectedMetrics {
		d, err := metric.Gather()
		if err != nil {
			log.Errorf("failed to gather data from metric [%s]: %s", metric.Name(), err)
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
			Name:  aws.String(dim.Name),
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

// PublishDataToCloudWatch publish the given data to CloudWatch
// under the provided namespace using the given client
func PublishDataToCloudWatch(data metrics.Data, namespace string, client cloudwatchiface.CloudWatchAPI) error {
	log.Debug("publishing data points")
	multierror := util.MultiError{}
	for _, d := range data.Batch(20) {
		datum := convertDataToCloudWatch(d)
		_, err := client.PutMetricData(&cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(namespace),
			MetricData: datum,
		})
		if err != nil {
			multierror.Add(err)
		}
	}

	return errors.Wrap(multierror.ErrorOrNil(), "failed to put metric data")
}

// Monitor gathers the data from the requested metrics, adds extra dimensions if requests and publish
// them to CloudWatch under the provided namespace using the given client.
func Monitor(metrics []metrics.Metric, extraDimensions []metrics.Dimension, namespace string, client cloudwatchiface.CloudWatchAPI) {
	data := GatherData(metrics)
	data.AddDimensions(extraDimensions...)
	err := PublishDataToCloudWatch(data, namespace, client)
	if err != nil {
		log.Errorf("failed to publish data to cloudwatch: %s", err)
	} else {
		log.Infof("published %d data points to namespace [%s]", len(data), namespace)
	}
}

// Run the monitor command
func Run(ctx context.Context, c Config) error {
	err := c.validate()
	if err != nil {
		return errors.Wrap(err, "invalid inputs")
	}

	c.logConfig()
	log.Info("starting monitoring")
	Monitor(c.getRequestedMetrics(), c.getExtraDimensions(), c.Namespace, c.Client)
	if !c.Once {
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			ticker := c.getTicker()
			for {
				select {
				case <-ticker.C:
					Monitor(c.getRequestedMetrics(), c.getExtraDimensions(), c.Namespace, c.Client)
				case <-ctx.Done():
					log.Info("stopping monitoring")
					return
				}
			}
		}()

		wg.Wait()
	}

	return nil
}
