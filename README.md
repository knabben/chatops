RUNNER
===

For now we have no logic implemented on our program, it will consume the queue
but it will probably complaint about this function:

```
createPod(kubeClient *client.Client, event Event, podName string) {
    pod := &api.Pod{
        ObjectMeta: api.ObjectMeta{
            Name:      podName,
            Namespace: ns,
            Labels:    map[string]string{"name": podName},
        },
        Spec: api.PodSpec{
            RestartPolicy: api.RestartPolicyOnFailure,
            Containers: []api.Container{
                {
                    Name:  podName,
                    Image: imgName,
                    Args:  args,
                    Env:   envVars,
                },
            },
        },
    }
    _, err := kubeClient.Pods(ns).Create(pod)
    ...
}
```

You can run the consume command as follows:

```
make compile
```

TODO
  * create from yaml file


HUBOT
===


Build Hubot Dockerfile, you can use a private registry to store it.

```
make hubot
```

To development and tests you can use the following command:

```
make hubot-run
```

Ensure that:

* You have a RabbitMQ container named rabbit
* You have the HUBOT_SLACK_TOKEN variable environment set.
