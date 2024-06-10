package app

import (
	"kdmid-queue-checker/app/daemon"
	"kdmid-queue-checker/app/query"
)

type Application struct {
	Daemon Daemon
	Query  Query
}

type Daemon struct {
	CheckSlot *daemon.CheckSlot
	Bot       *daemon.NotifierBot
}

type Query struct {
	ListUsers *query.ListUsersHandler
}
