Introduction
----
The idea here is to create a pipeline of communication between a Hubot (A Node.JS Robot) and a container consumer of these messages.

The container will be responsible to start and stop Dockers containers on the machine after receiving the events from the queue.


Install hubot Plugin
--------------------

Put the hubot/scripts/sender.coffee on your plugin folder, add amqplib as a package.json dependency.

Create a RabbitMQ channel called "runner" with auto_delete=True and another called tasks to receive the response from containers.
