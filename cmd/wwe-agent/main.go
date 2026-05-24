package main

import (
	"flag"
	"strings"
	"fmt"
	"log"
	"os"

	"github.com/yichouchou/awesome-wwe/internal/agent"
	"github.com/yichouchou/awesome-wwe/internal/index"
	"github.com/yichouchou/awesome-wwe/internal/server"
)

func main() {
	initDB := flag.Bool("init", false, "Initialize database")
	addEntity := flag.String("add-entity", "", "Add entity: id:type:name:tier:desc")
	addRel := flag.String("add-rel", "", "Add relationship: from:to:type:desc:chapter")
	query := flag.String("query", "", "Query entity by name or ID")
	listEntities := flag.Bool("list", false, "List all entities")
	stats := flag.Bool("stats", false, "Show database stats")
	generate := flag.String("generate", "", "Generate screenplay for scene")
	serverFlag := flag.Bool("server", false, "Start HTTP server")
	serverAddr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	if *initDB {
		if err := index.InitDB(); err != nil {
			log.Fatalf("Failed to init DB: %v", err)
		}
		fmt.Println("[OK] Database initialized at .wwe/index.db")
		os.Exit(0)
	}

	if *serverFlag {
		srv := server.New()
		log.Printf("Starting WWE Agent HTTP Server on %s", *serverAddr)
		log.Printf("Visit http://localhost%s/ to access the web UI", *serverAddr)
		if err := srv.Run(*serverAddr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
		os.Exit(0)
	}

	if *addEntity != "" {
		parts := strings.Split(*addEntity, ":")
		if len(parts) < 3 {
			log.Fatal("Invalid format. Use: id:type:name:tier:desc")
		}
		e := &index.Entity{ID: parts[0], Type: parts[1], Name: parts[2], Tier: "次要", Description: ""}
		if len(parts) > 3 {
			e.Tier = parts[3]
		}
		if len(parts) > 4 {
			e.Description = parts[4]
		}
		if err := index.AddEntity(e); err != nil {
			log.Fatalf("Failed to add entity: %v", err)
		}
		fmt.Printf("[OK] Added entity: %s (%s)\n", e.Name, e.Type)
		os.Exit(0)
	}

	if *addRel != "" {
		parts := strings.Split(*addRel, ":")
		if len(parts) < 3 {
			log.Fatal("Invalid format. Use: from:to:type:desc:chapter")
		}
		r := &index.Relationship{FromEntity: parts[0], ToEntity: parts[1], Type: parts[2], Chapter: 0}
		if len(parts) > 3 {
			r.Description = parts[3]
		}
		if len(parts) > 4 {
			fmt.Sscanf(parts[4], "%d", &r.Chapter)
		}
		if err := index.AddRelationship(r); err != nil {
			log.Fatalf("Failed to add relationship: %v", err)
		}
		fmt.Printf("[OK] Added relationship: %s --[%s]--> %s\n", r.FromEntity, r.Type, r.ToEntity)
		os.Exit(0)
	}

	if *query != "" {
		results, err := index.QueryEntities(*query)
		if err != nil {
			log.Fatalf("Query failed: %v", err)
		}
		if len(results) == 0 {
			fmt.Println("No entities found")
		} else {
			for _, e := range results {
				fmt.Printf("[%s] %s (%s) - %s [%s]\n", e.Type, e.Name, e.ID, e.Tier, e.Description)
			}
		}
		os.Exit(0)
	}

	if *listEntities {
		entities, err := index.ListEntities()
		if err != nil {
			log.Fatalf("List failed: %v", err)
		}
		for _, e := range entities {
			fmt.Printf("[%s] %s | %s | %s\n", e.Type, e.Name, e.ID, e.Tier)
		}
		os.Exit(0)
	}

	if *stats {
		s, err := index.Stats()
		if err != nil {
			log.Fatalf("Stats failed: %v", err)
		}
		fmt.Printf("Entities: %d | Aliases: %d | Relationships: %d\n", s["entities"], s["aliases"], s["relationships"])
		os.Exit(0)
	}

	if *generate != "" {
		a, err := agent.NewAgent(agent.WithScene(*generate))
		if err != nil {
			log.Fatalf("Failed to create agent: %v", err)
		}
		result, err := a.Generate()
		if err != nil {
			log.Fatalf("Generation failed: %v", err)
		}
		fmt.Printf("\n=== Generated Script ===\n%s\n", result)
		os.Exit(0)
	}

	fmt.Println(`WWE Script Agent - Go + MiniMax-M2.7

Usage:
  wwe-agent --init                    Initialize database
  wwe-agent --add-entity "id:type:name:tier:desc"  Add entity
  wwe-agent --add-rel "from:to:type:desc:chapter"   Add relationship
  wwe-agent --query <name>           Query entity
  wwe-agent --list                   List all entities
  wwe-agent --stats                  Show database stats
  wwe-agent --generate <scene>       Generate screenplay
  wwe-agent --server                 Start HTTP server (default :8080)
  wwe-agent --server --addr :9000    Start HTTP server on port 9000

HTTP Server:
  Visit http://localhost:8080/ to access the web UI
  - Dashboard with stats
  - Entity management (CRUD)
  - Relationship management
  - Script generation
  - Search functionality
`)
}
