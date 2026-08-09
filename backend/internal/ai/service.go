package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/riccardo/freezo/backend/internal/db"
	"github.com/riccardo/freezo/backend/internal/mcp"
)

type Config struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

type AIProcessRequest struct {
	Prompt string `json:"prompt"`
}

type AIExecuteRequest struct {
	Actions []ActionResult `json:"actions"`
}

type ActionResult struct {
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
	Summary  string                 `json:"summary,omitempty"`
	Output   string                 `json:"output,omitempty"`
}

type AIProcessResponse struct {
	Message             string         `json:"message"`
	Actions             []ActionResult `json:"actions"`
	PendingConfirmation bool           `json:"pending_confirmation"`
}

// OpenAI API Request / Response Types
type OpenAITool struct {
	Type     string      `json:"type"`
	Function OpenAIFunc  `json:"function"`
}

type OpenAIFunc struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type OpenAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Tools       []OpenAITool    `json:"tools,omitempty"`
	ToolChoice  string          `json:"tool_choice,omitempty"`
	Temperature float64         `json:"temperature"`
	Think       *bool           `json:"think,omitempty"`
}

type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type OpenAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type OpenAIChatResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func GetConfig() Config {
	baseURL, _ := db.GetSetting("ai_base_url")
	if baseURL == "" {
		baseURL = os.Getenv("LLM_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	apiKey, _ := db.GetSetting("ai_api_key")
	if apiKey == "" {
		apiKey = os.Getenv("LLM_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	model, _ := db.GetSetting("ai_model")
	if model == "" {
		model = os.Getenv("LLM_MODEL")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	return Config{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
	}
}

func UpdateConfig(cfg Config) error {
	if err := db.SetSetting("ai_base_url", strings.TrimRight(cfg.BaseURL, "/")); err != nil {
		return err
	}
	if err := db.SetSetting("ai_api_key", cfg.APIKey); err != nil {
		return err
	}
	return db.SetSetting("ai_model", cfg.Model)
}

func parseCount(v interface{}) int {
	if v == nil {
		return 1
	}
	switch val := v.(type) {
	case float64:
		if val > 0 {
			return int(val)
		}
	case int:
		if val > 0 {
			return val
		}
	case string:
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 1
}

func generateActionSummary(toolName string, args map[string]interface{}) string {
	if args == nil {
		return fmt.Sprintf("Execute %s", toolName)
	}

	switch toolName {
	case "add_items":
		var summaries []string
		if items, ok := args["items"].([]interface{}); ok && len(items) > 0 {
			for _, itemRaw := range items {
				if itemMap, ok := itemRaw.(map[string]interface{}); ok {
					name, _ := itemMap["name"].(string)
					freezer, _ := itemMap["freezer_name"].(string)
					category, _ := itemMap["category_name"].(string)
					weight, _ := itemMap["weight"].(string)
					count := parseCount(itemMap["count"])

					desc := fmt.Sprintf("%dx %s", count, name)
					if weight != "" {
						desc += fmt.Sprintf(" (%s)", weight)
					}
					if freezer != "" {
						desc += fmt.Sprintf(" -> %s", freezer)
					}
					if category != "" {
						desc += fmt.Sprintf(" [%s]", category)
					}
					summaries = append(summaries, desc)
				}
			}
		} else if name, ok := args["name"].(string); ok && name != "" {
			freezer, _ := args["freezer_name"].(string)
			category, _ := args["category_name"].(string)
			count := parseCount(args["count"])
			desc := fmt.Sprintf("%dx %s", count, name)
			if freezer != "" {
				desc += fmt.Sprintf(" -> %s", freezer)
			}
			if category != "" {
				desc += fmt.Sprintf(" [%s]", category)
			}
			summaries = append(summaries, desc)
		}

		if len(summaries) > 0 {
			return "Add " + strings.Join(summaries, ", ")
		}
		return "Add items to inventory"

	case "consume_items":
		name, _ := args["item_name"].(string)
		if name == "" {
			name, _ = args["name"].(string)
		}
		freezer, _ := args["freezer_name"].(string)
		count := parseCount(args["count"])
		if name != "" {
			desc := fmt.Sprintf("Consume %dx %s", count, name)
			if freezer != "" {
				desc += fmt.Sprintf(" from %s", freezer)
			}
			return desc
		}
		return "Consume items from inventory"

	case "move_items":
		name, _ := args["item_name"].(string)
		if name == "" {
			name, _ = args["name"].(string)
		}
		targetFreezer, _ := args["target_freezer_name"].(string)
		if targetFreezer == "" {
			targetFreezer, _ = args["freezer_name"].(string)
		}
		if name != "" || targetFreezer != "" {
			return fmt.Sprintf("Move %s -> %s", name, targetFreezer)
		}
		return "Move items between freezers"

	case "update_items_category":
		name, _ := args["item_name"].(string)
		if name == "" {
			name, _ = args["name"].(string)
		}
		cat, _ := args["category_name"].(string)
		if cat == "" {
			cat, _ = args["category_id"].(string)
		}
		if name != "" || cat != "" {
			return fmt.Sprintf("Recategorize %s -> %s", name, cat)
		}
		return "Update category of items"

	default:
		return fmt.Sprintf("Execute %s", toolName)
	}
}

func ProcessPrompt(prompt string) (*AIProcessResponse, error) {
	cfg := GetConfig()

	// 1. Fetch current freezers, categories, and items as context
	freezersJSON, _ := mcp.ExecuteTool("list_freezers", nil)
	categoriesJSON, _ := mcp.ExecuteTool("list_categories", nil)
	itemsJSON, _ := mcp.ExecuteTool("list_items", nil)

	systemPrompt := fmt.Sprintf(`You are the Freezo AI Assistant for freezer inventory management.
Your job is to interpret user natural language commands and call appropriate tool functions to update the food inventory in Freezo.

CURRENT FREEZO CONTEXT:
- Freezers: %s
- Categories: %s
- Current Items: %s

INSTRUCTIONS:
- Use tool calls (add_items, consume_items, move_items, update_items_category) to modify inventory matching user intent.
- Match freezer names and category names from current context when possible.
- If quantity/count is specified (e.g., 3 steaks), pass count: 3 in the tool call.
- Always pick the most appropriate category (e.g. Pork, Seafood, Beef, Poultry, Bread, or Uncategorized).
- If the user asks a question about inventory without changing anything, call list_items or answer directly.`, freezersJSON, categoriesJSON, itemsJSON)

	// Prepare OpenAI tools from MCP definitions
	mcpTools := mcp.GetToolDefinitions()
	var openAITools []OpenAITool
	for _, t := range mcpTools {
		openAITools = append(openAITools, OpenAITool{
			Type: "function",
			Function: OpenAIFunc{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	falseBool := false
	reqBody := OpenAIChatRequest{
		Model: cfg.Model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Tools:       openAITools,
		Temperature: 0.1,
		Think:       &falseBool,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", cfg.BaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))
	}

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read LLM response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM endpoint returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var chatResp OpenAIChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response JSON: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("LLM API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned by LLM")
	}

	choiceMessage := chatResp.Choices[0].Message
	var proposedActions []ActionResult

	// Collect returned tool calls as PROPOSED actions (do not execute yet!)
	for _, toolCall := range choiceMessage.ToolCalls {
		var args map[string]interface{}
		_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

		summary := generateActionSummary(toolCall.Function.Name, args)

		proposedActions = append(proposedActions, ActionResult{
			ToolName: toolCall.Function.Name,
			Args:     args,
			Summary:  summary,
		})
	}

	responseMessage := choiceMessage.Content
	pendingConfirmation := len(proposedActions) > 0

	if responseMessage == "" {
		if pendingConfirmation {
			responseMessage = fmt.Sprintf("I propose the following %d change(s) to your inventory. Please confirm to execute:", len(proposedActions))
		} else {
			responseMessage = "No inventory changes were proposed."
		}
	}

	return &AIProcessResponse{
		Message:             responseMessage,
		Actions:             proposedActions,
		PendingConfirmation: pendingConfirmation,
	}, nil
}

func ExecuteActions(actions []ActionResult) (*AIProcessResponse, error) {
	var executedActions []ActionResult

	for _, act := range actions {
		output, err := mcp.ExecuteTool(act.ToolName, act.Args)
		if err != nil {
			output = fmt.Sprintf("Error: %v", err)
		}

		summary := act.Summary
		if summary == "" {
			summary = generateActionSummary(act.ToolName, act.Args)
		}

		executedActions = append(executedActions, ActionResult{
			ToolName: act.ToolName,
			Args:     act.Args,
			Summary:  summary,
			Output:   output,
		})
	}

	return &AIProcessResponse{
		Message:             fmt.Sprintf("Successfully executed %d action(s) for your inventory.", len(executedActions)),
		Actions:             executedActions,
		PendingConfirmation: false,
	}, nil
}
