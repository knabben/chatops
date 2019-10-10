# Cloud Native Chatbot

Enables the operator to setup a Slack bot that receives commands,
execute it on a specific container, and return the output value to who 
requested it.

**NOTE: DON'T USE THIS IN PRODUCTION** 

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
  command: report
  podlabel: server1
  timestamp: 0
status:
  command: ""
  arguments: ""
  channel: ""
  timestamp: ""
  username: ""
```
 
The main fields here are:
- Spec.Command - The binary on PATH
- Spec.Podlabel - Used for search the Pod 

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

## Production

Enabling CRDs mgmt and pod/exec is really not a good RBAC permission 
for a regular user since you can't create granularity for what can be 
executed from the CRD creation, even if you ACL things in this app
other people using the account can ignore it.
