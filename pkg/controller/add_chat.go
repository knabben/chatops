package controller

import (
	"github.com/knabben/chatops/pkg/controller/chat"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, chat.Add)
}
