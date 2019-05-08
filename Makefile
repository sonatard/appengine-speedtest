deploy:
	gcloud app deploy --quiet app.yaml queue.yaml

benchmark:
	./benchmark.sh
