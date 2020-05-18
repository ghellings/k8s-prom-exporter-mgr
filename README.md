k8s-prom-exporter-mgr
=====================

This tool is for automatically managing prometheus exporters in kubernetes based on scraping data from APIs.  As an example, it will create exporters in kubernetes for a dynamic number of nodes in AWS based on Tags.

Usage
=====
# CMD Options
```
-version                        ; Prints version
-sleeptime=60                   ; Time between loops in seconds, defaults to 1000 
-configfile="path_to_config"    ; Defaults to /etc/k8s-prom-exporter-mgr/config
-once                           ; Don't loop, just run once
-loglevel                       ; Level of logoutput trace,debug,info,warn,error

```
# Configfile
```
k8sdeploytemplatespath: "/etc/k8s-prom-exporter-mgr/k8stemplates/" # Path to prometheus k8s templates
k8snamespace: default                                              # K8s namespace to launch exporters in
services:                                                           
  prom-exporter-apache:                                            # Name for k8s deployments
    srvtype: Ec2                                                   # Service to scrape
    srv:
      tags:                                                        # Array of tags to search service
      - tag: "Product"
        value: "API"
k8slabels:
  managed: prom-exporter-mgr                                       # K8s tag to identify deployments
```

# K8s Templates

* The template must live in the templates directory and be named the same name at the service with a '.yml' extension.
* It must also have labels matching the k8slabels from the configfile
* It must also have two ARGs with the second ARG matching the following regexp ```https?://([^:/]+)(?::|/).*``` similar to http://REPLACEME:8080/server-status?auto 

```
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: apache-exporter
  labels:
    managed: "prom-exporter-mgr"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: apache-exporter
      managed: prom-exporter-mgr
  minReadySeconds: 10
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  template:
    metadata:
      labels:
        app: apache-exporter
        managed: prom-exporter-mgr
      annotations:
        sumologic.com/format: "json_merge"
        sumologic.com/sourceCategory: "apache-exporter"
        sumologic.com/kubernetes_meta: "false"
        sumologic.com/exclude: "true"
    spec:
      containers:
      - name: prom-apache-exporter
        image: ghellings/apache_exporter:latest
        command: [
          "/apache_exporter/apache_exporter"
        ]
        args: [
          "-scrape_uri",
          "http://10.0.2.85:8080/server-status?auto"
        ]
        env:
          - name: KUBE_REDEPLOY
            value: "0"
        resources:          
          limits:
            memory: "20Mi"
            cpu: "50m"
          requests:
            memory: "5Mi"
            cpu: "1m"
        imagePullPolicy: Always
        ports:
        - containerPort: 9117
```

Testing on a Desktop
====================

Install Docker for Desktop and enable Kubernetes. Export vars for AWS keys. Checkout this repo, edit the configs in example-configs. From checkout directory run
```
docker build . -t k8s-prom-exporter-mgr
```
Follwed by ( make sure your kubectl context is docker-desktop )
```
kubectl run -i --tty k8s-prom-exporter-mgr --image=k8s-prom-exporter-mgr --restart=Never -n default --image-pull-policy=“Never” --env=AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID --env=AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY --env=AWS_DEFAULT_REGION=$AWS_DEFAULT_REGION -- bash 
```

Then execute ```k8s-prom-exporter-mgr```    

License and Author
==================

* Author:: Greg Hellings (<greg@thesub.net>)


Copyright 2020, Searchspring, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
 