package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Entity struct {
	ID          string
	Type        string
	Name        string
	Tier        string
	Description string
}

func main() {
	dbPath := ".wwe/index.db"
	os.MkdirAll(".wwe", 0755)
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to open DB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	
	schema := `
	CREATE TABLE IF NOT EXISTS entities (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		canonical_name TEXT NOT NULL,
		tier TEXT DEFAULT '次要',
		desc TEXT,
		is_protagonist INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS aliases (
		alias TEXT NOT NULL,
		entity_id TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		PRIMARY KEY (alias, entity_id)
	);
	CREATE TABLE IF NOT EXISTS relationships (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_entity TEXT NOT NULL,
		to_entity TEXT NOT NULL,
		type TEXT NOT NULL,
		description TEXT,
		chapter INTEGER,
		UNIQUE(from_entity, to_entity, type)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		fmt.Printf("Failed to create schema: %v\n", err)
		os.Exit(1)
	}
	
	starsDir := "meta/wwe-stars"
	files, _ := filepath.Glob(starsDir + "/*.md")
	
	fmt.Printf("Found %d star files\n", len(files))
	
	added := 0
	for _, f := range files {
		entity := parseStarFile(f)
		if entity == nil {
			continue
		}
		
		_, err = db.Exec(`
			INSERT OR REPLACE INTO entities (id, type, canonical_name, tier, desc)
			VALUES (?, ?, ?, ?, ?)
		`, entity.ID, entity.Type, entity.Name, entity.Tier, entity.Description)
		
		if err != nil {
			fmt.Printf("Failed to insert %s: %v\n", entity.Name, err)
			continue
		}
		
		addAliases(db, entity)
		fmt.Printf("✓ Added: %s (%s) - %s\n", entity.Name, entity.Type, entity.Tier)
		added++
	}
	
	fmt.Printf("\n✅ Successfully imported %d WWE stars to database\n", added)
}

func parseStarFile(filename string) *Entity {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil
	}
	
	name := strings.TrimSuffix(filepath.Base(filename), ".md")
	name = strings.ReplaceAll(name, "_", " ")
	
	body := string(content)
	entityType := "角色"
	tier := determineTier(name, body)
	desc := extractDescription(body, name)
	id := generateID(name)
	
	return &Entity{
		ID:          id,
		Type:        entityType,
		Name:        name,
		Tier:        tier,
		Description: desc,
	}
}

func determineTier(name string, body string) string {
	coreNames := []string{"Cody Rhodes", "Roman Reigns", "Seth Rollins", "Stone Cold", "The Rock", 
		"Triple H", "Rhea Ripley", "Bianca Belair", "Becky Lynch", "Finn Bálor", "Drew McIntyre",
		"Stone Cold Steve Austin"}
	
	for _, core := range coreNames {
		if strings.Contains(name, core) {
			return "核心"
		}
	}
	
	importantNames := []string{"Gunther", "LA Knight", "Jey Uso", "Logan Paul", "Kevin Owens", 
		"Shinsuke Nakamura", "Kofi Kingston", "Bron Breakker", "Damien Priest", "Iyo Sky", 
		"Tiffany Stratton", "Solo Sikoa", "The Rock", "Paul Heyman"}
	
	for _, imp := range importantNames {
		if strings.Contains(name, imp) {
			return "重要"
		}
	}
	
	return "次要"
}

func extractDescription(body string, name string) string {
	var desc strings.Builder
	desc.WriteString(name)
	
	if strings.Contains(body, "**别名**") {
		idx := strings.Index(body, "**别名**")
		section := body[idx:min(idx+200, len(body))]
		if lines := strings.Split(section, "\n"); len(lines) > 1 {
			aliasLine := strings.TrimSpace(lines[1])
			if len(aliasLine) > 2 {
				desc.WriteString(" | ")
				desc.WriteString(aliasLine)
			}
		}
	}
	
	if strings.Contains(body, "标志性动作") {
		idx := strings.Index(body, "标志性动作")
		section := body[idx:min(idx+300, len(body))]
		if lines := strings.Split(section, "\n"); len(lines) > 1 {
			moveLine := strings.TrimSpace(lines[1])
			if len(moveLine) > 2 && !strings.HasPrefix(moveLine, "**") {
				desc.WriteString(" | 招牌: ")
				desc.WriteString(moveLine)
			}
		}
	}
	
	result := desc.String()
	if len(result) > 300 {
		result = result[:300] + "..."
	}
	return result
}

func generateID(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "-", "_")
	id = strings.ReplaceAll(id, "'", "")
	id = strings.ReplaceAll(id, "__", "_")
	
	var clean strings.Builder
	for _, c := range id {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			clean.WriteRune(c)
		}
	}
	return clean.String()
}

func addAliases(db *sql.DB, entity *Entity) {
	aliases := generateAliases(entity.Name)
	for _, alias := range aliases {
		if alias != entity.Name && alias != "" {
			db.Exec(`INSERT OR IGNORE INTO aliases (alias, entity_id, entity_type) VALUES (?, ?, ?)`,
				alias, entity.ID, entity.Type)
		}
	}
}

func generateAliases(name string) []string {
	var aliases []string
	aliases = append(aliases, name)
	
	mappings := map[string][]string{
		"Stone Cold Steve Austin": {"Austin", "Steve Austin", "Stone Cold", "奥斯汀"},
		"The Rock": {"Rock", "Dwayne Johnson", "巨石强森"},
		"Cody Rhodes": {"Cody", "科迪"},
		"Roman Reigns": {"Reigns", "罗曼"},
		"Seth Rollins": {"Rollins", "塞斯"},
		"Finn Bálor": {"Balor", "Finn", "贝莱尔"},
		"Drew McIntyre": {"Drew", "德鲁"},
		"Rhea Ripley": {"Ripley", "蕾亚"},
		"Bianca Belair": {"Bianca", "比安卡"},
		"Becky Lynch": {"Becky", "贝基"},
		"Triple H": {"HHH", "Paul Levesque"},
		"Gunther": {"冈瑟"},
		"LA Knight": {"Knight", "LA奈特"},
		"Jey Uso": {"Jey", "杰"},
		"Kevin Owens": {"Owens", "凯文"},
		"Shinsuke Nakamura": {"Nakamura", "中村"},
		"Kofi Kingston": {"Kofi", "科菲"},
		"Logan Paul": {"Logan", "洛根"},
		"Bron Breakker": {"Bron", "布朗"},
		"Damien Priest": {"Priest", "达米安"},
		"Iyo Sky": {"Iyo", "紫雷"},
		"Tiffany Stratton": {"Tiffany", "蒂芙尼"},
		"Solo Sikoa": {"Solo", "索洛"},
		"Paul Heyman": {"Heyman", "保罗"},
		"Shawn Michaels": {"Michaels", "肖恩"},
		"Charlotte Flair": {"Charlotte", "夏洛特"},
		"Asuka": {"明日香"},
		"Bayley": {"贝莉"},
	}
	
	for key, value := range mappings {
		if strings.Contains(name, key) {
			aliases = append(aliases, value...)
		}
	}
	
	return aliases
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
