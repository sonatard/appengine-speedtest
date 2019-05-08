# appengine-speedtest

## Usage

1. Fix region
https://github.com/sonatard/appengine-speedtest/blob/7dca4729a5c5f360a01c72a50540818f827a0a1d/main.go#L30


2. Create project and deploy

```
export proj=[YOURPROJECT]
export region=[YOURREGION]

$ gcloud config set project ${proj}
$ gcloud projects create ${proj}
$ gcloud app create --region=${region}
$ gcloud services enable cloudtasks.googleapis.com
$ make deploy

$ make benchmark
```

[Benchmark](https://github.com/sonatard/appengine-speedtest/issues/1)
