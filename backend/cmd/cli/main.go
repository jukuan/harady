package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jukuan/harady/backend/internal/config"
	"github.com/jukuan/harady/backend/internal/db"
	"github.com/jukuan/harady/backend/internal/models"
	"github.com/jukuan/harady/backend/internal/store"
	"github.com/jukuan/harady/backend/internal/util"
)

func main() {
	cfg := config.Load()
	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		fatal(err)
	}
	defer conn.Close()
	cs := store.NewCityStore(conn)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		list(cs)
	case "show":
		show(cs, os.Args[2:])
	case "add":
		add(cs, os.Args[2:])
	case "update":
		update(cs, os.Args[2:])
	case "delete", "rm":
		del(cs, os.Args[2:])
	case "seed":
		seed(cs)
	case "reseed":
		reseed(cs)
	case "help", "-h", "--help":
		usage()
	default:
		fatal(fmt.Errorf("unknown command: %s", os.Args[1]))
	}
}

func usage() {
	fmt.Print(`harady-cli — manage the cities DB

USAGE:
  harady-cli list
  harady-cli show <id|name>
  harady-cli add    --name "Мінск" [--region "Мінская вобласць"] [--clues "сталіца;Няміга"]
  harady-cli update --id 1 [--name ...] [--region ...] [--clues ...]
  harady-cli delete <id>
  harady-cli seed         # add missing seed cities (idempotent)
  harady-cli reseed       # wipe the table and re-seed from the pack
`)
}

func list(cs *store.CityStore) {
	cities, err := cs.List()
	if err != nil {
		fatal(err)
	}
	if len(cities) == 0 {
		fmt.Println("(no cities)")
		return
	}
	fmt.Printf("%-4s  %-25s  %-25s\n", "ID", "NAME", "REGION")
	for _, c := range cities {
		fmt.Printf("%-4d  %-25s  %-25s\n", c.ID, c.Name, c.Region)
	}
}

func show(cs *store.CityStore, args []string) {
	if len(args) == 0 {
		fatal(errors.New("show requires <id|name>"))
	}
	var (
		c   *models.City
		err error
	)
	if id, e := strconv.ParseInt(args[0], 10, 64); e == nil {
		c, err = cs.GetByID(id)
	} else {
		c, err = cs.GetByName(args[0])
	}
	if err != nil {
		fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(c)
}

func add(cs *store.CityStore, args []string) {
	f := parseFlags(args)
	c := &models.City{
		Name:   f["name"],
		Region: f["region"],
		Clues:  util.SplitList(f["clues"]),
	}
	if c.Name == "" {
		fatal(errors.New("--name is required"))
	}
	id, err := cs.Create(c)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("created id=%d\n", id)
}

func update(cs *store.CityStore, args []string) {
	f := parseFlags(args)
	id, err := strconv.ParseInt(f["id"], 10, 64)
	if err != nil {
		fatal(errors.New("--id is required"))
	}
	c, err := cs.GetByID(id)
	if err != nil {
		fatal(err)
	}
	if v := f["name"]; v != "" {
		c.Name = v
	}
	if v, ok := f["region"]; ok {
		c.Region = v
	}
	if v, ok := f["clues"]; ok {
		c.Clues = util.SplitList(v)
	}
	if err := cs.Update(c); err != nil {
		fatal(err)
	}
	fmt.Println("updated")
}

func del(cs *store.CityStore, args []string) {
	if len(args) == 0 {
		fatal(errors.New("delete requires <id>"))
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fatal(errors.New("id must be numeric"))
	}
	if err := cs.Delete(id); err != nil {
		fatal(err)
	}
	fmt.Println("deleted")
}

func seed(cs *store.CityStore) {
	n, err := store.Seed(cs)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("seeded %d cities\n", n)
}

func reseed(cs *store.CityStore) {
	if err := cs.Wipe(); err != nil {
		fatal(err)
	}
	n, err := store.Seed(cs)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("reseeded %d cities\n", n)
}

func parseFlags(args []string) map[string]string {
	out := map[string]string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "--") {
			continue
		}
		key := strings.TrimPrefix(a, "--")
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			out[key] = args[i+1]
			i++
		} else {
			out[key] = ""
		}
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
