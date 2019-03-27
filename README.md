# appengine-speedtest

## Usage

```
export proj=[YOURPROJECT]

$ gcloud projects create ${proj}
$ gcloud services enable cloudtasks.googleapis.com
$ gcloud config set project ${proj}
$ make deploy

$ make curl
```

## Speed test result

 asia-northeast
 
![sample0](sample0.png)

![sample1](sample1.png)

![sample2](sample2.png)
