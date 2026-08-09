import React, { useState, useEffect } from 'react';
import { Modal } from './Modal';
import { Sparkles, Settings, Server, Send, Loader2, CheckCircle2, AlertCircle, Terminal, Check, X, ShieldAlert } from 'lucide-react';
import { api, type AIConfig, type AIProcessResponse } from '../api';

interface AIAssistantModalProps {
    isOpen: boolean;
    onClose: () => void;
    onInventoryChanged: () => void;
}

export function AIAssistantModal({ isOpen, onClose, onInventoryChanged }: AIAssistantModalProps) {
    const [activeTab, setActiveTab] = useState<'assistant' | 'settings' | 'mcp'>('assistant');
    const [prompt, setPrompt] = useState('');
    const [loading, setLoading] = useState(false);
    const [executing, setExecuting] = useState(false);
    const [response, setResponse] = useState<AIProcessResponse | null>(null);
    const [error, setError] = useState<string | null>(null);

    // Settings state
    const [config, setConfig] = useState<AIConfig>({
        base_url: 'https://api.openai.com/v1',
        api_key: '',
        model: 'gpt-4o-mini',
    });
    const [savingConfig, setSavingConfig] = useState(false);
    const [configMessage, setConfigMessage] = useState<string | null>(null);

    useEffect(() => {
        if (isOpen) {
            fetchConfig();
        }
    }, [isOpen]);

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
        setSavingConfig(true);
        setConfigMessage(null);
        try {
            const updated = await api.updateAIConfig(config);
            setConfig(updated);
            setConfigMessage('Settings saved successfully!');
            setTimeout(() => setConfigMessage(null), 3000);
        } catch (err: any) {
            setConfigMessage(`Failed to save: ${err.message}`);
        } finally {
            setSavingConfig(false);
        }
    };

    const handleProcessPrompt = async (promptToSubmit?: string) => {
        const text = promptToSubmit || prompt;
        if (!text.trim() || loading) return;

        setLoading(true);
        setError(null);
        setResponse(null);

        try {
            const result = await api.processAIPrompt(text);
            setResponse(result);
        } catch (err: any) {
            setError(err.message || 'Failed to process natural language command');
        } finally {
            setLoading(false);
        }
    };

    const handleConfirmActions = async () => {
        if (!response || !response.actions || response.actions.length === 0 || executing) return;

        setExecuting(true);
        setError(null);

        try {
            const result = await api.executeAIActions(response.actions);
            setResponse(result);
            onInventoryChanged();
        } catch (err: any) {
            setError(err.message || 'Failed to execute proposed actions');
        } finally {
            setExecuting(false);
        }
    };

    const handleCancelActions = () => {
        setResponse(null);
        setError(null);
    };

    const promptSuggestions = [
        "Added 3 Ribeye Steaks to Kitchen Freezer under Beef",
        "Ate 2 Chicken Breasts from Garage Freezer",
        "Move all Poultry from Kitchen Freezer to Garage Deep Freeze",
        "Change category of Pork Chops to Pork",
    ];

    const mcpServerUrl = `${window.location.protocol}//${window.location.host}/api/mcp`;

    return (
        <Modal isOpen={isOpen} onClose={onClose} title="AI Inventory Assistant">
            <div className="space-y-4">
                {/* Tab selector */}
                <div className="flex border-b border-gray-100 gap-2 pb-2">
                    <button
                        onClick={() => setActiveTab('assistant')}
                        className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                            activeTab === 'assistant'
                                ? 'bg-indigo-50 text-indigo-600'
                                : 'text-gray-600 hover:bg-gray-50'
                        }`}
                    >
                        <Sparkles className="w-4 h-4" />
                        Assistant
                    </button>
                    <button
                        onClick={() => setActiveTab('settings')}
                        className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                            activeTab === 'settings'
                                ? 'bg-indigo-50 text-indigo-600'
                                : 'text-gray-600 hover:bg-gray-50'
                        }`}
                    >
                        <Settings className="w-4 h-4" />
                        LLM Settings
                    </button>
                    <button
                        onClick={() => setActiveTab('mcp')}
                        className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                            activeTab === 'mcp'
                                ? 'bg-indigo-50 text-indigo-600'
                                : 'text-gray-600 hover:bg-gray-50'
                        }`}
                    >
                        <Server className="w-4 h-4" />
                        MCP Server
                    </button>
                </div>

                {/* Assistant Tab */}
                {activeTab === 'assistant' && (
                    <div className="space-y-4">
                        <p className="text-xs text-gray-500">
                            Type what food items you bought, cooked, moved, or recategorized in plain English.
                        </p>

                        <div className="relative">
                            <textarea
                                value={prompt}
                                onChange={(e) => setPrompt(e.target.value)}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' && !e.shiftKey) {
                                        e.preventDefault();
                                        handleProcessPrompt();
                                    }
                                }}
                                placeholder="e.g. Added 4 packs of bacon (1 lb each) to Kitchen Freezer under Pork"
                                className="w-full h-24 p-3 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
                            />
                            <button
                                onClick={() => handleProcessPrompt()}
                                disabled={loading || executing || !prompt.trim()}
                                className="absolute right-3 bottom-3 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-xs font-medium flex items-center gap-1 transition-colors"
                            >
                                {loading ? (
                                    <>
                                        <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                        Thinking...
                                    </>
                                ) : (
                                    <>
                                        <Send className="w-3.5 h-3.5" />
                                        Submit
                                    </>
                                )}
                            </button>
                        </div>

                        {/* Suggestions */}
                        <div>
                            <span className="text-xs font-medium text-gray-400 block mb-1.5">Try an example:</span>
                            <div className="flex flex-wrap gap-1.5">
                                {promptSuggestions.map((s, idx) => (
                                    <button
                                        key={idx}
                                        onClick={() => {
                                            setPrompt(s);
                                            handleProcessPrompt(s);
                                        }}
                                        className="text-xs bg-gray-50 hover:bg-indigo-50 hover:text-indigo-600 text-gray-600 px-2.5 py-1 rounded-md border border-gray-200/60 transition-colors text-left"
                                    >
                                        {s}
                                    </button>
                                ))}
                            </div>
                        </div>

                        {/* Error Alert */}
                        {error && (
                            <div className="p-3 bg-red-50 border border-red-100 text-red-700 rounded-xl text-xs flex items-start gap-2">
                                <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
                                <span>{error}</span>
                            </div>
                        )}

                        {/* Confirmation Required Box */}
                        {response && response.pending_confirmation && response.actions && response.actions.length > 0 && (
                            <div className="p-4 bg-amber-50/80 border border-amber-200 rounded-xl space-y-3">
                                <div className="flex items-start gap-2">
                                    <ShieldAlert className="w-4.5 h-4.5 text-amber-600 shrink-0 mt-0.5" />
                                    <div>
                                        <h4 className="text-xs font-bold text-amber-900">Action Confirmation Required</h4>
                                        <p className="text-xs text-amber-700 mt-0.5">{response.message}</p>
                                    </div>
                                </div>

                                <div className="space-y-2 pt-2 border-t border-amber-200/60">
                                    <span className="text-[11px] font-semibold uppercase tracking-wider text-amber-800">Proposed Changes:</span>
                                    {response.actions.map((act, i) => {
                                        const details: { label: string; value: string }[] = [];
                                        const args = act.args || {};

                                        if (args.items && Array.isArray(args.items)) {
                                            args.items.forEach((it: any) => {
                                                if (it.name) details.push({ label: 'Item', value: it.name });
                                                if (it.freezer_name) details.push({ label: 'Freezer', value: it.freezer_name });
                                                if (it.category_name) details.push({ label: 'Category', value: it.category_name });
                                                if (it.count) details.push({ label: 'Quantity', value: String(it.count) });
                                                if (it.weight) details.push({ label: 'Weight', value: it.weight });
                                            });
                                        } else {
                                            if (args.item_name || args.name) details.push({ label: 'Item', value: args.item_name || args.name });
                                            if (args.freezer_name || args.target_freezer_name) details.push({ label: 'Freezer', value: args.freezer_name || args.target_freezer_name });
                                            if (args.category_name) details.push({ label: 'Category', value: args.category_name });
                                            if (args.count) details.push({ label: 'Quantity', value: String(args.count) });
                                            if (args.weight) details.push({ label: 'Weight', value: args.weight });
                                        }

                                        return (
                                            <div key={i} className="text-xs bg-white p-3 rounded-lg border border-amber-200 text-gray-800 font-medium space-y-1">
                                                <div className="flex items-center justify-between">
                                                    <span className="font-semibold text-gray-900">{act.summary || `${act.tool_name}`}</span>
                                                    <span className="text-[10px] bg-amber-100 text-amber-800 px-2 py-0.5 rounded font-mono">{act.tool_name}</span>
                                                </div>
                                                {details.length > 0 && (
                                                    <div className="flex flex-wrap gap-1.5 pt-1">
                                                        {details.map((d, idx) => (
                                                            <span key={idx} className="text-[11px] bg-amber-50 text-amber-900 px-2 py-0.5 rounded-md font-medium border border-amber-200/60">
                                                                <span className="text-amber-700 font-semibold">{d.label}:</span> {d.value}
                                                            </span>
                                                        ))}
                                                    </div>
                                                )}
                                            </div>
                                        );
                                    })}
                                </div>

                                <div className="flex justify-end gap-2 pt-2">
                                    <button
                                        onClick={handleCancelActions}
                                        disabled={executing}
                                        className="px-3 py-1.5 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-xs font-medium flex items-center gap-1 transition-colors"
                                    >
                                        <X className="w-3.5 h-3.5" />
                                        Cancel
                                    </button>
                                    <button
                                        onClick={handleConfirmActions}
                                        disabled={executing}
                                        className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-medium flex items-center gap-1.5 transition-colors shadow-sm"
                                    >
                                        {executing ? (
                                            <>
                                                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                                Executing...
                                            </>
                                        ) : (
                                            <>
                                                <Check className="w-3.5 h-3.5" />
                                                Confirm & Execute
                                            </>
                                        )}
                                    </button>
                                </div>
                            </div>
                        )}

                        {/* Completed Response Result Card */}
                        {response && !response.pending_confirmation && (
                            <div className="p-4 bg-indigo-50/50 border border-indigo-100 rounded-xl space-y-3">
                                <div className="flex items-start gap-2">
                                    <CheckCircle2 className="w-4 h-4 text-indigo-600 shrink-0 mt-0.5" />
                                    <div className="text-xs text-gray-800 font-medium">
                                        {response.message}
                                    </div>
                                </div>

                                {response.actions && response.actions.length > 0 && (
                                    <div className="space-y-1.5 pt-2 border-t border-indigo-100/60">
                                        <span className="text-[11px] font-semibold uppercase tracking-wider text-indigo-700">Executed Actions:</span>
                                        {response.actions.map((act, i) => (
                                            <div key={i} className="text-xs bg-white p-2 rounded border border-indigo-100 text-gray-700">
                                                <div className="font-semibold text-indigo-700">{act.summary || act.tool_name}</div>
                                                {act.output && <div className="text-[10px] text-gray-500 font-mono mt-0.5">{act.output}</div>}
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                )}

                {/* Settings Tab */}
                {activeTab === 'settings' && (
                    <form onSubmit={handleSaveConfig} className="space-y-3">
                        <p className="text-xs text-gray-500">
                            Configure your LLM provider endpoint. Compatible with OpenAI, Ollama, OpenRouter, vLLM, or self-hosted OpenAI-compatible endpoints.
                        </p>

                        <div>
                            <label className="block text-xs font-medium text-gray-700 mb-1">API Base URL</label>
                            <input
                                type="text"
                                value={config.base_url}
                                onChange={(e) => setConfig({ ...config, base_url: e.target.value })}
                                placeholder="https://api.openai.com/v1"
                                className="w-full p-2 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                            />
                            <span className="text-[10px] text-gray-400">e.g., https://api.openai.com/v1 or http://localhost:11434/v1</span>
                        </div>

                        <div>
                            <label className="block text-xs font-medium text-gray-700 mb-1">Model Name</label>
                            <input
                                type="text"
                                value={config.model}
                                onChange={(e) => setConfig({ ...config, model: e.target.value })}
                                placeholder="gpt-4o-mini"
                                className="w-full p-2 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                            />
                            <span className="text-[10px] text-gray-400">e.g., gpt-4o-mini, llama3.2, claude-3-5-sonnet</span>
                        </div>

                        <div>
                            <label className="block text-xs font-medium text-gray-700 mb-1">API Key (Optional for cluster LLMs)</label>
                            <input
                                type="password"
                                value={config.api_key}
                                onChange={(e) => setConfig({ ...config, api_key: e.target.value })}
                                placeholder="sk-..."
                                className="w-full p-2 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                            />
                        </div>

                        {configMessage && (
                            <div className="text-xs text-indigo-600 font-medium p-2 bg-indigo-50 rounded-lg">
                                {configMessage}
                            </div>
                        )}

                        <button
                            type="submit"
                            disabled={savingConfig}
                            className="w-full py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-medium transition-colors"
                        >
                            {savingConfig ? 'Saving...' : 'Save Settings'}
                        </button>
                    </form>
                )}

                {/* MCP Server Tab */}
                {activeTab === 'mcp' && (
                    <div className="space-y-3">
                        <div className="p-3 bg-gray-50 border border-gray-200 rounded-xl space-y-2">
                            <div className="flex items-center gap-2 text-xs font-semibold text-gray-800">
                                <Terminal className="w-4 h-4 text-indigo-600" />
                                Model Context Protocol (MCP) Endpoint
                            </div>
                            <p className="text-xs text-gray-600">
                                Freezo hosts an MCP server to let external AI agents (like Claude Desktop, Antigravity, or Cursor) manage your freezer inventory.
                            </p>
                            <div className="bg-gray-900 text-green-400 p-2.5 rounded-lg text-xs font-mono select-all break-all">
                                {mcpServerUrl}
                            </div>
                        </div>

                        <div>
                            <span className="text-xs font-medium text-gray-700 block mb-1">Example Config (`claude_desktop_config.json`):</span>
                            <pre className="p-3 bg-gray-900 text-gray-200 rounded-xl text-[11px] font-mono overflow-x-auto">
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
                )}
            </div>
        </Modal>
    );
}
