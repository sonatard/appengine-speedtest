package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	clouddatastore "cloud.google.com/go/datastore"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

type Entity struct {
	A string
	B string
	C string
}

var (
	projectID  = os.Getenv("GOOGLE_CLOUD_PROJECT")
	queueID    = "queue"
	locationID = "asia-northeast1"
)

func benchmark(ctx context.Context, f func()) {
	for i := 0; i < 30; i++ {
		start := time.Now()
		f()
		log.Infof(ctx, "latency : %v", time.Since(start))
	}
}

func main() {
	ctx := context.Background()
	dsClient, err := clouddatastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	taskClient, err := cloudtasks.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
	})
	if err != nil {
		panic(err)
	}

	defer sd.Flush()

	trace.RegisterExporter(sd)

	// datastore
	http.HandleFunc("/cloud/datastore/put", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		e := &Entity{
			A: "cloud datastore",
			B: "cloud datastore",
			C: "cloud datastore",
		}

		k := clouddatastore.NameKey("Entity", "cloudID", nil)
		benchmark(ctx, func() {
			_, err = dsClient.Put(ctx, k, e)
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "cloud datastore put %v %v\n", e)
	})

	http.HandleFunc("/cloud/datastore/get", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		k := clouddatastore.NameKey("Entity", "cloudID", nil)
		e := new(Entity)
		benchmark(ctx, func() {
			err = dsClient.Get(ctx, k, e)
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "cloud datastore get %v\n", e)
	})

	http.HandleFunc("/cloud/datastore/txgetput", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error
		benchmark(ctx, func() {
			_, err = dsClient.RunInTransaction(ctx, func(tx *clouddatastore.Transaction) error {
				k := clouddatastore.NameKey("Entity", "cloudID", nil)
				e := new(Entity)
				err := tx.Get(k, e)
				if err != nil {
					return err
				}
				_, err = tx.Put(k, e)
				if err != nil {
					return err
				}

				return nil
			})
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "cloud datastore txgetput \n")
	})

	http.HandleFunc("/appengine/datastore/put", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		e := &Entity{
			A: "appengine datastore",
			B: "appengine datastore",
			C: "appengne datastore",
		}

		k := datastore.NewKey(ctx, "Entity", "appengineID", 0, nil)
		benchmark(ctx, func() {
			_, err = datastore.Put(ctx, k, e)
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "appengine datastore put %v\n", e)
	})

	http.HandleFunc("/appengine/datastore/get", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		k := datastore.NewKey(ctx, "Entity", "appengineID", 0, nil)
		e := new(Entity)
		benchmark(ctx, func() {
			err = datastore.Get(ctx, k, e)
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "appengine datastore get %v\n", e)
	})

	http.HandleFunc("/appengine/datastore/txgetput", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		var err error
		benchmark(ctx, func() {
			err = datastore.RunInTransaction(ctx, func(tx context.Context) error {
				k := datastore.NewKey(tx, "Entity", "appengineID", 0, nil)
				e := new(Entity)
				err := datastore.Get(tx, k, e)
				if err != nil {
					return err
				}
				_, err = datastore.Put(tx, k, e)
				if err != nil {
					return err
				}

				return nil
			}, nil)
		})

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "appengine datastore txgetput\n")
	})

	// taskqueue
	http.HandleFunc("/cloud/tasks", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if _, err := createTask(ctx, taskClient, "/_/health", projectID, locationID, queueID, "hello"); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "cloud tasks")
	})

	http.HandleFunc("/appengine/taskqueue", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		task := &taskqueue.Task{
			Path:    "/_/health",
			Payload: []byte("hello"),
		}
		benchmark(ctx, func() {
			_, err = taskqueue.Add(ctx, task, queueID)
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "appengine taskqueue\n")
	})

	http.HandleFunc("/_/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "HealthCheck: OK\n")
	})

	appengine.Main()
}

func createTask(ctx context.Context, client *cloudtasks.Client, uri, projectID, locationID, queueID, message string) (createdTask *taskspb.Task, err error) {
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID)

	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_POST,
					RelativeUri: uri,
				},
			},
		},
	}

	if message != "" {
		req.Task.GetAppEngineHttpRequest().Body = []byte(message)
	}

	benchmark(ctx, func() {
		createdTask, err = client.CreateTask(ctx, req)
	})
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %v", err)
	}

	return createdTask, nil
}
