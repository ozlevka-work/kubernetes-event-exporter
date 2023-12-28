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
	log.Debug().Str("event", ev.Message).Msg("Processing event Before drop loop")
	for _, v := range r.Drop {
		if v.MatchesEvent(ev) {
			return
		}
	}
	log.Debug().Str("event", ev.Message).Msg("Processing event Before match loop")

	// It has match rules, it should go to the matchers
	matchesAll := true
	for _, rule := range r.Match {
		if rule.MatchesEvent(ev) {
			if rule.Receiver != "" {
				registry.SendEvent(rule.Receiver, ev)
				// Send the event down the hole
			}
		} else {
			matchesAll = false
		}
	}

	log.Debug().Str("event", ev.Message).Msg("Processing event After match loop")

	// If all matches are satisfied, we can send them down to the rabbit hole
	if matchesAll {
		for _, subRoute := range r.Routes {
			subRoute.ProcessEvent(ev, registry)
		}
	}
}
