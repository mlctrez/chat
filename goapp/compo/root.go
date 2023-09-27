package compo

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/mlctrez/chat/goapp/compo/ui"
	"github.com/mlctrez/goapp-natsws"
)

var _ app.AppUpdater = (*Root)(nil)

type Root struct {
	app.Compo
}

func (r *Root) Render() app.UI {
	return app.Div().Body(
		&natsws.Component{},
		&ui.Chat{},
	)
}

func (r *Root) OnAppUpdate(ctx app.Context) {
	if ctx.AppUpdateAvailable() {
		ctx.Reload()
	}
}
