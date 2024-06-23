package ports

import (
	"net/http"
	
	"github.com/go-chi/render"
	
	"github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/common/auth"
	"github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/common/server/httperr"
	"github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/app"
	"github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/app/command"
)

type HttpServer struct {
	app app.Application
}

func NewHttpServer(application app.Application) HttpServer {
	return HttpServer{
		app: application,
	}
}

func (h HttpServer) MakeHourUnavailable(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}
	
	if user.Role != "trainer" {
		httperr.Unauthorised("invalid-role", nil, w, r)
		return
	}
	
	hourUpdate := &HourUpdate{}
	if err := render.Decode(r, hourUpdate); err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}
	
	err = h.app.Commands.MakeHoursUnavailable.Handle(r.Context(), command.MakeHoursUnavailable{Hours: hourUpdate.Hours})
	if err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
