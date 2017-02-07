Introduction
===

The system fetches a message from a Broker (just RabbitMQ is supported for now), run a Pod in the configurated format of the message, and send the result to another queue on the same broker.

Use Cases - See the Hubot folder for an example of how it can be used/integrated.

Message format
---

```
{"id": podName, "image": dockerImage, "args": ["a", "list", "of", "strings"]}
```

Parameters
---

If the project isn't running in a Kubernetes cluster, try to connect via local configuration on --kubeconfig.

```
--uri        - amqp://guest:guest@rabbitmq:5672 - AMQP default host
--kubeconfig - ~/.kube/config - KubeConfig path
--var        - var.yaml - file with Environemnt Variables to be rendered as secrets
```

Var.yaml
---

The format of var.yaml is:

```
- name: VARIABLE_NAME
  secret: secretObject
  key: keyObject
```

Ping
---

http://host:8080/ping


Building by hand
----

make docker


Helm Chart
---

helm install ./charts --name consumer
