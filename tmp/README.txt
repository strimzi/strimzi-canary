Some resources for experimenting with strimzi-canary

1. Preparation
Make sure that you have your strimzi-canary docker image. If not, build strimzi-canary and
make its image available to your cluster.

Assuming you already have installed strimzi and its monitoring stack [https://strimzi.io/docs/operators/latest/deploying.html#assembly-metrics-setup-str], and strimzi-canary (e.g., installed ts package/install/*.yaml files)

2. Install the pod-monitor for strimzi-canary by installing canary-monitor.yaml

Edit canary-monitor.yaml to adjust some parameters such as the namespace selector to match your kafka installation.

Execute the commands (assuming your kafka is installed at namespace my-kafka-project):

 kubectl -n my-kafka-project apply -f canary-monitor.yaml

3. Installing the grafana dashboard

Upload the grafana dashboard grafana-dashboards/strimzi-kafka-canary.json using Grafana UI.

See grafana-dashboard.png for a sample screenshot.
