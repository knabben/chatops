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
