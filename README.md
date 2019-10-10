# Cloud Native Chatbot

Enables the operator to setup a Slack bot that receives commands
execute it on specific containers and return to who requested the 
ouput value.


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
 
### Running locally

To run the controller locally, after settings up the commands:

```
export GO111MODULE=on
make run
```