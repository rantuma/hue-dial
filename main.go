package main

import (
	"flag"

	"github.com/rantuma/hue-dial/application/event"
	"github.com/rantuma/hue-dial/infrastructure/config"
	"github.com/rantuma/hue-dial/infrastructure/hue"
	"github.com/rantuma/hue-dial/infrastructure/logging"
	"github.com/rantuma/hue-dial/infrastructure/setup"
	pkglogger "github.com/rantuma/hue-dial/pkg/logger"
)

func main() {
	forceSetup := flag.Bool("setup", false, "force the setup wizard to run")
	flag.Parse()

	log := logging.New(pkglogger.Debug)

	store, err := config.New()
	if err != nil {
		log.Panicf("failed to create config store: %s", err.Error())
	}

	if *forceSetup || !store.Exists() {
		_, err = setup.Wizard(log, store)
		if err != nil {
			log.Panicf("setup wizard failed: %s", err.Error())
		}
	}

	setupCfg, err := store.Load()
	if err != nil {
		log.Panicf("failed to load configuration: %s", err.Error())
	}

	cfg := setupCfg.ToConfig()

	hueAdapter, err := hue.New(cfg.Hue.Key, cfg.Hue.BridgeIP)
	if err != nil {
		log.Panicf("failed to create Hue adapter: %s", err.Error())
	}

	eventService := event.New(log, hueAdapter, hueAdapter, cfg)

	err = eventService.Start()
	if err != nil {
		log.Panicf("failed to start event service: %s", err.Error())
	}
}
