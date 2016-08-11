Introduction
----
The idea here is to create a pipeline of communication between Hubot and a Docker host.

The code will be responsible to create new Docker containers on the machine after receiving the from a chat.

You will need a RabbitMQ broker and a user with queue creation permission.

Install Hubot Plugin
--------------------

```
export RABBITMQ_URL="amqp://user:pass@xxx.xxx.xxx.xxx/"
```

Put the hubot/scripts/sender.coffee on your plugin folder and add amqplib as a package.json dependency.

Running the consumer
--------------------

To run the consumer you can build it, and run passing the AMQP host and the Docker that will run the commands:

```
go build runner.go
./runner --uri "amqp://user:pass@xxx.xxx.xxx.xxx" --endpoint "tcp://docker_host:2374"
```
