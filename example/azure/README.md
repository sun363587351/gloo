1. Now for a quick demo showing how easy it is to migrate serverless functions
between AWS, Google, and Microsoft Azure.

2. Here I have a simple Alexa skill that's configured to send a request
to my Kubernetes cluster and read the response body aloud. Gloo is acting
as my kubernetes ingress, so Gloo will receive the requests and send them
to the configured backend.

3. Let's see what happens when I run the Alex Skill now.

<show>

Socket hangup - that's because we haven't created a route yet. Let's
start with a route that invokes a Lambda function.

4.
* First, I'll create an upstream object so to connect Gloo to my Amazon account.


```
cat <<EOF | glooctl upstream create -f -
name: my-aws
type: aws
spec:
  region: us-east-1
  secret_ref: aws-lambda-us-east-1
EOF
```


* After a second or two, Gloo will have discovered the availabe functions
on my account, and annotated my upstream object with them.


```
glooctl upstream get my-aws
```

* Let's create a route to my Lambda function and see that I can invoke it
with Alexa


```
glooctl route create \
    --path-exact /serverless-demo \
    --upstream my-aws \
    --function 'gloo-hello:$LATEST'
```


* Let's run the alexa skill again.

 [show]

Awesome! We just invoked a lambda function
with Alexa using Gloo.

5. Let's change our route so Alexa can invoke a Google Cloud Function

* First i'll need to connect Gloo to my Google Cloud account.

```
cat <<EOF | glooctl upstream create -f -
name: my-google
type: google
spec:
  region: us-east1
  project_id: k8s-cluster-144619
metadata:
  annotations:
    "gloo.solo.io/google_secret_ref": "gcf-myproject-secret"
EOF
```

* Functions will be discovered for Google as well.

```
glooctl upstream get my-google
```

* Now I'll update the route to point to google instead
```
glooctl route delete \
    --path-exact /serverless-demo \
    --upstream my-aws \
    --function 'gloo-hello:$LATEST'

glooctl route create \
    --path-exact /serverless-demo \
    --upstream my-google \
    --function projects/k8s-cluster-144619/locations/us-central1/functions/gloo-hello-working
```

* Let's see that Alexa can now call my Google Function.

<show>

Pretty cool.

6. One more time, but with Azure
* Connect Gloo toy my Azure account.
```

cat <<EOF | glooctl upstream create -f -
name: my-azure
type: azure
spec:
  function_app_name: functions-g7ffnq4
metadata:
  annotations:
    "gloo.solo.io/azure_publish_profile": "azure-funcs-secret"
EOF

```

* My functions will be discovered automatically here too.
```

glooctl upstream get my-azure

```

* Modify the route

```

glooctl route delete \
    --path-exact /serverless-demo \
    --upstream my-google \
    --function projects/k8s-cluster-144619/locations/us-central1/functions/gloo-hello-working

glooctl route create \
    --path-exact /serverless-demo \
    --upstream my-azure \
    --function gloo-hello

```

* Back to Alexa

<show>


With Gloo, the calling conventions and invocation APIs are abstracted from
clients, allowing seamless migration of serverless functions across clouds.



Cleanup

```
glooctl virtualhost delete default
glooctl upstream delete my-aws
glooctl upstream delete my-google
glooctl upstream delete my-azure
```


## lambda

glooctl secret create aws --name aws-lambda-us-east-1

cat <<EOF | glooctl upstream create -f -
name: my-aws
type: aws
spec:
  region: us-east-1
  secret_ref: aws-lambda-us-east-1
EOF


## gcf

kubectl create secret generic \
    gcf-myproject-secret \
    --from-literal \
    json_key_file="$(cat ~/Downloads/k8s-cluster-64b10ee30dd0.json)"

cat <<EOF | glooctl upstream create -f -
name: my-google
type: google
spec:
  region: us-east1
  project_id: k8s-cluster-144619
metadata:
  annotations:
    "gloo.solo.io/google_secret_ref": "gcf-myproject-secret"
EOF

## azure

kubectl create secret generic \
    azure-funcs-secret \
    --from-literal \
    publish_profile="$(cat ~/Downloads/functions-g7ffnq4.PublishSettings)"

cat <<EOF | glooctl upstream create -f -
name: my-azure
type: azure
spec:
  function_app_name: functions-g7ffnq4
metadata:
  annotations:
    "gloo.solo.io/azure_publish_profile": "azure-funcs-secret"
EOF

## virtualhost

glooctl route create \
    --path-exact /serverless-demo \
    --upstream my-aws \
    --function