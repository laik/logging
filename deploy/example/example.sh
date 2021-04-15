cat << "EOF" | kubectl apply -f -
---
apiVersion: logging.yamecloud.io/v1
kind: Slack
metadata:
  name: kube-system-logging-slack
  namespace: kube-system
spec:
  selector: app=123
EOF


cat << "EOF" | kubectl apply -f -
---
apiVersion: logging.yamecloud.io/v1
kind: Sink
metadata:
  name: kube-system-logging-sink
  namespace: kube-system
spec:
  type: kafka
  address: 10.200.100.200:9092
  partition: 3
EOF