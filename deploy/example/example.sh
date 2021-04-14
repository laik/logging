cat << "EOF" | kubectl apply -f -
---
apiVersion: logging.yamecloud.io/v1
kind: Slack
metadata:
  name: kube-system-logging-slack
  namespace: kube-system
spec:
  add_tasks:
    - filter:
        expr: '[INFO]'
        max_length: "102"
      ns: kube-system
      pods:
        - container: echoer-api-86c648d678-z2p9p
          ips:
            - 127.0.0.1
          node: node1
          offset: 0
          pod: echoer-api-86c648d678-z2p9p
      service_name: echoer-api
  delete_tasks: []
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