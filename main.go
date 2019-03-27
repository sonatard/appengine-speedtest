package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	clouddatastore "cloud.google.com/go/datastore"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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

	// datastore
	http.HandleFunc("/cloud/datastore/put", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		e := &Entity{
			A: "cloud datastore",
			B: "cloud datastore",
			C: "cloud datastore",
		}

		k := clouddatastore.NameKey("Entity", "cloudID", nil)
		if _, err := dsClient.Put(ctx, k, e); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "cloud datastore put %v\n", e)
	})

	http.HandleFunc("/cloud/datastore/get", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		k := clouddatastore.NameKey("Entity", "cloudID", nil)
		e := new(Entity)
		if err := dsClient.Get(ctx, k, e); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "cloud datastore get %v\n", e)
	})

	http.HandleFunc("/appengine/datastore/put", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		e := &Entity{
			A: "appengine datastore",
			B: "appengine datastore",
			C: "appengne datastore",
		}

		k := datastore.NewKey(ctx, "Entity", "appengineID", 0, nil)
		if _, err := datastore.Put(ctx, k, e); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "appengine datastore put %v\n", e)
	})

	http.HandleFunc("/appengine/datastore/get", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		k := datastore.NewKey(ctx, "Entity", "appengineID", 0, nil)
		e := new(Entity)
		if err := datastore.Get(ctx, k, e); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, "appengine datastore get %v\n", e)
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
		_, err := taskqueue.Add(ctx, task, queueID)
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

func createTask(ctx context.Context, client *cloudtasks.Client, uri, projectID, locationID, queueID, message string) (*taskspb.Task, error) {
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

	createdTask, err := client.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("cloudtasks.CreateTask: %v", err)
	}

	return createdTask, nil
}
