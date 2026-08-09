import { useState, useEffect } from 'react';
import { Sparkles, Terminal, Copy, Check, Save, Cpu } from 'lucide-react';
import { api, type AIConfig } from '../api';

export function AISettingsSection() {
  const [config, setConfig] = useState<AIConfig>({
    base_url: 'https://api.openai.com/v1',
    api_key: '',
    model: 'gpt-4o-mini',
  });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    fetchConfig();
  }, []);

  const fetchConfig = async () => {
    try {
      const cfg = await api.getAIConfig();
      setConfig(cfg);
    } catch (err) {
      console.error('Failed to load AI config:', err);
    }
  };

  const handleSaveConfig = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setMessage(null);
    try {
      const updated = await api.updateAIConfig(config);
      setConfig(updated);
      setMessage('AI Provider settings saved successfully!');
      setTimeout(() => setMessage(null), 3500);
    } catch (err: any) {
      setMessage(`Failed to save: ${err.message}`);
    } finally {
      setSaving(false);
    }
  };

  const mcpServerUrl = typeof window !== 'undefined'
    ? `${window.location.protocol}//${window.location.host}/api/mcp`
    : '/api/mcp';

  const handleCopyMcpUrl = () => {
    navigator.clipboard.writeText(mcpServerUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="pt-8 border-t border-gray-200 mt-8 space-y-6">
      <div>
        <div className="flex items-center gap-2">
          <div className="p-1.5 bg-indigo-50 rounded-lg text-indigo-600">
            <Sparkles className="w-5 h-5" />
          </div>
          <h2 className="text-lg font-semibold text-gray-800">AI Assistant & MCP Server</h2>
        </div>
        <p className="text-sm text-gray-500 mt-1 max-w-2xl">
          Configure the LLM provider for the built-in AI Assistant or connect external agents to Freezo via the Model Context Protocol (MCP).
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* LLM Provider Configuration */}
        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4 flex flex-col justify-between">
          <div>
            <div className="flex items-center gap-2 mb-3">
              <Cpu className="w-4 h-4 text-indigo-600" />
              <h3 className="font-semibold text-gray-900 text-sm">LLM Provider Configuration</h3>
            </div>
            <p className="text-xs text-gray-500 mb-4">
              Compatible with OpenAI, Ollama, OpenRouter, vLLM, or any self-hosted OpenAI-compatible endpoint.
            </p>

            <form id="ai-config-form" onSubmit={handleSaveConfig} className="space-y-3.5">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">API Base URL</label>
                <input
                  type="text"
                  value={config.base_url}
                  onChange={(e) => setConfig({ ...config, base_url: e.target.value })}
                  placeholder="https://api.openai.com/v1"
                  className="w-full p-2.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition-all"
                />
                <span className="text-[10px] text-gray-400 mt-1 block">e.g., https://api.openai.com/v1 or http://localhost:11434/v1</span>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Model Name</label>
                <input
                  type="text"
                  value={config.model}
                  onChange={(e) => setConfig({ ...config, model: e.target.value })}
                  placeholder="gpt-4o-mini"
                  className="w-full p-2.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition-all"
                />
                <span className="text-[10px] text-gray-400 mt-1 block">e.g., gpt-4o-mini, llama3.2, claude-3-5-sonnet</span>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">API Key (Optional for local cluster models)</label>
                <input
                  type="password"
                  value={config.api_key}
                  onChange={(e) => setConfig({ ...config, api_key: e.target.value })}
                  placeholder="sk-..."
                  className="w-full p-2.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition-all"
                />
              </div>

              {message && (
                <div className={`text-xs p-2.5 rounded-lg font-medium flex items-center gap-1.5 ${
                  message.startsWith('Failed') ? 'bg-red-50 text-red-700 border border-red-100' : 'bg-indigo-50 text-indigo-700 border border-indigo-100'
                }`}>
                  {message.startsWith('Failed') ? null : <Check className="w-3.5 h-3.5" />}
                  {message}
                </div>
              )}
            </form>
          </div>

          <div className="pt-3 border-t border-gray-100">
            <button
              type="submit"
              form="ai-config-form"
              disabled={saving}
              className="w-full py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-xs font-medium flex items-center justify-center gap-1.5 transition-colors shadow-xs cursor-pointer"
            >
              <Save className="w-3.5 h-3.5" />
              {saving ? 'Saving Settings...' : 'Save LLM Settings'}
            </button>
          </div>
        </div>

        {/* MCP Server Details */}
        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4 flex flex-col justify-between">
          <div>
            <div className="flex items-center gap-2 mb-3">
              <Terminal className="w-4 h-4 text-indigo-600" />
              <h3 className="font-semibold text-gray-900 text-sm">Model Context Protocol (MCP) Server</h3>
            </div>
            <p className="text-xs text-gray-500 mb-3">
              Freezo exposes an MCP server endpoint. External AI environments (Claude Desktop, Antigravity, Cursor) can query and modify your freezer inventory directly.
            </p>

            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">MCP Endpoint URL</label>
                <div className="flex items-center gap-2">
                  <div className="bg-gray-900 text-green-400 p-2.5 rounded-lg text-xs font-mono select-all flex-1 truncate">
                    {mcpServerUrl}
                  </div>
                  <button
                    type="button"
                    onClick={handleCopyMcpUrl}
                    className="p-2.5 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-xs transition-colors shrink-0 cursor-pointer"
                    title="Copy URL"
                  >
                    {copied ? <Check className="w-4 h-4 text-green-600" /> : <Copy className="w-4 h-4" />}
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Configuration Example (`claude_desktop_config.json`)</label>
                <pre className="p-3 bg-gray-900 text-gray-200 rounded-xl text-[11px] font-mono overflow-x-auto leading-relaxed">
{`{
  "mcpServers": {
    "freezo": {
      "url": "${mcpServerUrl}"
    }
  }
}`}
                </pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
