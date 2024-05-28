package app

import "kdmid-queue-checker/app/daemon"

type Application struct {
	Daemon Daemon
}

type Daemon struct {
	CheckSlot *daemon.CheckSlot
	Bot       *daemon.NotifierBot
}
