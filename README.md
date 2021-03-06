Containers allow us to run our code wherever we want, Docker is a popular option to build our containers. However, sometimes you love *Dockerfile* and sometimes (or more) you hate *Dockerfile*.

>"44% of the containers in the wild had know vulnerabilities for which patches existed. This just shows you that writing your own *Dockerfile* can actually be quite fraught with errors unless you're an expert or maybe you have a platform team that provides *Dockerfiles* for you." Src: [Buildpacks in Google Cloud](https://youtu.be/suhCr5W_bFc?t=562) 

Writing a *Dockerfile* could be complicated, considering image size, security, multi-stage, versions and so on, there are too many ways to do that!

Sure, we can handle all that things to get the perfect image but if you are using containers, sooner or later you will have to deploy your container in a **Kubernetes** cluster (like Cloud Run or GKE on GCP) and [Kubernetes has deprecated Docker](https://kubernetes.io/blog/2020/12/02/dont-panic-kubernetes-and-docker/).

We are Devs, we should focus on the code...

So... Here comes **_Buildpacks_** to save the day!

Our main objective for today is to deploy a simple API on Cloud Run without a *Dockerfile* using **_Buildpacks_**.

![Architecture](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/inllpyqfsztshccp81eh.png)

###The Code

We're going to deploy a simple API using Go
```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response definition for response API
type Response struct {
	Message string `json:"message"`
}

func handler(w http.ResponseWriter, r *http.Request) {

	response := Response{}
	response.Message = "Hello world!"
	responseJSON, err := json.Marshal(response)

	if err != nil {
		log.Fatalf("Failed to parse json: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
```

##Using *Dockerfile*
Let's begin with the 'popular' way, I mean using a *Dockerfile*

```dockerfile
# Source: https://github.com/GoogleCloudPlatform/golang-samples/blob/master/run/helloworld/Dockerfile

FROM golang:1.16-buster as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN go build -mod=readonly -v -o server

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server

# Run the web service on container startup.
CMD ["/app/server"]
```

Don't forget the .dockerignore! 
```shell
# The .dockerignore file excludes files from the container build process.
#
# https://docs.docker.com/engine/reference/builder/#dockerignore-file

# Exclude locally vendored dependencies.
vendor/

# Exclude "build-time" ignore files.
.dockerignore
.gcloudignore

# Exclude git history and configuration.
.gitignore
.git
```

With no doubt somebody will say:

-"Hey! your *Dockerfile* is incomplete! with multi-stage your image will be smaller"

And... yes, you're right! with this *Dockerfile* the image is **82.6 mb**, I found somewhere a long time ago a *Dockerfile* with multi-stage for Go, _it's [here](https://github.com/AlvarDev/buildpacks-go-example/blob/dockerfile-multistage/Dockerfile) just for the record_, and the image size was reduced to **13.7 mb**.

Great right? Not at all, First I don't remember where I found the multi-stage *Dockerfile* or if the post even exists today, second for a maintainer it could be difficult to understand the process if an update (security patch) is required.

##Using Buildpacks
**_Buildpacks_** is an open-source technology that makes it fast and easy for you to create secure, production-ready container images from source code and without a Dockerfile, following the best practices. Src: [Announcing Google Cloud buildpacks—container images made easy](https://cloud.google.com/blog/products/containers-kubernetes/google-cloud-now-supports-buildpacks)

The first thing you should do for using **_Buildpacks_** is to remove the *Dockerfile* and the .dockerignore.

That's it! you keep what it matters: your code.

To install buildpacks go to: [https://buildpacks.io/docs/tools/pack/](https://buildpacks.io/docs/tools/pack/)

To build the image use this:
```shell
pack build my-go-api:v0.1 --builder --builder gcr.io/buildpacks/builder
```

To test the image use this:
```shell
docker run --rm -p 8080:8080 my-go-api:v0.1

# In another terminal
curl --request GET --url http://localhost:8080/

# Response expected
# {"message": "Hello world!"}
```
Awesome! we have our image running locally, now let's deploy to Cloud Run
```shell
# You can push the image that you build
# but I want to show you that buildpacks can be
# used with gcloud too.

export PROJECT_ID=$(gcloud config list --format 'value(core.project)')

gcloud alpha builds submit --pack image=gcr.io/$PROJECT_ID/my-go-api:v0.1

# Deploy to Cloud Run
gcloud run deploy my-go-api-service \ 
  --image gcr.io/$PROJECT_ID/my-go-api:v0.1 \
  --region southamerica-east1 \
  --platform managed 

curl https://my-go-api-service-[your-hash]-rj.a.run.app
# {"message": "Hello world!"}
```

Here is the [repo](https://github.com/AlvarDev/buildpacks-go-example) with all the code in this post, check the branchs.

Hope this post helps you!