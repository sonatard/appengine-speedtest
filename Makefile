deploy:
	gcloud app deploy --quiet app.yaml queue.yaml


endpoint=https://$(proj).appspot.com

curl:
	curl $(endpoint)/_/health
	curl $(endpoint)/cloud/datastore/put
	curl $(endpoint)/cloud/datastore/get
	curl $(endpoint)/appengine/datastore/put
	curl $(endpoint)/appengine/datastore/get
	curl $(endpoint)/cloud/tasks
	curl $(endpoint)/appengine/taskqueue
