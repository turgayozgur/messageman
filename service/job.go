package service

import (
	"github.com/valyala/fasthttp"
)

// Queue push message to workers.
func (s *Server) Queue(ctx *fasthttp.RequestCtx) {
	queueName := string(ctx.QueryArgs().Peek("name"))

	if queueName == "" {
		s.badRequest(ctx, "\"name\" parameter required.")
		return
	}

	body := ctx.PostBody()
	if body == nil || len(body) == 0 {
		s.badRequest(ctx, "The request body required.")
		return
	}

	var mainAPI string
	if s.mainAPI != "" {
		mainAPI = s.mainAPI
	} else {
		// We can get the main api name from header if the x-service-name header provided.
		mainAPI = string(ctx.Request.Header.Peek("x-service-name"))
	}

	err := s.messager.Send(mainAPI, queueName, body)
	if err != nil {
		s.error(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	s.write(ctx, fasthttp.StatusOK, &ResponseModel{Message: "Successfully queued."})
}
