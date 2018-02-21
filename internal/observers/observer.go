// Package observers contains the core logic for observers, which are what
// observe the environment to discover running services. An Observer only has
// to implement one method, Configure, which receives the configuration for
// that observer.  That Configure method might be called multiple times, so an
// observer should be prepared for its config to change and take appropriate
// steps.
//
// The ultimate output of an observer are objects that implement
// services.Endpoint, which represent the discovered service endpoints.
package observers

import (
	"github.com/pkg/errors"
	"github.com/signalfx/signalfx-agent/internal/core/config"
	"github.com/signalfx/signalfx-agent/internal/core/services"
	"github.com/signalfx/signalfx-agent/internal/utils"
	log "github.com/sirupsen/logrus"
)

// Shutdownable describes an observer that has a shutdown routine.  Observers
// should implement this if they spin up any goroutines that need to be
// stopped.
type Shutdownable interface {
	Shutdown()
}

// ObserverFactory creates an unconfigured instance of an observer
type ObserverFactory func(*ServiceCallbacks) interface{}

var observerFactories = map[string]ObserverFactory{}

// ConfigTemplates are blank (zero-value) instances of the configuration struct for a
// particular observer type.
var ConfigTemplates = map[string]interface{}{}

// Register an observer of _type with the agent.  configTemplate should be a
// zero-valued struct that is of the same type that is accepted by the
// Configure method of the observer.
func Register(_type string, factory ObserverFactory, configTemplate interface{}) {
	if _, ok := observerFactories[_type]; ok {
		log.WithFields(log.Fields{
			"observerType": _type,
		}).Error("Observer type already registered")
		return
	}
	observerFactories[_type] = factory
	ConfigTemplates[_type] = configTemplate
}

// ServiceCallbacks are a set of functions that the observers call upon
// detecting new services and disappearing services.
type ServiceCallbacks struct {
	Added   func(services.Endpoint)
	Removed func(services.Endpoint)
}

func configureObserver(observer interface{}, conf *config.ObserverConfig) error {
	log.WithFields(log.Fields{
		"config": *conf,
	}).Debug("Configuring observer")

	finalConfig := utils.CloneInterface(ConfigTemplates[conf.Type])

	if err := config.FillInConfigTemplate("ObserverConfig", finalConfig, conf); err != nil {
		return err
	}

	if err := config.ValidateCustomConfig(finalConfig); err != nil {
		return errors.Wrap(err, "Observer config is invalid")
	}

	return config.CallConfigure(observer, finalConfig)
}
