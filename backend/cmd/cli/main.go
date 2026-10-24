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
	case "stats":
		stats(cs, os.Args[2:])
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

func stats(cs *store.CityStore, args []string) {
	if len(args) == 0 {
		fmt.Println("usage: harady-cli stats {top|bottom|show|summary|games|recent|reset} [args]")
		os.Exit(1)
	}
	switch args[0] {
	case "top":
		limit := 20
		if len(args) >= 2 {
			if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
				limit = n
			}
		}
		rows, err := cs.TopCities(limit)
		if err != nil {
			fatal(err)
		}
		printStats(rows)
	case "bottom":
		limit := 20
		if len(args) >= 2 {
			if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
				limit = n
			}
		}
		rows, err := cs.BottomCities(limit)
		if err != nil {
			fatal(err)
		}
		printStats(rows)
	case "show":
		if len(args) < 2 {
			fatal(errors.New("stats show <city>"))
		}
		stat, err := cs.CityStatByName(args[1])
		if err != nil {
			fatal(err)
		}
		fmt.Printf("%-30s %5d  (%s)\n", stat.Name, stat.Count, stat.Region)
	case "games":
		limit := 20
		if len(args) >= 2 {
			if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
				limit = n
			}
		}
		rows, err := cs.LongestGames(limit)
		if err != nil {
			fatal(err)
		}
		printGames(rows)
	case "recent":
		limit := 20
		if len(args) >= 2 {
			if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
				limit = n
			}
		}
		rows, err := cs.RecentGames(limit)
		if err != nil {
			fatal(err)
		}
		printGames(rows)
	case "summary":
		sum, err := cs.Summary()
		if err != nil {
			fatal(err)
		}
		fmt.Println("── cities ──")
		fmt.Printf("cities in pack       : %d\n", sum.TotalCities)
		fmt.Printf("submissions recorded : %d\n", sum.TotalSubmissions)
		fmt.Printf("cities ever played   : %d\n", sum.UniqueCitiesUsed)
		fmt.Printf("cities never played  : %d\n", sum.NeverUsed)
		fmt.Printf("learned cities       : %d\n", sum.LearnedCities)
		gsum, err := cs.GameSummary()
		if err != nil {
			fatal(err)
		}
		fmt.Println()
		fmt.Println("── games ──")
		fmt.Printf("games played         : %d\n", gsum.TotalGames)
		fmt.Printf("total moves          : %d\n", gsum.TotalMoves)
		fmt.Printf("longest chain ever   : %d\n", gsum.MaxChainLength)
		fmt.Printf("longest (no bots)    : %d\n", gsum.MaxChainHuman)
		fmt.Printf("avg chain length     : %.1f\n", gsum.AvgChainLength)
		fmt.Printf("games with bots      : %d\n", gsum.GamesWithBots)
	case "reset":
		if err := cs.ResetSubmissions(); err != nil {
			fatal(err)
		}
		if err := cs.ResetGames(); err != nil {
			fatal(err)
		}
		fmt.Println("submissions and games reset")
	default:
		fatal(fmt.Errorf("unknown stats subcommand: %s", args[0]))
	}
}

func printStats(rows []store.CityStat) {
	if len(rows) == 0 {
		fmt.Println("(no data)")
		return
	}
	fmt.Printf("%-4s %-30s %-25s %6s\n", "ID", "NAME", "REGION", "COUNT")
	for _, r := range rows {
		fmt.Printf("%-4d %-30s %-25s %6d\n", r.ID, r.Name, r.Region, r.Count)
	}
}

func printGames(rows []*store.GameStat) {
	if len(rows) == 0 {
		fmt.Println("(no games)")
		return
	}
	fmt.Printf("%-4s %-8s %6s %8s %5s %-15s %-16s\n", "ID", "ROOM", "CHAIN", "PLAYERS", "BOTS", "WINNER", "ENDED")
	for _, g := range rows {
		bots := "no"
		if g.HasBots {
			bots = "yes"
		}
		winner := g.WinnerNickname
		if winner == "" {
			winner = "—"
		}
		fmt.Printf("%-4d %-8s %6d %8d %5s %-15s %-16s\n",
			g.ID, g.RoomCode, g.ChainLength, g.PlayersCount, bots, winner, g.EndedAt.Format("2006-01-02 15:04"))
	}
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
