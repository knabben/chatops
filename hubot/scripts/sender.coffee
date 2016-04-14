module.exports = (robot) ->

  require('amqplib/callback_api').connect process.env.RABBITMQ_URL, (err, conn) ->

    start_queue = (res, message) ->
      on_open = (err, ch) ->
        # Create queue and send command
        ch.assertQueue 'runner', {autoDelete: true}
        ch.sendToQueue 'runner', new Buffer(JSON.stringify(message))

        # Consume messages as response and post to channel
        ch.consume 'tasks', (msg) ->
          res.send(msg.content.toString())
          ch.ack(msg)
        conn.createChannel on_open

      robot.hear /script (.*) (.*)?/i, (res) ->
        script = res.match[1]
        params = res.match[2]

        start_queue res, {'script': script, 'params': params}
        res.send "Request sent"
