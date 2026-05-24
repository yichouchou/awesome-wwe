package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/yichouchou/awesome-wwe/internal/index"
	"github.com/yichouchou/awesome-wwe/internal/llm"
)

// Agent handles WWE script generation
type Agent struct {
	client  *llm.Client
	scene   string
}

// Option configures the agent
type Option func(*Agent)

// WithScene sets the scene for generation
func WithScene(scene string) Option {
	return func(a *Agent) {
		a.scene = scene
	}
}

// NewAgent creates a new WWE script agent
func NewAgent(opts ...Option) (*Agent, error) {
	cfg := llm.DefaultConfig()

	client, err := llm.NewClient(llm.Config{
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
		ModelName: cfg.ModelName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	agent := &Agent{
		client: client,
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent, nil
}

// Generate generates a WWE script based on the scene
func (a *Agent) Generate() (string, error) {
	ctx := context.Background()

	// Build prompt with context
	prompt := a.buildPrompt()

	// Call LLM
	return a.client.Generate(ctx, systemPrompt, prompt)
}

// buildPrompt builds the generation prompt
func (a *Agent) buildPrompt() string {
	var sb strings.Builder

	sb.WriteString("生成 WWE 摔角剧本。\n\n")

	if a.scene != "" {
		sb.WriteString("场景: " + a.scene + "\n")
	}

	// Get relevant entities from index
	entities, _ := index.ListEntities()
	if len(entities) > 0 {
		sb.WriteString("\n当前角色库:\n")
		for _, e := range entities {
			if e.Tier == "核心" || e.Tier == "重要" {
				sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", e.Name, e.Type, e.Description))
			}
		}
	}

	sb.WriteString("\n请生成一段精彩的 WWE 剧本对白和场景描述。")
	return sb.String()
}

// systemPrompt is the system prompt for the agent
const systemPrompt = `你是一个专业的 WWE 摔角剧本写作助手。

你的职责：
1. 生成符合 WWE 摔角风格的剧本对白
2. 角色台词要符合其人设和性格
3. 场景描述要生动、有戏剧性
4. 合理运用摔角术语和 WWE 文化元素

输出格式：
- 场景标题
- 角色对话（带角色名）
- 场景动作描述
- 剧情发展

注意：
- 保持角色一致性
- 对话要简洁有力，符合摔角风格
- 适当加入挑衅和戏剧性冲突`