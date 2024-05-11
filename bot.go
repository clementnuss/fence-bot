package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

var (
	validUsers map[int64]string
	b          *tele.Bot
	m          *tele.ReplyMarkup
)

type User struct {
	FirstName string
	LastName  string
	Username  string
	UserID    int
}

func bot() {
	pref := tele.Settings{
		Token:       os.Getenv("TELEGRAM_TOKEN"),
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
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
	b, err = tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	statusButton := tele.InlineButton{Unique: "status", Text: "Statut"}
	on := tele.InlineButton{Unique: "on", Text: "On ⚡️"}
	off := tele.InlineButton{Unique: "off", Text: "Off"}

	m = b.NewMarkup()
	m.InlineKeyboard = append(m.InlineKeyboard, []tele.InlineButton{off, statusButton, on})

	b.Handle("/start", func(c tele.Context) error {
		return c.Send(fenceStatus(), m)
	})

	b.Handle(&statusButton, func(c tele.Context) error {
		mqttStatusUpdate()
		time.Sleep(200 * time.Millisecond)
		_, _ = b.Edit(c.Message(), fenceStatus(), m)
		return c.Respond(&tele.CallbackResponse{})
	})

	b.Handle(&on, func(c tele.Context) error {
		return commandSwitch(true, c)
	})

	b.Handle(&off, func(c tele.Context) error {
		return commandSwitch(false, c)
	})

	b.Start()
}

func commandSwitch(status bool, c tele.Context) error {
	attrs := chatToAttrs(c.Chat())
	if _, ok := validUsers[c.Chat().ID]; !ok {
		slog.Info("unauthenticated user", attrs...)
		return c.Respond(&tele.CallbackResponse{Text: "Utilisateur non autorisé"})
	}

	attrs = append(attrs, slog.Attr{Key: "desired_switch_status", Value: slog.StringValue(boolToStr(status))})
	mqttCommandSwitch(status)
	slog.Info("updated fence status", attrs...)

	time.Sleep(250 * time.Millisecond)
	_, _ = b.Edit(c.Message(), fenceStatus(), m)
	return c.Respond(&tele.CallbackResponse{})
}

func chatToAttrs(chat *tele.Chat) (attrs []any) {
	attrs = append(attrs, slog.Attr{Key: "first_name", Value: slog.StringValue(chat.FirstName)})
	attrs = append(attrs, slog.Attr{Key: "last_name", Value: slog.StringValue(chat.LastName)})
	attrs = append(attrs, slog.Attr{Key: "username", Value: slog.StringValue(chat.Username)})
	attrs = append(attrs, slog.Attr{Key: "user_id", Value: slog.Int64Value(chat.ID)})
	return
}
