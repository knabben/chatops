/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChatSpec defines the desired state of Chat
type ChatSpec struct {
	JobImage  string `json:"jobImage"`
	Command   string `json:"command"`
	Timestamp int64  `json:"timestamp"`
}

// ChatStatus defines the observed state of Chat
type ChatStatus struct {
	Command   string `json:"command"`
	Username  string `json:"username"`
	Timestamp string `json:"timestamp"`
	Channel   string `json:"channel"`
}

// +kubebuilder:object:root=true

// Chat is the Schema for the chats API
type Chat struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChatSpec   `json:"spec,omitempty"`
	Status ChatStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChatList contains a list of Chat
type ChatList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Chat `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Chat{}, &ChatList{})
}
