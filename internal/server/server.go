package server

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yichouchou/awesome-wwe/internal/index"
)

// Server represents the HTTP server
type Server struct {
	mux     *http.ServeMux
	dbPath  string
}

// New creates a new HTTP server
func New(dbPath string) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		dbPath: dbPath,
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Static files
	s.mux.HandleFunc("/static/", s.handleStatic)

	// API endpoints
	s.mux.HandleFunc("/api/entities", s.handleAPIEntities)
	s.mux.HandleFunc("/api/relationships", s.handleAPIRelationships)
	s.mux.HandleFunc("/api/generate", s.handleAPIGenerate)
	s.mux.HandleFunc("/api/stats", s.handleAPIStats)

	// Web pages
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/entities", s.handleEntitiesPage)
	s.mux.HandleFunc("/relationships", s.handleRelationshipsPage)
	s.mux.HandleFunc("/generate", s.handleGeneratePage)
}

// Run starts the HTTP server
func (s *Server) Run(addr string) error {
	log.Printf("[WWE Agent] HTTP Server starting on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// handleStatic serves static files
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(".", r.URL.Path)
	http.ServeFile(w, r, filePath)
}

// handleIndex serves the main page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	stats, _ := index.Stats()
	entities, _ := index.ListEntities()

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Awesome WWE - 剧本智能体</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); min-height: 100vh; color: #fff; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        header { text-align: center; padding: 40px 0; }
        h1 { font-size: 3em; color: #e94560; text-shadow: 0 0 20px rgba(233,69,96,0.5); }
        .subtitle { color: #a0a0a0; margin-top: 10px; font-size: 1.2em; }
        .stats { display: flex; justify-content: center; gap: 40px; margin: 30px 0; }
        .stat-card { background: rgba(255,255,255,0.1); padding: 20px 40px; border-radius: 15px; text-align: center; backdrop-filter: blur(10px); }
        .stat-number { font-size: 2.5em; color: #e94560; font-weight: bold; }
        .stat-label { color: #a0a0a0; margin-top: 5px; }
        .nav { display: flex; justify-content: center; gap: 20px; margin: 30px 0; flex-wrap: wrap; }
        .nav a { padding: 15px 30px; background: #e94560; color: #fff; text-decoration: none; border-radius: 25px; transition: all 0.3s; font-weight: bold; }
        .nav a:hover { transform: translateY(-3px); box-shadow: 0 10px 30px rgba(233,69,96,0.4); }
        .section { background: rgba(255,255,255,0.05); border-radius: 15px; padding: 30px; margin: 20px 0; }
        h2 { color: #e94560; margin-bottom: 20px; font-size: 1.8em; }
        .entity-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); gap: 15px; }
        .entity-card { background: rgba(255,255,255,0.1); padding: 15px; border-radius: 10px; border-left: 4px solid #e94560; }
        .entity-name { font-size: 1.2em; font-weight: bold; color: #fff; }
        .entity-type { display: inline-block; padding: 3px 10px; background: #e94560; border-radius: 10px; font-size: 0.8em; margin-left: 10px; }
        .entity-tier { color: #a0a0a0; font-size: 0.9em; margin-top: 5px; }
        .footer { text-align: center; padding: 30px; color: #a0a0a0; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🤼 Awesome WWE</h1>
            <p class="subtitle">WWE 剧本智能体 - Go + MiniMax-M2.7</p>
        </header>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">实体</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">关系</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div class="stat-label">角色</div>
            </div>
        </div>
        
        <div class="nav">
            <a href="/entities">📋 角色库</a>
            <a href="/relationships">🔗 关系图</a>
            <a href="/generate">✨ 生成剧本</a>
        </div>
        
        <div class="section">
            <h2>最近添加的角色</h2>
            <div class="entity-grid">
                %s
            </div>
        </div>
        
        <div class="footer">
            <p>Powered by Go + MiniMax-M2.7 + SQLite</p>
        </div>
    </div>
</body>
</html>`, stats["entities"], stats["relationships"], countEntitiesByType(entities), renderEntityCards(entities))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func countEntitiesByType(entities []index.Entity) int {
	count := 0
	for _, e := range entities {
		if e.Type == "角色" {
			count++
		}
	}
	return count
}

func renderEntityCards(entities []index.Entity) string {
	if len(entities) == 0 {
		return "<p style='color:#a0a0a0;'>暂无角色数据</p>"
	}
	
	var sb strings.Builder
	for i, e := range entities {
		if i >= 6 {
			break
		}
		desc := e.Description
		if len(desc) > 60 {
			desc = desc[:60] + "..."
		}
		sb.WriteString(fmt.Sprintf(`<div class="entity-card">
            <div class="entity-name">%s<span class="entity-type">%s</span></div>
            <div class="entity-tier">%s</div>
            <p style="color:#a0a0a0;font-size:0.9em;margin-top:5px;">%s</p>
        </div>`, e.Name, e.Type, e.Tier, desc))
	}
	return sb.String()
}

// handleEntitiesPage serves the entities list page
func (s *Server) handleEntitiesPage(w http.ResponseWriter, r *http.Request) {
	entities, _ := index.ListEntities()
	
	var sb strings.Builder
	for _, e := range entities {
		sb.WriteString(fmt.Sprintf(`<tr>
            <td>%s</td>
            <td><span class="type-badge">%s</span></td>
            <td>%s</td>
            <td>%s</td>
        </tr>`, e.Name, e.Type, e.Tier, e.Description))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>角色库 - Awesome WWE</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); min-height: 100vh; color: #fff; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        header { display: flex; justify-content: space-between; align-items: center; padding: 20px 0; }
        h1 { color: #e94560; }
        .back-btn { padding: 10px 20px; background: #e94560; color: #fff; text-decoration: none; border-radius: 20px; }
        table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }
        th { background: rgba(233,69,96,0.3); color: #e94560; }
        tr:hover { background: rgba(255,255,255,0.05); }
        .type-badge { padding: 3px 10px; background: #e94560; border-radius: 10px; font-size: 0.8em; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>📋 角色库</h1>
            <a href="/" class="back-btn">返回首页</a>
        </header>
        <table>
            <thead>
                <tr>
                    <th>名称</th>
                    <th>类型</th>
                    <th>层级</th>
                    <th>描述</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
    </div>
</body>
</html>`, sb.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// handleRelationshipsPage serves the relationships page
func (s *Server) handleRelationshipsPage(w http.ResponseWriter, r *http.Request) {
	entities, _ := index.ListEntities()
	
	// Get relationships for each entity
	var sb strings.Builder
	for _, e := range entities {
		rels, _ := index.GetRelationships(e.ID)
		for _, rel := range rels {
			sb.WriteString(fmt.Sprintf(`<tr>
                <td>%s</td>
                <td><span class="rel-type">%s</span></td>
                <td>%s</td>
                <td>%s</td>
            </tr>`, rel.FromEntity, rel.Type, rel.ToEntity, rel.Description))
		}
	}

	if sb.Len() == 0 {
		sb.WriteString(`<tr><td colspan="4" style="text-align:center;color:#a0a0a0;">暂无关系数据</td></tr>`)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>关系图 - Awesome WWE</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); min-height: 100vh; color: #fff; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        header { display: flex; justify-content: space-between; align-items: center; padding: 20px 0; }
        h1 { color: #e94560; }
        .back-btn { padding: 10px 20px; background: #e94560; color: #fff; text-decoration: none; border-radius: 20px; }
        table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 15px; text-align: left; border-bottom: 1px solid rgba(255,255,255,0.1); }
        th { background: rgba(233,69,96,0.3); color: #e94560; }
        tr:hover { background: rgba(255,255,255,0.05); }
        .rel-type { padding: 3px 10px; background: #4a90d9; border-radius: 10px; font-size: 0.8em; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔗 关系图</h1>
            <a href="/" class="back-btn">返回首页</a>
        </header>
        <table>
            <thead>
                <tr>
                    <th>起始实体</th>
                    <th>关系类型</th>
                    <th>目标实体</th>
                    <th>描述</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
    </div>
</body>
</html>`, sb.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// handleGeneratePage serves the script generation page
func (s *Server) handleGeneratePage(w http.ResponseWriter, r *http.Request) {
	entities, _ := index.ListEntities()
	
	var sb strings.Builder
	for _, e := range entities {
		sb.WriteString(fmt.Sprintf(`<option value="%s">%s (%s)</option>`, e.ID, e.Name, e.Type))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>生成剧本 - Awesome WWE</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%); min-height: 100vh; color: #fff; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        header { display: flex; justify-content: space-between; align-items: center; padding: 20px 0; }
        h1 { color: #e94560; }
        .back-btn { padding: 10px 20px; background: #e94560; color: #fff; text-decoration: none; border-radius: 20px; }
        .form-group { margin: 20px 0; }
        label { display: block; margin-bottom: 10px; color: #e94560; font-weight: bold; }
        input, select, textarea { width: 100%%; padding: 15px; border: none; border-radius: 10px; background: rgba(255,255,255,0.1); color: #fff; font-size: 1em; }
        textarea { min-height: 150px; resize: vertical; }
        button { padding: 15px 40px; background: #e94560; color: #fff; border: none; border-radius: 25px; font-size: 1.1em; cursor: pointer; transition: all 0.3s; }
        button:hover { transform: translateY(-3px); box-shadow: 0 10px 30px rgba(233,69,96,0.4); }
        .result { margin-top: 30px; padding: 20px; background: rgba(255,255,255,0.05); border-radius: 15px; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>✨ 生成剧本</h1>
            <a href="/" class="back-btn">返回首页</a>
        </header>
        
        <div class="form-group">
            <label>场景描述</label>
            <input type="text" id="scene" placeholder="例如: Raw第100期 Stone Cold vs The Rock 冠军争夺战">
        </div>
        
        <div class="form-group">
            <label>选择角色</label>
            <select id="character" multiple style="height:150px;">
                %s
            </select>
            <small style="color:#a0a0a0;">按住 Ctrl 多选</small>
        </div>
        
        <div class="form-group">
            <label>剧情要求</label>
            <textarea id="requirements" placeholder="描述你想要的剧情走向..."></textarea>
        </div>
        
        <button onclick="generate()">生成剧本</button>
        
        <div id="result" class="result" style="display:none;"></div>
        
        <script>
        async function generate() {
            const scene = document.getElementById('scene').value;
            const result = document.getElementById('result');
            result.style.display = 'block';
            result.textContent = '正在生成，请稍候...';
            
            try {
                const resp = await fetch('/api/generate?scene=' + encodeURIComponent(scene));
                const data = await resp.json();
                result.textContent = data.content || JSON.stringify(data, null, 2);
            } catch (e) {
                result.textContent = '生成失败: ' + e.message;
            }
        }
        </script>
    </div>
</body>
</html>`, sb.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// handleAPIEntities returns entities as JSON
func (s *Server) handleAPIEntities(w http.ResponseWriter, r *http.Request) {
	entities, err := index.ListEntities()
	if err != nil {
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"entities":[`)
	for i, e := range entities {
		if i > 0 {
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, `{"id":"%s","name":"%s","type":"%s","tier":"%s","desc":"%s"}`,
			e.ID, e.Name, e.Type, e.Tier, e.Description)
	}
	fmt.Fprintf(w, `]}`)
}

// handleAPIRelationships returns relationships as JSON
func (s *Server) handleAPIRelationships(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("entity")
	
	w.Header().Set("Content-Type", "application/json")
	
	if entityID != "" {
		rels, err := index.GetRelationships(entityID)
		if err != nil {
			fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
			return
		}
		fmt.Fprintf(w, `{"relationships":[`)
		for i, rel := range rels {
			if i > 0 {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprintf(w, `{"from":"%s","to":"%s","type":"%s","desc":"%s"}`,
				rel.FromEntity, rel.ToEntity, rel.Type, rel.Description)
		}
		fmt.Fprintf(w, `]}`)
	} else {
		fmt.Fprintf(w, `{"relationships":[]}`)
	}
}

// handleAPIGenerate handles script generation request
func (s *Server) handleAPIGenerate(w http.ResponseWriter, r *http.Request) {
	scene := r.URL.Query().Get("scene")
	if scene == "" {
		scene = "Raw 剧情"
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	// Placeholder - in real implementation, call the agent
	fmt.Fprintf(w, `{"scene":"%s","content":"剧本生成功能需要配置 MiniMax API Key。请在环境中设置 MINIMAX_API_KEY 后重新尝试。\n\n当前场景: %s"}`, scene, scene)
}

// handleAPIStats returns database stats
func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	stats, err := index.Stats()
	if err != nil {
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"entities":%d,"aliases":%d,"relationships":%d}`, 
		stats["entities"], stats["aliases"], stats["relationships"])
}