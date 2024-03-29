package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
)

func bot() {
	pref := tele.Settings{
		Token:       os.Getenv("TELEGRAM_TOKEN"),
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
	}

	validUsers := make(map[int64]string)
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

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	statusButton := tele.InlineButton{Unique: "status", Text: "Statut"}
	on := tele.InlineButton{Unique: "on", Text: "On ⚡️"}
	off := tele.InlineButton{Unique: "off", Text: "Off"}

	m := b.NewMarkup()
	m.InlineKeyboard = append(m.InlineKeyboard, []tele.InlineButton{off, statusButton, on})

	b.Handle("/start", func(c tele.Context) error {
		return c.Send(fenceStatus(), m)
	})

	b.Handle(&statusButton, func(c tele.Context) error {
		statusUpdate()
		time.Sleep(200 * time.Millisecond)
		_, _ = b.Edit(c.Message(), fenceStatus(), m)
		return c.Respond(&tele.CallbackResponse{})
	})

	b.Handle(&on, func(c tele.Context) error {
		attrs := chatToAttrs(c.Chat())
		if _, ok := validUsers[c.Chat().ID]; ok {
			toggleOn()
			slog.Info("fence toggled on", attrs...)
		} else {
			slog.Info("unauthenticated user", attrs...)
			return c.Respond(&tele.CallbackResponse{Text: "Utilisateur non autorisé"})
		}
		time.Sleep(250 * time.Millisecond)
		_, _ = b.Edit(c.Message(), fenceStatus(), m)
		return c.Respond(&tele.CallbackResponse{})
	})

	b.Handle(&off, func(c tele.Context) error {
		attrs := chatToAttrs(c.Chat())
		if _, ok := validUsers[c.Chat().ID]; ok {
			slog.Info("fence toggled off", attrs...)
			toggleOff()
		} else {
			slog.Info("unauthenticated user", attrs...)
			return c.Respond(&tele.CallbackResponse{Text: "Utilisateur non autorisé"})
		}
		time.Sleep(250 * time.Millisecond)
		_, _ = b.Edit(c.Message(), fenceStatus(), m)
		return c.Respond(&tele.CallbackResponse{})
	})

	b.Start()
}

type User struct {
	FirstName string
	LastName  string
	Username  string
	UserID    int
}

func chatToAttrs(chat *tele.Chat) (attrs []any) {
	attrs = append(attrs, slog.Attr{Key: "first_name", Value: slog.StringValue(chat.FirstName)})
	attrs = append(attrs, slog.Attr{Key: "last_name", Value: slog.StringValue(chat.LastName)})
	attrs = append(attrs, slog.Attr{Key: "username", Value: slog.StringValue(chat.Username)})
	attrs = append(attrs, slog.Attr{Key: "ID", Value: slog.Int64Value(chat.ID)})
	return
}

func fenceStatus() (status string) {
	d := time.Since(stat.SwitchStatus.LastUpdate)
	if d > 24*time.Hour {
		status = "⚠️ Relais déconnecté depuis plus d'un jour\n"
	} else if d > 60*time.Second {
		status = fmt.Sprintf("⚠️ Relais connecté pour la dernière fois il y a %v\n", d)
	} else {
		status = fmt.Sprintf(`
	Électricité dans la clôture: %s
	Interrupteur rotatif: %s
	Tension du réseau: %.2f V
	Puissance moyenne: %.2f W
	Dernière données du relais il y a %v`,
			boolToEmoji(stat.SwitchStatus.Output),
			boolToEmoji(stat.InputStatus.State),
			stat.SwitchStatus.Voltage,
			stat.SwitchStatus.AveragePower,
			time.Duration.Round(d, time.Millisecond),
		)
	}

	return
}

func boolToEmoji(b bool) string {
	if b {
		return "✅"
	} else {
		return "❌"
	}

}
