Some resources for experimenting with strimzi-canary

1. Preparation
Make sure that you have your strimzi-canary docker image. If not, build strimzi-canary and
make its image available to your cluster.

2. Installing strimzi-canary and and its pod monitor to your cluster
(Skip this step if you have already installed strizmi-canary and its metrics data is available in your prometheus.)

This is a very simple deployment that installs one instance of strimzi-canary for experimentation.

Edit canary.yaml and canary-monitor.yaml to adjust some parameters such as the image name, KAFKA_BOOTSTRAP_SERVERS,
the namespace selector to match your installation.

Execute the commands:

 kubectl apply -f canary.yaml
 kubectl apply -f canary-monitor.yaml

3. Installing the grafana dashboard

Upload the grafana dashboard grafana-dashboards/strimzi-kafka-canary.json using Grafana UI.

See grafana-dashboard.png for a sample screenshot.
