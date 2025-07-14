package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/niksmo/messaging/internal/messaging"
	"github.com/niksmo/messaging/pkg/logger"
)

type msgEmitter interface {
	Emit(key string, msg messaging.Message) error
}

type msgView interface {
	Get(key string) (any, error)
}

type mux interface {
	HandleFunc(pattern string,
		handler func(http.ResponseWriter, *http.Request))
}

type httpHandler struct {
	l logger.Logger
	e msgEmitter
	v msgView
}

func NewHandler(l logger.Logger, mux mux, e msgEmitter, v msgView) {
	h := &httpHandler{l, e, v}
	mux.HandleFunc("POST /{name}", h.sendHandler)
	mux.HandleFunc("GET /{name}", h.feedHandler)
}

func (h *httpHandler) sendHandler(w http.ResponseWriter, r *http.Request) {
	const op = "httpHandler.sendHandler"
	log := h.l.WithOp(op)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var m messaging.Message
	err = json.Unmarshal(data, &m)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	senderName := h.getNamePath(r)
	m.From = senderName
	err = h.e.Emit(senderName, m)
	if err != nil {
		log.Error().Err(err).Msg("failed to emit")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprintf(w, "sent message: %q\nto: %q\n", m.Content, m.To)
	if err != nil {
		log.Error().Err(err).Msg("failed to write response")
		return
	}
	log.Info().Str(
		"sentMsg", m.Content).Str("senderName", senderName).Send()
}

func (h *httpHandler) feedHandler(w http.ResponseWriter, r *http.Request) {
	const op = "httpHandler.feedHandler"
	log := h.l.WithOp(op)

	readerName := h.getNamePath(r)

	ml, err := h.v.Get(readerName)
	if err != nil {
		log.Error().Err(err).Msg("failed get data from view")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ml == nil {
		log.Info().Str("readerName", readerName).Msg("no content")
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintln(w, "no messages for you")
		return
	}

	mlt, ok := ml.([]messaging.Message)
	if !ok {
		log.Error().Type("msgListType", ml).Msg("unexpected type")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Messages:")
	for i, m := range mlt {
		n := i + 1
		fmt.Fprintf(w, "%d from: %q content: %q\n", n, m.From, m.Content)
	}
	log.Info().Int(
		"msgListSize", len(mlt)).Str("readerName", readerName).Send()
}

func (h *httpHandler) getNamePath(r *http.Request) string {
	return r.PathValue("name")
}
