FROM node:6-slim

WORKDIR /app

COPY . /app

RUN apt-get update
RUN apt-get install git -y

RUN npm install
RUN npm install coffee-script -g

CMD ["bin/hubot", "--adapter", "slack"]
