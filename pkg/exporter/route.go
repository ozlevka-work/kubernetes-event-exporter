package exporter

import (
	"github.com/resmoio/kubernetes-event-exporter/pkg/kube"
	"github.com/rs/zerolog/log"
)

// Route allows using rules to drop events or match events to specific receivers.
// It also allows using routes recursively for complex route building to fit
// most of the needs
type Route struct {
	Drop   []Rule
	Match  []Rule
	Routes []Route
}

func (r *Route) ProcessEvent(ev *kube.EnhancedEvent, registry ReceiverRegistry) {
	// First determine whether we will drop the event: If any of the drop is matched, we break the loop
	for _, v := range r.Drop {
		if v.MatchesEvent(ev) {
			return
		}
	}

	// It has match rules, it should go to the matchers
	matchesAll := true
	for _, rule := range r.Match {
		if rule.MatchesEvent(ev) {
			log.Debug().Str("event", ev.Message).Msg("Event matches rule")
			if rule.Receiver != "" {
				log.Debug().Str("event", ev.Message).Str("receiver", rule.Receiver).Msg("Sending event to receiver")
				registry.SendEvent(rule.Receiver, ev)
				// Send the event down the hole
			}
		} else {
			matchesAll = false
		}
	}

	// If all matches are satisfied, we can send them down to the rabbit hole
	if matchesAll {
		for _, subRoute := range r.Routes {
			subRoute.ProcessEvent(ev, registry)
		}
	}
}
