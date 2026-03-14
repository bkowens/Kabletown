package handlers

import (
	"net/http"
)

type PluginHandler struct {}

func NewPluginHandler() *PluginHandler {
	return &PluginHandler{}
}

func (h *PluginHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
