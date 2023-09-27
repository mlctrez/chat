package compo

import (
	"bufio"
	"fmt"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	natsws "github.com/mlctrez/goapp-natsws"
	"github.com/nats-io/nats.go"
	"net/http"
	"strings"
	"time"
)

var _ app.Mounter = (*Chat)(nil)
var _ app.AppUpdater = (*Chat)(nil)

type Chat struct {
	app.Compo
	messages []string
	conn     natsws.Connection
	who      string
}

func (d *Chat) OnAppUpdate(ctx app.Context) {
	if ctx.AppUpdateAvailable() {
		ctx.Reload()
	}
}

const Messages = "chatMessages"

func (d *Chat) OnMount(ctx app.Context) {

	errSession := ctx.LocalStorage().Get("who", &d.who)
	if d.who == "" {
		d.who = "anon"
	}
	if errSession != nil {
		fmt.Println("errSession", errSession)
	}

	app.Window().GetElementByID("who").Set("innerHTML", d.who)

	if resp, err := http.Get("/last"); err == nil {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			d.messages = append(d.messages, scanner.Text())
		}
		d.Update()
	}

	natsws.Observe(ctx, &d.conn).OnChange(func() {

		reason := d.conn.ChangeReason()
		switch reason {
		case natsws.Connect:
			if err := d.conn.Subscribe(Messages, d.onMessage); err != nil {
				d.messages = append(d.messages, err.Error())
			}
		}
		d.changeButton(reason)

		d.Update()
	})
}

func (d *Chat) changeButton(reason natsws.ChangeReason) {
	var cls = "btn-warning"
	switch reason {
	case natsws.Connect, natsws.Reconnect:
		cls = "btn-success"
	case natsws.Disconnect:
		cls = "btn-danger"
	}
	newClass := fmt.Sprintf("btn %s dropdown-toggle", cls)
	app.Window().GetElementByID("who").Call("setAttribute", "class", newClass)

}
func (d *Chat) onMessage(msg *nats.Msg) {
	d.messages = append(d.messages, string(msg.Data))
	d.Update()
}

func (d *Chat) sendMessage(ctx app.Context, msg string) {
	if d.who == "" {
		d.who = "anon"
	}

	ts := time.Now().Format("Mon " + time.Kitchen)
	chatMessage := fmt.Sprintf("%s %s : %s", ts, d.who, msg)
	err := d.conn.Publish(Messages, []byte(chatMessage))
	if err != nil {
		app.Log("publish error", err)
	}
}

func (d *Chat) Render() app.UI {
	var reversed []string

	reversed = append(reversed, d.messages...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}

	users := []string{"Luke", "Leia", "Mom", "Dad"}

	return app.Div().Class("container-fluid").Style("padding-top", "1em").Body(
		&natsws.Component{},
		app.Button().ID("who").Class("btn btn-secondary dropdown-toggle").
			Type("button").DataSet("bs-toggle", "dropdown").
			Aria("expanded", "false"),
		app.Ul().Class("dropdown-menu").Body(
			app.Range(users).Slice(func(i int) app.UI {
				return app.Li().Body(
					app.A().Class("dropdown-item").Text(users[i]).OnClick(func(ctx app.Context, e app.Event) {
						d.who = users[i]
						app.Window().GetElementByID("who").Set("innerHTML", d.who)
						ctx.LocalStorage().Set("who", d.who)
					}),
				)
			}),
		),
		app.Button().Type("button").Class("btn btn-primary").Text("Leaving").OnClick(func(ctx app.Context, e app.Event) {
			d.sendMessage(ctx, "Leaving")
		}),
		app.Button().Type("button").Class("btn btn-primary").Text("Arrived").OnClick(func(ctx app.Context, e app.Event) {
			d.sendMessage(ctx, "Arrived")
		}),
		//app.Hr(),
		app.Input().ID("message").Class("form-control").Size(20).OnKeyPress(func(ctx app.Context, e app.Event) {
			if strings.ToLower(e.Get("key").String()) == "enter" {
				inputField := app.Window().GetElementByID("message")
				msg := inputField.Get("value").String()
				inputField.Set("value", nil)
				d.sendMessage(ctx, msg)
			}
		}),
		app.Ul().Class("list-group").Body(app.Range(reversed).Slice(func(i int) app.UI {
			return app.Li().Class("list-group-item").Text(reversed[i])
		})),
	)
}
