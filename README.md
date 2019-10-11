# Cloud Native Chatbot

Enables the operator to setup a Slack bot that receives commands,
brings a job to life and return the output back to the user who 
requested it.

### Installation

Using your cluster credentials you must use the following command
to install the correct CRD:

```
make install

# Chat custom resource definition

apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: chats.chat.ops.com
spec:
  group: chat.ops.com
  names:
    kind: Chat
    listKind: ChatList
    plural: chats
    singular: chat
  scope: Namespaced
```

### Install CR samples

To enable the command the user must install the correct Chat CRs,
it is possible to use more than one.

```
$ cat config/samples/chat_v1_chat.yaml
apiVersion: chat.ops.com/v1
kind: Chat
metadata:
  name: chat-sample
spec:
  command: "curl https://report1"
  jobImage: private/image
  timestamp: 0
status:
  username: ""
  command: ""
  channel: ""
  timestamp: ""
```

*TODO - ConfigMap to ACL users vs. CRD and commands*

The main fields here are:

- Spec.command - The binary path used in the job
- Spec.jobImage - The image URL used in the job
- Spec.userACL - A list of Slack users who are

## Development

### Configure Slack

Create a new file ./env.yaml with your Slack token:

```
CTRL_SLACK_TOKEN: xoxb-random-689172615444-hhhhNhhTO5hhhhItyyy4yyf9
```

### Running locally

To run the controller locally, after settings up the commands:

```
export GO111MODULE=on
make run
```