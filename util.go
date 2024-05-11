package main

import (
	"fmt"
	"time"
)

func boolToEmoji(b bool) string {
	if b {
		return "✅"
	} else {
		return "❌"
	}

}

func boolToStr(status bool) string {
	if status {
		return "on"
	} else {
		return "off"
	}
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
	Dernières données du relais il y a %v`,
			boolToEmoji(stat.SwitchStatus.Output),
			boolToEmoji(stat.InputStatus.State),
			stat.SwitchStatus.Voltage,
			stat.SwitchStatus.AveragePower,
			time.Duration.Round(d, time.Millisecond),
		)
	}

	return
}
