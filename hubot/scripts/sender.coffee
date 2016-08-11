# Description:
#   Send data to execution queue
#
# Configuration:
#   RABBITMQ_URL
#
# Commands:
#   hubot script
#
# Author:
#   knabben

module.exports = (robot) ->
  amqp = process.env.RABBITMQ_URL

  require('amqplib/callback_api').connect amqp, (err, conn) ->

    start_queue = (res, message) ->
      on_open = (err, ch) ->
        # Create queue and send command
        ch.assertQueue 'runner'
        ch.sendToQueue 'runner', new Buffer(JSON.stringify(message))

        # Consume messages as response and post to channel
        ch.assertQueue 'response'
        ch.consume 'response', (msg) ->
          res.send(msg.content.toString())
          ch.ack(msg)

      conn.createChannel on_open

    robot.hear /script (.*) (.*)?/i, (res) ->
      command = res.match[1]
      params = res.match[2]

      start_queue res, {'command': command, 'params': params}
      res.send "Request sent"
