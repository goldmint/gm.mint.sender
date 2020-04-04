package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/pubsub"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

var (
	_ Alerter = &GCloudAlerter{}
)

// OnGCE reports whether this process is running on Google Compute Engine
func OnGCE() bool {
	return metadata.OnGCE()
}

// GCloudAlerter sends alerts via Google Cloud Pub/Sub
type GCloudAlerter struct {
	logger  *logrus.Entry
	origin  string
	client  *pubsub.Client
	topic   *pubsub.Topic
	limiter *timeLimiter
}

// NewGCloud instance
func NewGCloud(origin string, logger *logrus.Entry) (*GCloudAlerter, func(), error) {

	// creds
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel1()
	creds, err := google.FindDefaultCredentials(ctx1, compute.ComputeScope)
	if err != nil {
		return nil, nil, err
	}

	// new cli
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel2()
	cli, err := pubsub.NewClient(ctx2, creds.ProjectID)
	if err != nil {
		return nil, nil, err
	}

	// check topic
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel3()
	topic := cli.Topic("admin_alert.telegram")
	ok, err := topic.Exists(ctx3)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, fmt.Errorf("topic does not exist")
	}

	return &GCloudAlerter{
			logger:  logger,
			client:  cli,
			topic:   topic,
			origin:  origin,
			limiter: newTimeLimiter(),
		}, func() {
			topic.Stop()
			cli.Close()
		}, nil
}

// Info implementation
func (a *GCloudAlerter) Info(f string, arg ...interface{}) {
	a.send("info", f, arg...)
}

// Warn implementation
func (a *GCloudAlerter) Warn(f string, arg ...interface{}) {
	a.send("warning", f, arg...)
}

// Errorf implementation
func (a *GCloudAlerter) Error(f string, arg ...interface{}) {
	a.send("error", f, arg...)
}

// LimitInfo implementation
func (a *GCloudAlerter) LimitInfo(max time.Duration, f string, arg ...interface{}) {
	if a.limiter.limit(max, f, arg...) {
		a.Info(f, arg...)
	}
}

// LimitWarn implementation
func (a *GCloudAlerter) LimitWarn(max time.Duration, f string, arg ...interface{}) {
	if a.limiter.limit(max, f, arg...) {
		a.Warn(f, arg...)
	}
}

// LimitError implementation
func (a *GCloudAlerter) LimitError(max time.Duration, f string, arg ...interface{}) {
	if a.limiter.limit(max, f, arg...) {
		a.Error(f, arg...)
	}
}

// ---

func (a *GCloudAlerter) send(level string, f string, arg ...interface{}) {
	// message
	data := struct {
		From    string `json:"from"`
		Level   string `json:"level"`
		Message string `json:"message"`
	}{
		a.origin,
		strings.ToLower(level),
		fmt.Sprintf(f, arg...),
	}
	b, err := json.Marshal(data)
	if err != nil {
		a.logger.WithError(err).Errorf("Failed to marshal message")
		return
	}

	// publish
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	result := a.topic.Publish(ctx, &pubsub.Message{
		Data: b,
	})
	if _, err := result.Get(ctx); err != nil {
		a.logger.WithError(err).Errorf("Failed to publish message")
		return
	}
}
