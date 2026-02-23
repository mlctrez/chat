package compo

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	natsws "github.com/mlctrez/goapp-natsws"
	"github.com/nats-io/nats.go"
)

var _ app.Mounter = (*Chat)(nil)
var _ app.AppUpdater = (*Chat)(nil)

type Chat struct {
	app.Compo
	messages []string
	conn     natsws.Connection
	who      string
	flashes  int
}

func (d *Chat) OnAppUpdate(ctx app.Context) {
	if ctx.AppUpdateAvailable() {
		ctx.Reload()
	}
}

const Messages = "chatMessages"
const btnMargin = "10px"

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
	if string(msg.Data) == "CLEAR_MESSAGES" {
		d.messages = []string{}
		d.flashes = 0
		d.Update()
		return
	}
	d.messages = append(d.messages, string(msg.Data))
	d.Update()

	d.flashes += 5
	if d.flashes > 5 {
		// already flashing
		return
	}

	go func() {
		for {
			if d.flashes <= 0 {
				break
			}
			app.Window().Call("changeBackground", "red")
			time.Sleep(500 * time.Millisecond)
			app.Window().Call("changeBackground", "")
			time.Sleep(500 * time.Millisecond)
			d.flashes--
		}
	}()
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

func (d *Chat) actionButton(text string, msg string) app.UI {
	return app.Button().Type("button").Class("btn btn-primary").
		Style("margin-right", btnMargin).
		Style("margin-bottom", "5px").
		Text(text).OnClick(func(ctx app.Context, e app.Event) {
		d.sendMessage(ctx, msg)
	})
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
			Style("margin-right", btnMargin).
			Style("margin-bottom", "5px").
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
		d.actionButton("Leaving", "Leaving"),
		d.actionButton("To Left", "Arrived : lot to left"),
		d.actionButton("To Right", "Arrived : lot to light"),
		//app.Hr(),
		app.Input().ID("message").Class("form-control").Size(20).
			Style("margin-bottom", "5px").
			OnKeyPress(func(ctx app.Context, e app.Event) {
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
		app.Div().Style("margin-top", "50px").Body(
			app.Button().Class("btn btn-warning").Text("clear").OnClick(func(ctx app.Context, e app.Event) {
				_ = d.conn.Publish(Messages, []byte("CLEAR_MESSAGES"))
			}),
		),
	)
}
