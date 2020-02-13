package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/pubsub"
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
	origin    string
	projectID string
	topic     string
	limiter   *timeLimiter
}

// NewGCloud instance
func NewGCloud(origin string) (*GCloudAlerter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	creds, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return nil, err
	}
	return &GCloudAlerter{
		topic:     "admin_alert.telegram",
		projectID: creds.ProjectID,
		origin:    origin,
		limiter:   newTimeLimiter(),
	}, nil
}

// Info implementation
func (a *GCloudAlerter) Info(f string, arg ...interface{}) error {
	return a.send("info", f, arg...)
}

// Warn implementation
func (a *GCloudAlerter) Warn(f string, arg ...interface{}) error {
	return a.send("warning", f, arg...)
}

// Errorf implementation
func (a *GCloudAlerter) Error(f string, arg ...interface{}) error {
	return a.send("error", f, arg...)
}

// LimitInfo implementation
func (a *GCloudAlerter) LimitInfo(max time.Duration, f string, arg ...interface{}) error {
	if a.limiter.limit(max, "") {
		return a.Info(f, arg...)
	}
	return nil
}

// LimitWarn implementation
func (a *GCloudAlerter) LimitWarn(max time.Duration, f string, arg ...interface{}) error {
	if a.limiter.limit(max, "") {
		return a.Warn(f, arg...)
	}
	return nil
}

// LimitError implementation
func (a *GCloudAlerter) LimitError(max time.Duration, f string, arg ...interface{}) error {
	if a.limiter.limit(max, "") {
		return a.Error(f, arg...)
	}
	return nil
}

// LimitTagInfo implementation
func (a *GCloudAlerter) LimitTagInfo(max time.Duration, tag, f string, arg ...interface{}) error {
	if a.limiter.limit(max, tag) {
		return a.Info(f, arg...)
	}
	return nil
}

// LimitTagWarn implementation
func (a *GCloudAlerter) LimitTagWarn(max time.Duration, tag, f string, arg ...interface{}) error {
	if a.limiter.limit(max, tag) {
		return a.Warn(f, arg...)
	}
	return nil
}

// LimitTagError implementation
func (a *GCloudAlerter) LimitTagError(max time.Duration, tag, f string, arg ...interface{}) error {
	if a.limiter.limit(max, tag) {
		return a.Error(f, arg...)
	}
	return nil
}

// ---

func (a *GCloudAlerter) send(level string, f string, arg ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// new cli
	cli, err := pubsub.NewClient(ctx, a.projectID)
	if err != nil {
		return err
	}

	// check topic
	topic := cli.Topic(a.topic)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("topic doesn't exist")
	}

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
		return err
	}

	result := cli.Topic(a.topic).Publish(ctx, &pubsub.Message{
		Data: b,
	})
	<-result.Ready()
	return nil
}
