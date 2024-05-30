package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	telebotv3 "gopkg.in/telebot.v3"
)

var (
	validUsers map[int64]string
	b          *telebotv3.Bot
	m          *telebotv3.ReplyMarkup
)

type User struct {
	FirstName string
	LastName  string
	Username  string
	UserID    int
}

func bot() {
	settings := telebotv3.Settings{
		Token:       os.Getenv("TELEGRAM_TOKEN"),
		Poller:      &telebotv3.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
	}

	validUsers = make(map[int64]string)
	{
		for _, str := range strings.Split(os.Getenv("VALID_USERS_LIST"), ",") {
			spl := strings.Split(str, ":")
			id, err := strconv.ParseInt(spl[0], 10, 64)
			if err != nil {
				log.Fatal("Couldn't parse user id, exiting\n", err)
				return
			}
			name := "UNKNOWN"
			if len(spl) > 0 {
				name = spl[1]
			}
			validUsers[id] = name
		}
	}

	var err error
	b, err = telebotv3.NewBot(settings)
	if err != nil {
		log.Fatal(err)
		return
	}

	statusButton := telebotv3.InlineButton{Unique: "status", Text: "Statut"}
	on := telebotv3.InlineButton{Unique: "on", Text: "On ⚡️"}
	off := telebotv3.InlineButton{Unique: "off", Text: "Off"}

	m = b.NewMarkup()
	m.InlineKeyboard = append(m.InlineKeyboard,
		[]telebotv3.InlineButton{off, statusButton, on})

	b.Handle("/start", func(c telebotv3.Context) error {
		return c.Send(fenceStatus(), m)
	})

	b.Handle(&statusButton, func(c telebotv3.Context) error {
		mqttStatusUpdate()
		time.Sleep(200 * time.Millisecond)
		_, _ = b.Edit(c.Message(), fenceStatus(), m)
		return c.Respond(&telebotv3.CallbackResponse{})
	})

	b.Handle(&on, func(c telebotv3.Context) error {
		return commandSwitch(true, c)
	})

	b.Handle(&off, func(c telebotv3.Context) error {
		return commandSwitch(false, c)
	})

	b.Start()
}

func commandSwitch(status bool, c telebotv3.Context) error {
	attrs := chatToAttrs(c.Chat())
	if _, ok := validUsers[c.Chat().ID]; !ok {
		slog.Info("unauthenticated user", attrs...)
		return c.Respond(&telebotv3.CallbackResponse{Text: "Utilisateur non autorisé"})
	}

	attrs = append(attrs, slog.Attr{Key: "desired_switch_status", Value: slog.StringValue(boolToStr(status))})
	mqttCommandSwitch(status)
	slog.Info("updated fence status", attrs...)

	time.Sleep(250 * time.Millisecond)
	_, _ = b.Edit(c.Message(), fenceStatus(), m)
	return c.Respond(&telebotv3.CallbackResponse{})
}

func chatToAttrs(chat *telebotv3.Chat) (attrs []any) {
	attrs = append(attrs, slog.Attr{Key: "first_name", Value: slog.StringValue(chat.FirstName)})
	attrs = append(attrs, slog.Attr{Key: "last_name", Value: slog.StringValue(chat.LastName)})
	attrs = append(attrs, slog.Attr{Key: "username", Value: slog.StringValue(chat.Username)})
	attrs = append(attrs, slog.Attr{Key: "user_id", Value: slog.Int64Value(chat.ID)})
	return
}
