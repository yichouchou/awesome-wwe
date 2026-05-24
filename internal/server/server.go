package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/yichouchou/awesome-wwe/internal/index"
)

type Server struct {
	mux *http.ServeMux
}

func New() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/api/init", s.handleAPIInit)
	s.mux.HandleFunc("/api/entities", s.handleAPIEntities)
	s.mux.HandleFunc("/api/entities/add", s.handleAPIAddEntity)
	s.mux.HandleFunc("/api/entities/delete", s.handleAPIDeleteEntity)
	s.mux.HandleFunc("/api/entities/search", s.handleAPISearchEntities)
	s.mux.HandleFunc("/api/relationships", s.handleAPIRelationships)
	s.mux.HandleFunc("/api/relationships/add", s.handleAPIAddRelationship)
	s.mux.HandleFunc("/api/stats", s.handleAPIStats)
	s.mux.HandleFunc("/api/generate", s.handleAPIGenerate)
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/entities", s.handleEntitiesPage)
	s.mux.HandleFunc("/relationships", s.handleRelationshipsPage)
	s.mux.HandleFunc("/generate", s.handleGeneratePage)
	s.mux.HandleFunc("/add-entity", s.handleAddEntityPage)
	s.mux.HandleFunc("/add-relationship", s.handleAddRelationshipPage)
	s.mux.HandleFunc("/search", s.handleSearchPage)
}

func (s *Server) Run(addr string) error {
	fmt.Printf("\n🎬 Awesome WWE Agent - HTTP Server\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("   🌐 http://localhost%s\n", addr)
	fmt.Printf("   📊 Dashboard: http://localhost%s/\n", addr)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("   📋 GET  /api/stats              - 统计信息\n")
	fmt.Printf("   📋 GET  /api/entities           - 实体列表\n")
	fmt.Printf("   ➕ POST /api/entities/add       - 添加实体\n")
	fmt.Printf("   🔍 GET  /api/entities/search   - 搜索实体\n")
	fmt.Printf("   🗑️  POST /api/entities/delete   - 删除实体\n")
	fmt.Printf("   🔗 GET  /api/relationships      - 关系列表\n")
	fmt.Printf("   ➕ POST /api/relationships/add  - 添加关系\n")
	fmt.Printf("   ✨ GET  /api/generate?scene=X  - 生成剧本\n")
	fmt.Printf("   🔄 POST /api/init               - 初始化数据库\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleAPIInit(w http.ResponseWriter, r *http.Request) {
	if err := index.InitDB(); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "数据库初始化成功"})
}

func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	stats, err := index.Stats()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "stats": stats})
}

func (s *Server) handleAPIEntities(w http.ResponseWriter, r *http.Request) {
	entities, err := index.ListEntities()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "entities": entities})
}

func (s *Server) handleAPIAddEntity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}
	var data struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		Name        string `json:"name"`
		Tier        string `json:"tier"`
		Description string `json:"desc"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	if data.Tier == "" {
		data.Tier = "次要"
	}
	e := &index.Entity{ID: data.ID, Type: data.Type, Name: data.Name, Tier: data.Tier, Description: data.Description}
	if err := index.AddEntity(e); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "实体添加成功"})
}

func (s *Server) handleAPIDeleteEntity(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Missing id"})
		return
	}
	if err := index.DeleteEntity(id); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "实体删除成功"})
}

func (s *Server) handleAPISearchEntities(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Missing query"})
		return
	}
	entities, err := index.QueryEntities(query)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "entities": entities})
}

func (s *Server) handleAPIRelationships(w http.ResponseWriter, r *http.Request) {
	rels, err := index.GetAllRelationships()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "relationships": rels})
}

func (s *Server) handleAPIAddRelationship(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}
	var data struct {
		From        string `json:"from"`
		To          string `json:"to"`
		Type        string `json:"type"`
		Description string `json:"desc"`
		Chapter     int    `json:"chapter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	r2 := &index.Relationship{FromEntity: data.From, ToEntity: data.To, Type: data.Type, Description: data.Description, Chapter: data.Chapter}
	if err := index.AddRelationship(r2); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "关系添加成功"})
}

func (s *Server) handleAPIGenerate(w http.ResponseWriter, r *http.Request) {
	scene := r.URL.Query().Get("scene")
	if scene == "" {
		scene = "WWE Raw 剧情"
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success":true,"scene":"%s","content":"剧本生成需要配置 MINIMAX_API_KEY 环境变量。当前场景: %s\n请设置: export MINIMAX_API_KEY=your-key"}`, scene, scene)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	stats, _ := index.Stats()
	entities, _ := index.ListEntities()
	charCount := countByType(entities, "角色")
	locCount := countByType(entities, "地点")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Awesome WWE - WWE剧本智能体</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#fff;min-height:100vh}
.container{max-width:1400px;margin:0 auto;padding:30px}
header{text-align:center;padding:50px 0;border-bottom:1px solid rgba(255,255,255,0.1);margin-bottom:40px}
h1{font-size:3.5em;color:#e94560;text-shadow:0 0 40px rgba(233,69,96,0.5);margin-bottom:15px}
.subtitle{color:#888;font-size:1.3em}
.stats-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:25px;margin-bottom:50px}
.stat-card{background:linear-gradient(135deg,rgba(233,69,96,0.15),rgba(233,69,96,0.05));border:1px solid rgba(233,69,96,0.3);padding:35px;border-radius:20px;text-align:center;transition:all 0.3s}
.stat-card:hover{transform:translateY(-8px);box-shadow:0 20px 50px rgba(233,69,96,0.3)}
.stat-number{font-size:3.5em;color:#e94560;font-weight:bold}
.stat-label{color:#888;margin-top:10px;font-size:1.1em}
.section{background:rgba(255,255,255,0.03);border-radius:20px;padding:35px;margin-bottom:30px;border:1px solid rgba(255,255,255,0.1)}
h2{color:#e94560;margin-bottom:25px;font-size:1.8em}
.nav-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:20px}
.nav-card{background:linear-gradient(135deg,rgba(233,69,96,0.15),rgba(233,69,96,0.05));border:1px solid rgba(233,69,96,0.2);padding:28px;border-radius:15px;text-decoration:none;color:#fff;transition:all 0.3s;display:block}
.nav-card:hover{transform:translateY(-8px);box-shadow:0 20px 50px rgba(233,69,96,0.3)}
.nav-icon{font-size:3em;margin-bottom:15px}
.nav-title{font-size:1.4em;font-weight:bold;margin-bottom:10px}
.nav-desc{color:#888;font-size:1em}
.entity-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:15px}
.entity-card{background:rgba(255,255,255,0.05);padding:20px;border-radius:12px;border-left:4px solid #e94560;transition:all 0.3s}
.entity-card:hover{background:rgba(255,255,255,0.1)}
.entity-name{font-size:1.2em;font-weight:bold;margin-bottom:8px;color:#fff}
.entity-desc{color:#888;font-size:0.95em;margin:10px 0;line-height:1.5}
.entity-meta{display:flex;gap:10px;flex-wrap:wrap;margin-top:12px}
.badge{padding:5px 15px;border-radius:20px;font-size:0.85em}
.badge-type{background:#e94560}
.badge-tier{background:#4a90d9}
.footer{text-align:center;padding:50px;color:#666;border-top:1px solid rgba(255,255,255,0.1);margin-top:50px}
.empty-msg{text-align:center;padding:40px;color:#666}
</style>
</head>
<body>
<div class="container">
<header>
<h1>🤼 Awesome WWE</h1>
<p class="subtitle">WWE 剧本智能体 - Go + MiniMax-M2.7 + SQLite</p>
</header>

<div class="stats-grid">
<div class="stat-card"><div class="stat-number">%d</div><div class="stat-label">📋 实体总数</div></div>
<div class="stat-card"><div class="stat-number">%d</div><div class="stat-label">🔗 关系总数</div></div>
<div class="stat-card"><div class="stat-number">%d</div><div class="stat-label">👥 角色数量</div></div>
<div class="stat-card"><div class="stat-number">%d</div><div class="stat-label">📍 地点数量</div></div>
</div>

<div class="section">
<h2>🚀 功能导航</h2>
<div class="nav-grid">
<a href="/entities" class="nav-card"><div class="nav-icon">📋</div><div class="nav-title">角色库管理</div><div class="nav-desc">查看、搜索、筛选所有实体</div></a>
<a href="/add-entity" class="nav-card"><div class="nav-icon">➕</div><div class="nav-title">添加实体</div><div class="nav-desc">创建新的角色、地点、势力</div></a>
<a href="/add-relationship" class="nav-card"><div class="nav-icon">🔗</div><div class="nav-title">添加关系</div><div class="nav-desc">建立实体间的关联</div></a>
<a href="/search" class="nav-card"><div class="nav-icon">🔍</div><div class="nav-title">搜索查询</div><div class="nav-desc">快速查找实体和关系</div></a>
<a href="/relationships" class="nav-card"><div class="nav-icon">🌐</div><div class="nav-title">关系图谱</div><div class="nav-desc">查看实体关系网络</div></a>
<a href="/generate" class="nav-card"><div class="nav-icon">✨</div><div class="nav-title">生成剧本</div><div class="nav-desc">AI 驱动的剧情生成</div></a>
</div>
</div>

<div class="section">
<h2>📺 最近角色</h2>
<div class="entity-grid">%s</div>
</div>

<div class="footer">
<p>Powered by Awesome WWE | Go + MiniMax-M2.7</p>
</div>
</div>
</body>
</html>`, stats["entities"], stats["relationships"], charCount, locCount, renderEntityCards(entities))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func countByType(entities []index.Entity, t string) int {
	c := 0
	for _, e := range entities {
		if e.Type == t {
			c++
		}
	}
	return c
}

func renderEntityCards(entities []index.Entity) string {
	if len(entities) == 0 {
		return "<div class='empty-msg'>暂无实体数据 - <a href='/add-entity' style='color:#e94560'>添加第一个实体</a></div>"
	}
	var sb strings.Builder
	for i, e := range entities {
		if i >= 8 {
			break
		}
		desc := e.Description
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		sb.WriteString(fmt.Sprintf(`<div class="entity-card"><div class="entity-name">%s</div><div class="entity-desc">%s</div><div class="entity-meta"><span class="badge badge-type">%s</span><span class="badge badge-tier">%s</span></div></div>`, e.Name, desc, e.Type, e.Tier))
	}
	return sb.String()
}

func (s *Server) handleEntitiesPage(w http.ResponseWriter, r *http.Request) {
	entities, _ := index.ListEntities()
	var rows strings.Builder
	for _, e := range entities {
		desc := e.Description
		if len(desc) > 120 {
			desc = desc[:120] + "..."
		}
		rows.WriteString(fmt.Sprintf(`<tr><td><strong>%s</strong></td><td><span class="badge badge-type">%s</span></td><td><span class="badge badge-tier">%s</span></td><td style="max-width:350px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">%s</td><td><button onclick="deleteEntity('%s')" class="btn-delete">🗑️ 删除</button></td></tr>`, e.Name, e.Type, e.Tier, desc, e.ID))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>角色库管理 - Awesome WWE</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#fff;min-height:100vh}
.container{max-width:1400px;margin:0 auto;padding:30px}
header{display:flex;justify-content:space-between;align-items:center;padding:30px 0;border-bottom:1px solid rgba(255,255,255,0.1);margin-bottom:30px}
h1{color:#e94560;font-size:2.2em}.btn{padding:12px 25px;background:#e94560;color:#fff;text-decoration:none;border-radius:25px;font-weight:bold;border:none;cursor:pointer}
.btn:hover{transform:translateY(-3px);box-shadow:0 10px 30px rgba(233,69,96,0.4)}
.filter-bar{display:flex;gap:15px;margin-bottom:25px;flex-wrap:wrap}
.filter-bar input,.filter-bar select{padding:12px 20px;background:rgba(255,255,255,0.1);border:1px solid rgba(255,255,255,0.2);border-radius:10px;color:#fff;font-size:1em}
.filter-bar select option{background:#1a1a2e}
table{width:100%%;border-collapse:collapse;background:rgba(255,255,255,0.03);border-radius:15px;overflow:hidden}
th{background:rgba(233,69,96,0.2);padding:18px;text-align:left;color:#e94560;font-weight:bold}
td{padding:18px;border-bottom:1px solid rgba(255,255,255,0.05)}
tr:hover{background:rgba(255,255,255,0.05)}
.badge{padding:5px 15px;border-radius:20px;font-size:0.85em}.badge-type{background:#e94560}.badge-tier{background:#4a90d9}
.btn-delete{padding:10px 20px;background:#dc3545;color:#fff;border:none;border-radius:20px;cursor:pointer}
.btn-delete:hover{background:#c82333}
.empty{text-align:center;padding:60px;color:#888}
</style></head>
<body>
<div class="container">
<header><h1>📋 角色库管理</h1>
<div style="display:flex;gap:15px;">
<a href="/add-entity" class="btn">➕ 添加实体</a>
<a href="/" class="btn">🏠 返回首页</a>
</div></header>

<div class="filter-bar">
<select id="filterType" onchange="filterEntities()"><option value="">所有类型</option><option value="角色">角色</option><option value="势力">势力</option><option value="地点">地点</option><option value="物品">物品</option><option value="招式">招式</option></select>
<select id="filterTier" onchange="filterEntities()"><option value="">所有层级</option><option value="核心">核心</option><option value="重要">重要</option><option value="次要">次要</option><option value="装饰">装饰</option></select>
<input type="text" id="searchInput" placeholder="🔍 搜索名称或描述..." onkeyup="filterEntities()">
</div>

%s

<script>
function filterEntities(){const type=document.getElementById('filterType').value;const tier=document.getElementById('filterTier').value;const search=document.getElementById('searchInput').value.toLowerCase();document.querySelectorAll('tbody tr').forEach(row=>{const t=row.querySelector('.badge-type')?.textContent||'';const ti=row.querySelector('.badge-tier')?.textContent||'';const name=row.querySelector('strong')?.textContent||'';const desc=row.querySelector('td:nth-child(4)')?.textContent||'';const matchType=!type||t.includes(type);const matchTier=!tier||ti.includes(tier);const matchSearch=!search||name.toLowerCase().includes(search)||desc.toLowerCase().includes(search);row.style.display=matchType&&matchTier&&matchSearch?'':'none'})}
async function deleteEntity(id){if(!confirm('确定删除这个实体吗？'))return;try{const resp=await fetch('/api/entities/delete?id='+id,{method:'POST'});const data=await resp.json();if(data.success)location.reload();else alert('删除失败: '+data.error)}catch(e){alert('删除失败: '+e.message)}}
</script>
</body></html>`, formatEntitiesTable(rows.String()))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func formatEntitiesTable(content string) string {
	if content == "" {
		return "<div class='empty'>📋 暂无实体数据 <a href='/add-entity' style='color:#e94560'>添加第一个实体</a></div>"
	}
	return "<table><thead><tr><th>名称</th><th>类型</th><th>层级</th><th>描述</th><th>操作</th></tr></thead><tbody>" + content + "</tbody></table>"
}

func (s *Server) handleAddEntityPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>添加实体 - Awesome WWE</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#fff;min-height:100vh}
.container{max-width:800px;margin:0 auto;padding:30px}
header{display:flex;justify-content:space-between;align-items:center;padding:30px 0;border-bottom:1px solid rgba(255,255,255,0.1);margin-bottom:30px}
h1{color:#e94560;font-size:2.2em}.back-btn{padding:12px 25px;background:#e94560;color:#fff;text-decoration:none;border-radius:25px}
.form-card{background:rgba(255,255,255,0.03);border-radius:20px;padding:40px;border:1px solid rgba(255,255,255,0.1)}
.form-group{margin-bottom:25px}label{display:block;margin-bottom:10px;color:#e94560;font-weight:bold;font-size:1.1em}
input,select,textarea{width:100%%;padding:15px 20px;background:rgba(255,255,255,0.1);border:1px solid rgba(255,255,255,0.2);border-radius:12px;color:#fff;font-size:1em;transition:all 0.3s}
input:focus,select:focus,textarea:focus{outline:none;border-color:#e94560;box-shadow:0 0 20px rgba(233,69,96,0.3)}
textarea{min-height:120px;resize:vertical}select option{background:#1a1a2e}
.submit-btn{width:100%%;padding:18px;background:linear-gradient(135deg,#e94560,#c73e54);color:#fff;border:none;border-radius:12px;font-size:1.2em;font-weight:bold;cursor:pointer;transition:all 0.3s}
.submit-btn:hover{transform:translateY(-3px);box-shadow:0 15px 40px rgba(233,69,96,0.4)}
.hint{color:#888;font-size:0.9em;margin-top:8px}
.success{padding:20px;background:rgba(40,167,69,0.2);border:1px solid #28a745;border-radius:12px;margin-bottom:20px;display:none}
</style></head>
<body>
<div class="container">
<header><h1>➕ 添加实体</h1><a href="/entities" class="back-btn">← 返回列表</a></header>
<div id="successMsg" class="success">✓ 实体添加成功！</div>
<div class="form-card">
<form id="entityForm">
<div class="form-group"><label>实体ID *</label><input type="text" id="entityId" placeholder="例如: stone_cold, the_rock" required><p class="hint">唯一标识符，使用英文下划线格式</p></div>
<div class="form-group"><label>类型 *</label><select id="entityType" required><option value="角色">角色</option><option value="势力">势力</option><option value="地点">地点</option><option value="物品">物品</option><option value="招式">招式</option></select></div>
<div class="form-group"><label>名称 *</label><input type="text" id="entityName" placeholder="例如: Stone Cold Steve Austin" required></div>
<div class="form-group"><label>层级</label><select id="entityTier"><option value="核心">核心 - 主要角色</option><option value="重要">重要 - 重要配角</option><option value="次要" selected>次要 - 一般角色</option><option value="装饰">装饰 - 背景角色</option></select></div>
<div class="form-group"><label>描述</label><textarea id="entityDesc" placeholder="描述实体的背景、特点、能力等..."></textarea></div>
<button type="submit" class="submit-btn">✨ 添加实体</button>
</form></div></div>
<script>
document.getElementById('entityForm').onsubmit=async function(e){e.preventDefault();const data={id:document.getElementById('entityId').value,type:document.getElementById('entityType').value,name:document.getElementById('entityName').value,tier:document.getElementById('entityTier').value,desc:document.getElementById('entityDesc').value};try{const resp=await fetch('/api/entities/add',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});const result=await resp.json();if(result.success){document.getElementById('successMsg').style.display='block';setTimeout(()=>location.href='/entities',1500)}else{alert('添加失败: '+result.error)}}catch(e){alert('添加失败: '+e.message)}};
</script>
</body></html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleAddRelationshipPage(w http.ResponseWriter, r *http.Request) {
	entities, _ := index.ListEntities()
	var opts strings.Builder
	for _, e := range entities {
		opts.WriteString(fmt.Sprintf(`<option value="%s">%s (%s)</option>`, e.ID, e.Name, e.Type))
	}
	if opts.Len() == 0 {
		opts.WriteString("<option value=''>暂无实体，请先添加实体</option>")
	}

	types := []string{"对手", "盟友", "同伴", "上司", "下属", "师徒", "血缘", "同门", "恩怨", "合作", "竞争", "其他"}
	var typeOpts strings.Builder
	for _, t := range types {
		typeOpts.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, t, t))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>添加关系 - Awesome WWE</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#fff;min-height:100vh}
.container{max-width:800px;margin:0 auto;padding:30px}
header{display:flex;justify-content:space-between;align-items:center;padding:30px 0;border-bottom:1px solid rgba(255,255,255,0.1);margin-bottom:30px}
h1{color:#e94560;font-size:2.2em}.back-btn{padding:12px 25px;background:#e94560;color:#fff;text-decoration:none;border-radius:25px}
.form-card{background:rgba(255,255,255,0.03);border-radius:20px;padding:40px;border:1px solid rgba(255,255,255,0.1)}
.form-group{margin-bottom:25px}label{display:block;margin-bottom:10px;color:#e94560;font-weight:bold;font-size:1.1em}
input,select{width:100%%;padding:15px 20px;background:rgba(255,255,255,0.1);border:1px solid rgba(255,255,255,0.2);border-radius:12px;color:#fff;font-size:1em}
input:focus,select:focus{outline:none;border-color:#e94560;box-shadow:0 0 20px rgba(233,69,96,0.3)}select option{background:#1a1a2e}
.submit-btn{width:100%%;padding:18px;background:linear-gradient(135deg,#e94560,#c73e54);color:#fff;border:none;border-radius:12px;font-size:1.2em;font-weight:bold;cursor:pointer;transition:all 0.3s}
.submit-btn:hover{transform:translateY(-3px);box-shadow:0 15px 40px rgba(233,69,96,0.4)}
.arrow{text-align:center;font-size:2.5em;color:#e94560;padding:15px 0}
.success{padding:20px;background:rgba(40,167,69,0.2);border:1px solid #28a745;border-radius:12px;margin-bottom:20px;display:none}
</style></head>
<body>
<div class="container">
<header><h1>🔗 添加关系</h1><a href="/relationships" class="back-btn">← 返回列表</a></header>
<div id="successMsg" class="success">✓ 关系添加成功！</div>
<div class="form-card">
<form id="relForm">
<div class="form-group"><label>起始实体 *</label><select id="fromEntity" required><option value="">选择起始实体...</option>%s</select></div>
<div class="arrow">↓</div>
<div class="form-group"><label>关系类型 *</label><select id="relType" required>%s</select></div>
<div class="arrow">↓</div>
<div class="form-group"><label>目标实体 *</label><select id="toEntity" required><option value="">选择目标实体...</option>%s</select></div>
<div class="form-group"><label>描述</label><input type="text" id="relDesc" placeholder="例如: WWE黄金时代经典对决"></div>
<div class="form-group"><label>章节 (可选)</label><input type="number" id="relChapter" placeholder="关系首次出现的章节"></div>
<button type="submit" class="submit-btn">✨ 添加关系</button>
</form></div></div>
<script>
document.getElementById('relForm').onsubmit=async function(e){e.preventDefault();const data={from:document.getElementById('fromEntity').value,to:document.getElementById('toEntity').value,type:document.getElementById('relType').value,desc:document.getElementById('relDesc').value,chapter:parseInt(document.getElementById('relChapter').value)||0};try{const resp=await fetch('/api/relationships/add',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});const result=await resp.json();if(result.success){document.getElementById('successMsg').style.display='block';setTimeout(()=>location.href='/relationships',1500)}else{alert('添加失败: '+result.error)}}catch(e){alert('添加失败: '+e.message)}};
</script>
</body></html>`, opts.String(), typeOpts.String(), opts.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleSearchPage(w http.ResponseWriter, r *http.Request) {
	entities, _ := index.ListEntities()
	var opts strings.Builder
	for _, e := range entities {
		opts.WriteString(fmt.Sprintf(`<option value="%s">%s (%s)</option>`, e.ID, e.Name, e.Type))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>搜索查询 - Awesome WWE</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#fff;min-height:100vh}
.container{max-width:1200px;margin:0 auto;padding:30px}
header{display:flex;justify-content:space-between;align-items:center;padding:30px 0;border-bottom:1px solid rgba(255,255,255,0.1);margin-bottom:30px}
h1{color:#e94560;font-size:2.2em}.back-btn{padding:12px 25px;background:#e94560;color:#fff;text-decoration:none;border-radius:25px}
.search-box{background:rgba(255,255,255,0.03);border-radius:20px;padding:30px;border:1px solid rgba(255,255,255,0.1);margin-bottom:30px}
.form-group{margin-bottom:20px}label{display:block;margin-bottom:10px;color:#e94560;font-weight:bold}
input,select{width:100%%;padding:15px 20px;background:rgba(255,255,255,0.1);border:1px solid rgba(255,255,255,0.2);border-radius:12px;color:#fff;font-size:1em}
input:focus,select:focus{outline:none;border-color:#e94560}select option{background:#1a1a2e}
.search-btn{width:100%%;padding:18px;background:#e94560;color:#fff;border:none;border-radius:12px;font-size:1.2em;font-weight:bold;cursor:pointer}
.search-btn:hover{transform:translateY(-3px);box-shadow:0 10px 30px rgba(233,69,96,0.4)}
.result-card{background:rgba(255,255,255,0.05);border-radius:15px;padding:25px;margin-bottom:15px;border-left:4px solid #e94560}
.result-name{font-size:1.3em;font-weight:bold;color:#fff;margin-bottom:10px}
.result-desc{color:#888;font-size:0.95em;margin-bottom:10px}
.result-meta{display:flex;gap:10px}
.badge{padding:5px 15px;border-radius:20px;font-size:0.85em}
.badge-type{background:#e94560}
.badge-tier{background:#4a90d9}
.results-section{margin-top:30px}
.results-title{color:#e94560;font-size:1.5em;margin-bottom:20px}
.no-results{text-align:center;padding:40px;color:#888}
</style></head>
<body>
<div class="container">
<header><h1>🔍 搜索查询</h1><a href="/" class="back-btn">🏠 返回首页</a></header>

<div class="search-box">
<form id="searchForm">
<div class="form-group">
<label>快速搜索</label>
<input type="text" id="quickSearch" placeholder="输入名称或描述关键字..." oninput="doSearch()">
</div>
<div class="form-group">
<label>选择实体查看详情和关系</label>
<select id="entitySelect" onchange="showEntityDetails()">
<option value="">-- 选择实体 --</option>
%s
</select>
</div>
<button type="button" class="search-btn" onclick="showAllEntities()">🔍 搜索所有实体</button>
</form>
</div>

<div id="resultsSection" class="results-section" style="display:none;">
<h2 class="results-title">搜索结果</h2>
<div id="resultsContainer"></div>
</div>

<div id="detailsSection" class="results-section" style="display:none;">
<h2 class="results-title">实体详情</h2>
<div id="entityDetails"></div>
</div>
</div>

<script>
let allEntities = [];
async function doSearch() {
    const q = document.getElementById('quickSearch').value;
    if (!q) { document.getElementById('resultsSection').style.display='none'; return; }
    const resp = await fetch('/api/entities/search?q=' + encodeURIComponent(q));
    const data = await resp.json();
    showResults(data.entities || []);
}
async function showAllEntities() {
    const resp = await fetch('/api/entities');
    const data = await resp.json();
    showResults(data.entities || []);
}
async function showEntityDetails() {
    const id = document.getElementById('entitySelect').value;
    if (!id) { document.getElementById('detailsSection').style.display='none'; return; }
    const resp = await fetch('/api/entities');
    const data = await resp.json();
    const entity = (data.entities || []).find(e => e.ID === id);
    if (!entity) return;
    const relResp = await fetch('/api/relationships?entity=' + id);
    const relData = await relResp.json();
    let html = '<div class="result-card"><div class="result-name">' + entity.Name + '</div>';
    html += '<div class="result-desc">' + entity.Description + '</div>';
    html += '<div class="result-meta"><span class="badge badge-type">' + entity.Type + '</span><span class="badge badge-tier">' + entity.Tier + '</span></div>';
    if (relData.relationships && relData.relationships.length > 0) {
        html += '<div style="margin-top:15px;"><strong>关系:</strong><ul>';
        relData.relationships.forEach(r => { html += '<li>' + r.FromEntity + ' --[' + r.Type + ']--> ' + r.ToEntity + ': ' + r.Description + '</li>'; });
        html += '</ul></div>';
    }
    html += '</div>';
    document.getElementById('entityDetails').innerHTML = html;
    document.getElementById('detailsSection').style.display = 'block';
}
function showResults(entities) {
    const container = document.getElementById('resultsContainer');
    if (entities.length === 0) { container.innerHTML = '<div class="no-results">未找到匹配的实体</div>'; }
    else {
        let html = '';
        entities.forEach(e => {
            html += '<div class="result-card"><div class="result-name">' + e.Name + '</div><div class="result-desc">' + e.Description + '</div><div class="result-meta"><span class="badge badge-type">' + e.Type + '</span><span class="badge badge-tier">' + e.Tier + '</span></div></div>';
        });
        container.innerHTML = html;
    }
    document.getElementById('resultsSection').style.display = 'block';
}
</script>
</body></html>`)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleRelationshipsPage(w http.ResponseWriter, r *http.Request) {
	rels, _ := index.GetAllRelationships()
	var rows strings.Builder
	for _, rel := range rels {
		rows.WriteString(fmt.Sprintf(`<tr><td>%s</td><td><span class="rel-type">%s</span></td><td>%s</td><td>%s</td><td><button onclick="deleteRel(%d)" class="btn-delete">🗑️</button></td></tr>`, rel.FromEntity, rel.Type, rel.ToEntity, rel.Description, rel.ID))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>关系图谱 - Awesome WWE</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#fff;min-height:100vh}
.container{max-width:1400px;margin:0 auto;padding:30px}
header{display:flex;justify-content:space-between;align-items:center;padding:30px 0;border-bottom:1px solid rgba(255,255,255,0.1);margin-bottom:30px}
h1{color:#e94560;font-size:2.2em}.btn{padding:12px 25px;background:#e94560;color:#fff;text-decoration:none;border-radius:25px;font-weight:bold;border:none;cursor:pointer}
.btn:hover{transform:translateY(-3px)}
.stats-row{display:flex;gap:30px;margin-bottom:30px;flex-wrap:wrap}
.stat-item{background:rgba(255,255,255,0.05);padding:20px 30px;border-radius:12px}
.stat-value{font-size:2em;color:#e94560;font-weight:bold}
.stat-label{color:#888;font-size:0.9em}
table{width:100%%;border-collapse:collapse;background:rgba(255,255,255,0.03);border-radius:15px;overflow:hidden}
th{background:rgba(233,69,96,0.2);padding:18px;color:#e94560;font-weight:bold;text-align:left}
td{padding:18px;border-bottom:1px solid rgba(255,255,255,0.05)}
tr:hover{background:rgba(255,255,255,0.05)}
.rel-type{padding:5px 15px;background:#4a90d9;border-radius:20px;font-size:0.85em}
.btn-delete{padding:8px 15px;background:#dc3545;color:#fff;border:none;border-radius:20px;cursor:pointer}
.btn-delete:hover{background:#c82333}
.empty{text-align:center;padding:60px;color:#888}
</style></head>
<body>
<div class="container">
<header><h1>🌐 关系图谱</h1>
<div style="display:flex;gap:15px;">
<a href="/add-relationship" class="btn">➕ 添加关系</a>
<a href="/" class="btn">🏠 返回首页</a>
</div></header>

<div class="stats-row">
<div class="stat-item"><div class="stat-value">%d</div><div class="stat-label">总关系数</div></div>
</div>

%s

<script>
async function deleteRel(id) { if(!confirm('确定删除?'))return; const resp=await fetch('/api/relationships/delete?id='+id,{method:'POST'}); const data=await resp.json(); if(data.success)location.reload(); }
</script>
</body></html>`, len(rels), formatRelationshipsTable(rows.String()))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func formatRelationshipsTable(content string) string {
	if content == "" {
		return "<div class='empty'>暂无关系数据 <a href='/add-relationship' style='color:#e94560'>添加第一</div>"
	}
	return "<table><thead><tr><th>起始实体</th><th>关系类型</th><th>目标实体</th><th>描述</th><th>操作</th></tr></thead><tbody>" + content + "</tbody></table>"
}

func (s *Server) handleGeneratePage(w http.ResponseWriter, r *http.Request) {
	entities, _ := index.ListEntities()
	var opts strings.Builder
	for _, e := range entities {
		opts.WriteString(fmt.Sprintf(`<option value="%s">%s (%s)</option>`, e.ID, e.Name, e.Type))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>生成剧本 - Awesome WWE</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#fff;min-height:100vh}
.container{max-width:900px;margin:0 auto;padding:30px}
header{display:flex;justify-content:space-between;align-items:center;padding:30px 0;border-bottom:1px solid rgba(255,255,255,0.1);margin-bottom:30px}
h1{color:#e94560;font-size:2.2em}.back-btn{padding:12px 25px;background:#e94560;color:#fff;text-decoration:none;border-radius:25px}
.form-card{background:rgba(255,255,255,0.03);border-radius:20px;padding:40px;border:1px solid rgba(255,255,255,0.1)}
.form-group{margin-bottom:25px}label{display:block;margin-bottom:10px;color:#e94560;font-weight:bold;font-size:1.1em}
input,select,textarea{width:100%%;padding:15px 20px;background:rgba(255,255,255,0.1);border:1px solid rgba(255,255,255,0.2);border-radius:12px;color:#fff;font-size:1em}
input:focus,select:focus,textarea:focus{outline:none;border-color:#e94560;box-shadow:0 0 20px rgba(233,69,96,0.3)}
textarea{min-height:150px;resize:vertical}select option{background:#1a1a2e}
.generate-btn{width:100%%;padding:20px;background:linear-gradient(135deg,#e94560,#c73e54);color:#fff;border:none;border-radius:12px;font-size:1.3em;font-weight:bold;cursor:pointer;transition:all 0.3s}
.generate-btn:hover{transform:translateY(-5px);box-shadow:0 20px 50px rgba(233,69,96,0.5)}
.result-box{margin-top:30px;padding:30px;background:rgba(255,255,255,0.05);border-radius:15px;white-space:pre-wrap;line-height:1.8}
.loading{text-align:center;padding:30px;color:#888}
</style></head>
<body>
<div class="container">
<header><h1>✨ 生成剧本</h1><a href="/" class="back-btn">← 返回首页</a></header>

<div class="form-card">
<div class="form-group">
<label>场景描述 *</label>
<input type="text" id="sceneInput" placeholder="例如: Raw第100期 Stone Cold vs The Rock 冠军争夺战">
</div>
<div class="form-group">
<label>选择角色 (可多选)</label>
<select id="characters" multiple style="height:180px;">
%s
</select>
<small style="color:#888;margin-top:5px;display:block;">按住 Ctrl 可多选</small>
</div>
<div class="form-group">
<label>剧情要求</label>
<textarea id="requirements" placeholder="描述你想要的剧情走向、对手关系、结局等..."></textarea>
</div>
<button type="button" class="generate-btn" onclick="generateScript()">✨ 开始生成剧本</button>
</div>

<div id="resultBox" class="result-box" style="display:none;"></div>
<div id="loading" class="loading" style="display:none;">正在生成，请稍候...</div>
</div>

<script>
async function generateScript() {
    const scene = document.getElementById('sceneInput').value;
    if (!scene) { alert('请输入场景描述'); return; }
    const selects = document.getElementById('characters').selectedOptions;
    const chars = Array.from(selects).map(o => o.value).join(',');
    const req = document.getElementById('requirements').value;
    document.getElementById('loading').style.display = 'block';
    document.getElementById('resultBox').style.display = 'none';
    try {
        const resp = await fetch('/api/generate?scene=' + encodeURIComponent(scene) + '&chars=' + encodeURIComponent(chars) + '&req=' + encodeURIComponent(req));
        const data = await resp.json();
        document.getElementById('loading').style.display = 'none';
        document.getElementById('resultBox').style.display = 'block';
        document.getElementById('resultBox').textContent = data.content || JSON.stringify(data, null, 2);
    } catch(e) {
        document.getElementById('loading').style.display = 'none';
        alert('生成失败: ' + e.message);
    }
}
</script>
</body></html>`, opts.String())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
