package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DBPath is the path to the database
var DBPath = filepath.Join(".wwe", "index.db")

// Entity represents a WWE character/location/faction
type Entity struct {
	ID            string `json:"id"`
	Type          string `json:"type"` // 角色, 势力, 地点, 物品, 招式
	Name          string `json:"name"`
	Tier          string `json:"tier"` // 核心, 重要, 次要, 装饰
	Description   string `json:"desc"`
	IsProtagonist bool   `json:"is_protagonist"`
	FirstChapter  int    `json:"first_chapter"`
	LastChapter   int    `json:"last_chapter"`
}

// Relationship represents a relationship between entities
type Relationship struct {
	ID          int    `json:"id"`
	FromEntity  string `json:"from_entity"`
	ToEntity    string `json:"to_entity"`
	Type        string `json:"type"` // 对手, 盟友, 同伴, 上司, 下属 etc.
	Description string `json:"description"`
	Chapter     int    `json:"chapter"`
}

// Alias represents an alias for an entity
type Alias struct {
	Alias      string `json:"alias"`
	EntityID   string `json:"entity_id"`
	EntityType string `json:"entity_type"`
}

// InitDB initializes the database
func InitDB() error {
	// Ensure .wwe directory exists
	dir := filepath.Dir(DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite", DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS entities (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL CHECK(type IN ('角色', '势力', '地点', '物品', '招式')),
		canonical_name TEXT NOT NULL,
		tier TEXT DEFAULT '次要' CHECK(tier IN ('核心', '重要', '次要', '装饰')),
		desc TEXT,
		is_protagonist INTEGER DEFAULT 0,
		first_appearance INTEGER,
		last_appearance INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS aliases (
		alias TEXT NOT NULL,
		entity_id TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		PRIMARY KEY (alias, entity_id, entity_type)
	);

	CREATE TABLE IF NOT EXISTS relationships (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_entity TEXT NOT NULL,
		to_entity TEXT NOT NULL,
		type TEXT NOT NULL,
		description TEXT,
		chapter INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(from_entity, to_entity, type)
	);

	CREATE TABLE IF NOT EXISTS state_changes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_id TEXT NOT NULL,
		field TEXT NOT NULL,
		old_value TEXT,
		new_value TEXT,
		reason TEXT,
		chapter INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(type);
	CREATE INDEX IF NOT EXISTS idx_entities_tier ON entities(tier);
	CREATE INDEX IF NOT EXISTS idx_relationships_from ON relationships(from_entity);
	CREATE INDEX IF NOT EXISTS idx_relationships_to ON relationships(to_entity);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// GetDB returns a database connection
func GetDB() (*sql.DB, error) {
	return sql.Open("sqlite", DBPath)
}

// AddEntity adds a new entity
func AddEntity(e *Entity) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO entities (id, type, canonical_name, tier, desc, is_protagonist, first_appearance, last_appearance)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, e.ID, e.Type, e.Name, e.Tier, e.Description, e.IsProtagonist, e.FirstChapter, e.LastChapter)
	return err
}

// QueryEntities searches entities by name or ID
func QueryEntities(query string) ([]Entity, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, type, canonical_name, tier, desc
		FROM entities
		WHERE id LIKE ? OR canonical_name LIKE ? OR id IN (
			SELECT entity_id FROM aliases WHERE alias LIKE ?
		)
		LIMIT 50
	`, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Tier, &e.Description); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

// ListEntities returns all entities
func ListEntities() ([]Entity, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, type, canonical_name, tier, desc FROM entities ORDER BY tier, type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Tier, &e.Description); err != nil {
			return nil, err
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

// AddRelationship adds a new relationship
func AddRelationship(r *Relationship) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT OR REPLACE INTO relationships (from_entity, to_entity, type, description, chapter)
		VALUES (?, ?, ?, ?, ?)
	`, r.FromEntity, r.ToEntity, r.Type, r.Description, r.Chapter)
	return err
}

// GetRelationships returns relationships for an entity
func GetRelationships(entityID string) ([]Relationship, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, from_entity, to_entity, type, description, chapter
		FROM relationships
		WHERE from_entity = ? OR to_entity = ?
	`, entityID, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []Relationship
	for rows.Next() {
		var r Relationship
		if err := rows.Scan(&r.ID, &r.FromEntity, &r.ToEntity, &r.Type, &r.Description, &r.Chapter); err != nil {
			return nil, err
		}
		rels = append(rels, r)
	}
	return rels, rows.Err()
}

// AddAlias adds an alias for an entity
func AddAlias(a *Alias) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`INSERT OR IGNORE INTO aliases (alias, entity_id, entity_type) VALUES (?, ?, ?)`,
		a.Alias, a.EntityID, a.EntityType)
	return err
}

// Stats returns database statistics
func Stats() (map[string]int, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stats := make(map[string]int)
	var count int

	db.QueryRow(`SELECT COUNT(*) FROM entities`).Scan(&count)
	stats["entities"] = count

	db.QueryRow(`SELECT COUNT(*) FROM aliases`).Scan(&count)
	stats["aliases"] = count

	db.QueryRow(`SELECT COUNT(*) FROM relationships`).Scan(&count)
	stats["relationships"] = count

	return stats, nil
}

// GetAllRelationships returns all relationships
func GetAllRelationships() ([]Relationship, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, from_entity, to_entity, type, description, chapter FROM relationships ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []Relationship
	for rows.Next() {
		var r Relationship
		if err := rows.Scan(&r.ID, &r.FromEntity, &r.ToEntity, &r.Type, &r.Description, &r.Chapter); err != nil {
			return nil, err
		}
		rels = append(rels, r)
	}
	return rels, rows.Err()
}

// DeleteEntity deletes an entity by ID
func DeleteEntity(id string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM relationships WHERE from_entity = ? OR to_entity = ?", id, id)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM aliases WHERE entity_id = ?", id)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM entities WHERE id = ?", id)
	return err
}

// DeleteRelationship deletes a relationship by ID
func DeleteRelationship(id int) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM relationships WHERE id = ?", id)
	return err
}
