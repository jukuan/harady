harady/
├── bin/
│   └── deploy.sh
├── backend/
│   ├── .env.example
│   ├── go.mod
│   ├── cmd/
│   │   ├── server/main.go
│   │   └── cli/main.go
│   ├── data/           # SQLite db (gitignored)
│   └── internal/
│       ├── config/config.go
│       ├── db/db.go
│       ├── db/migrations.go
│       ├── models/models.go
│       ├── store/city_store.go
│       ├── game/events.go
│       ├── game/player.go
│       ├── game/room.go
│       ├── game/hub.go
│       ├── game/bot.go
│       ├── ws/client.go
│       └── httpapi/router.go
├── frontend/           # next step
├── .gitignore
└── README.md
