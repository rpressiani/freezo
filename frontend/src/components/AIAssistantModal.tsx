import { useState, useEffect, useRef } from 'react';
import { Modal } from './Modal';
import { Sparkles, Send, Loader2, CheckCircle2, AlertCircle, Check, X, ShieldAlert, Settings, Bot, User } from 'lucide-react';
import { api, type AIActionResult } from '../api';

export interface ChatMessage {
    id: string;
    sender: 'user' | 'assistant';
    text?: string;
    timestamp: Date;
    actions?: AIActionResult[];
    pendingConfirmation?: boolean;
    executedActions?: AIActionResult[];
    error?: string;
    executing?: boolean;
}

interface AIAssistantModalProps {
    isOpen: boolean;
    onClose: () => void;
    onInventoryChanged: () => void;
    onOpenSettings?: () => void;
}

export function AIAssistantModal({ isOpen, onClose, onInventoryChanged, onOpenSettings }: AIAssistantModalProps) {
    const [prompt, setPrompt] = useState('');
    const [loading, setLoading] = useState(false);
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const messagesEndRef = useRef<HTMLDivElement>(null);

    // Initialize greeting on open if chat history is empty
    useEffect(() => {
        if (isOpen && messages.length === 0) {
            setMessages([
                {
                    id: 'welcome-msg',
                    sender: 'assistant',
                    text: "Hello! I'm your Freezo AI Assistant. Tell me what items you bought, cooked, moved, or recategorized in plain English.",
                    timestamp: new Date(),
                },
            ]);
        }
    }, [isOpen]);

    // Scroll to bottom when messages update
    useEffect(() => {
        if (isOpen) {
            messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
        }
    }, [messages, loading, isOpen]);

    const handleProcessPrompt = async (promptToSubmit?: string) => {
        const text = (promptToSubmit || prompt).trim();
        if (!text || loading) return;

        const userMsgId = `user-${Date.now()}`;
        const userMsg: ChatMessage = {
            id: userMsgId,
            sender: 'user',
            text: text,
            timestamp: new Date(),
        };

        setMessages((prev) => [...prev, userMsg]);
        setPrompt('');
        setLoading(true);

        try {
            const result = await api.processAIPrompt(text);
            const assistantMsg: ChatMessage = {
                id: `asst-${Date.now()}`,
                sender: 'assistant',
                text: result.message,
                timestamp: new Date(),
                actions: result.actions,
                pendingConfirmation: result.pending_confirmation,
            };
            setMessages((prev) => [...prev, assistantMsg]);
        } catch (err: any) {
            const errorMsg: ChatMessage = {
                id: `err-${Date.now()}`,
                sender: 'assistant',
                error: err.message || 'Failed to process natural language request.',
                timestamp: new Date(),
            };
            setMessages((prev) => [...prev, errorMsg]);
        } finally {
            setLoading(false);
        }
    };

    const handleConfirmActions = async (messageId: string, actionsToExecute: AIActionResult[]) => {
        setMessages((prev) =>
            prev.map((msg) => (msg.id === messageId ? { ...msg, executing: true } : msg))
        );

        try {
            const result = await api.executeAIActions(actionsToExecute);
            setMessages((prev) =>
                prev.map((msg) =>
                    msg.id === messageId
                        ? {
                              ...msg,
                              executing: false,
                              pendingConfirmation: false,
                              text: result.message,
                              executedActions: result.actions,
                          }
                        : msg
                )
            );
            onInventoryChanged();
        } catch (err: any) {
            setMessages((prev) =>
                prev.map((msg) =>
                    msg.id === messageId
                        ? {
                              ...msg,
                              executing: false,
                              error: err.message || 'Failed to execute proposed actions',
                          }
                        : msg
                )
            );
        }
    };

    const handleCancelActions = (messageId: string) => {
        setMessages((prev) =>
            prev.map((msg) =>
                msg.id === messageId
                    ? {
                          ...msg,
                          pendingConfirmation: false,
                          text: 'Action cancelled.',
                      }
                    : msg
            )
        );
    };

    const promptSuggestions = [
        "Added 3 Ribeye Steaks to Kitchen Freezer under Beef",
        "Ate 2 Chicken Breasts from Garage Freezer",
        "Move all Poultry from Kitchen Freezer to Garage Deep Freeze",
        "Change category of Pork Chops to Pork",
    ];

    const modalTitle = (
        <div className="flex items-center gap-2.5">
            <div className="p-1.5 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-lg shadow-xs">
                <Sparkles className="w-5 h-5" />
            </div>
            <div>
                <h2 className="text-lg font-bold text-gray-900 leading-tight">AI Assistant</h2>
                <p className="text-[11px] text-gray-400 font-normal">Natural Language Freezer Management</p>
            </div>
        </div>
    );

    const headerActions = onOpenSettings ? (
        <button
            onClick={() => {
                onClose();
                onOpenSettings();
            }}
            className="flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium text-gray-600 hover:text-indigo-600 hover:bg-indigo-50 border border-gray-200 hover:border-indigo-100 rounded-lg transition-all cursor-pointer mr-1"
            title="Configure LLM & MCP Endpoint"
        >
            <Settings className="w-3.5 h-3.5" />
            <span>Settings</span>
        </button>
    ) : undefined;

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={modalTitle} headerActions={headerActions} maxWidth="max-w-2xl">
            <div className="flex flex-col h-[520px] -mx-6 -my-6">
                {/* Chat Scroll Container */}
                <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-slate-50/50">
                    {messages.map((msg) => {
                        const isUser = msg.sender === 'user';
                        return (
                            <div
                                key={msg.id}
                                className={`flex items-start gap-2.5 ${isUser ? 'flex-row-reverse' : 'flex-row'}`}
                            >
                                {/* Avatar */}
                                <div
                                    className={`w-7 h-7 rounded-full flex items-center justify-center shrink-0 text-xs shadow-xs ${
                                        isUser
                                            ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white'
                                            : 'bg-white text-indigo-600 border border-indigo-100'
                                    }`}
                                >
                                    {isUser ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />}
                                </div>

                                {/* Message Content Container */}
                                <div className={`flex flex-col space-y-1 max-w-[85%] ${isUser ? 'items-end' : 'items-start'}`}>
                                    <div
                                        className={`rounded-2xl px-4 py-3 text-xs leading-relaxed ${
                                            isUser
                                                ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-tr-none shadow-sm font-medium'
                                                : 'bg-white text-gray-800 border border-gray-200/80 rounded-tl-none shadow-xs'
                                        }`}
                                    >
                                        {msg.text && <p className="whitespace-pre-wrap">{msg.text}</p>}

                                        {/* Error state */}
                                        {msg.error && (
                                            <div className="p-3 bg-red-50 border border-red-100 text-red-700 rounded-xl text-xs flex items-start gap-2">
                                                <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
                                                <span>{msg.error}</span>
                                            </div>
                                        )}

                                        {/* Pending Confirmation Card */}
                                        {msg.pendingConfirmation && msg.actions && msg.actions.length > 0 && (
                                            <div className="mt-3 p-3.5 bg-amber-50/90 border border-amber-200 rounded-xl space-y-3 text-left">
                                                <div className="flex items-start gap-2">
                                                    <ShieldAlert className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
                                                    <div>
                                                        <h4 className="text-xs font-bold text-amber-900">Confirmation Required</h4>
                                                        <p className="text-[11px] text-amber-700 mt-0.5">Please review the proposed inventory changes:</p>
                                                    </div>
                                                </div>

                                                <div className="space-y-2 pt-1 border-t border-amber-200/60">
                                                    {msg.actions.map((act, i) => {
                                                        const details: { label: string; value: string }[] = [];
                                                        const args = act.args || {};

                                                        if (args.items && Array.isArray(args.items)) {
                                                            args.items.forEach((it: any) => {
                                                                if (it.name) details.push({ label: 'Item', value: it.name });
                                                                if (it.freezer_name) details.push({ label: 'Freezer', value: it.freezer_name });
                                                                if (it.category_name) details.push({ label: 'Category', value: it.category_name });
                                                                if (it.count) details.push({ label: 'Qty', value: String(it.count) });
                                                                if (it.weight) details.push({ label: 'Weight', value: it.weight });
                                                            });
                                                        } else {
                                                            if (args.item_name || args.name) details.push({ label: 'Item', value: args.item_name || args.name });
                                                            if (args.freezer_name || args.target_freezer_name) details.push({ label: 'Freezer', value: args.freezer_name || args.target_freezer_name });
                                                            if (args.category_name) details.push({ label: 'Category', value: args.category_name });
                                                            if (args.count) details.push({ label: 'Qty', value: String(args.count) });
                                                            if (args.weight) details.push({ label: 'Weight', value: args.weight });
                                                        }

                                                        return (
                                                            <div key={i} className="text-xs bg-white p-2.5 rounded-lg border border-amber-200 text-gray-800 space-y-1">
                                                                <div className="flex items-center justify-between gap-2">
                                                                    <span className="font-semibold text-gray-900">{act.summary || act.tool_name}</span>
                                                                    <span className="text-[10px] bg-amber-100 text-amber-800 px-1.5 py-0.5 rounded font-mono">{act.tool_name}</span>
                                                                </div>
                                                                {details.length > 0 && (
                                                                    <div className="flex flex-wrap gap-1 pt-1">
                                                                        {details.map((d, idx) => (
                                                                            <span key={idx} className="text-[10px] bg-amber-50 text-amber-900 px-1.5 py-0.5 rounded border border-amber-200/60">
                                                                                <span className="text-amber-700 font-semibold">{d.label}:</span> {d.value}
                                                                            </span>
                                                                        ))}
                                                                    </div>
                                                                )}
                                                            </div>
                                                        );
                                                    })}
                                                </div>

                                                <div className="flex justify-end gap-2 pt-1">
                                                    <button
                                                        onClick={() => handleCancelActions(msg.id)}
                                                        disabled={msg.executing}
                                                        className="px-2.5 py-1 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-xs font-medium flex items-center gap-1 transition-colors cursor-pointer"
                                                    >
                                                        <X className="w-3.5 h-3.5" />
                                                        Cancel
                                                    </button>
                                                    <button
                                                        onClick={() => handleConfirmActions(msg.id, msg.actions!)}
                                                        disabled={msg.executing}
                                                        className="px-3.5 py-1 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-medium flex items-center gap-1.5 transition-colors shadow-xs cursor-pointer"
                                                    >
                                                        {msg.executing ? (
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

                                        {/* Executed Actions Summary Card */}
                                        {msg.executedActions && msg.executedActions.length > 0 && (
                                            <div className="mt-2.5 p-3 bg-indigo-50/70 border border-indigo-100 rounded-xl space-y-2 text-left">
                                                <div className="flex items-center gap-1.5 text-indigo-900 font-semibold text-xs">
                                                    <CheckCircle2 className="w-3.5 h-3.5 text-indigo-600 shrink-0" />
                                                    <span>Executed Changes:</span>
                                                </div>
                                                <div className="space-y-1">
                                                    {msg.executedActions.map((act, i) => (
                                                        <div key={i} className="text-xs bg-white p-2 rounded border border-indigo-100/80 text-gray-700">
                                                            <div className="font-medium text-indigo-800">{act.summary || act.tool_name}</div>
                                                            {act.output && <div className="text-[10px] text-gray-500 font-mono mt-0.5">{act.output}</div>}
                                                        </div>
                                                    ))}
                                                </div>
                                            </div>
                                        )}
                                    </div>

                                    <span className="text-[10px] text-gray-400 px-1">
                                        {msg.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                    </span>
                                </div>
                            </div>
                        );
                    })}

                    {/* Thinking Indicator */}
                    {loading && (
                        <div className="flex items-start gap-2.5">
                            <div className="w-7 h-7 rounded-full bg-white border border-indigo-100 text-indigo-600 flex items-center justify-center shrink-0 text-xs shadow-xs">
                                <Bot className="w-4 h-4" />
                            </div>
                            <div className="bg-white border border-gray-200/80 rounded-2xl rounded-tl-none px-4 py-3 text-xs text-gray-500 flex items-center gap-2 shadow-xs">
                                <Loader2 className="w-4 h-4 animate-spin text-indigo-600" />
                                <span>Thinking...</span>
                            </div>
                        </div>
                    )}

                    {/* Prompt Suggestions Chips (visible when history is short) */}
                    {messages.length <= 1 && (
                        <div className="pt-2">
                            <span className="text-[11px] font-medium text-gray-400 block mb-2">Try an example command:</span>
                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-1.5">
                                {promptSuggestions.map((s, idx) => (
                                    <button
                                        key={idx}
                                        onClick={() => handleProcessPrompt(s)}
                                        disabled={loading}
                                        className="text-left text-xs bg-white hover:bg-indigo-50/80 hover:border-indigo-200 text-gray-700 p-2.5 rounded-xl border border-gray-200 transition-all cursor-pointer shadow-2xs group flex items-start justify-between"
                                    >
                                        <span>{s}</span>
                                        <Sparkles className="w-3 h-3 text-indigo-400 opacity-0 group-hover:opacity-100 transition-opacity shrink-0 mt-0.5" />
                                    </button>
                                ))}
                            </div>
                        </div>
                    )}

                    <div ref={messagesEndRef} />
                </div>

                {/* Bottom Chat Input Bar */}
                <div className="p-3 bg-white border-t border-gray-100">
                    <form
                        onSubmit={(e) => {
                            e.preventDefault();
                            handleProcessPrompt();
                        }}
                        className="flex items-center gap-2"
                    >
                        <input
                            type="text"
                            value={prompt}
                            onChange={(e) => setPrompt(e.target.value)}
                            disabled={loading}
                            placeholder="Type a command (e.g. Added 2 packs of bacon to Kitchen Freezer)..."
                            className="flex-1 px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-xl text-xs focus:bg-white focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition-all"
                        />
                        <button
                            type="submit"
                            disabled={loading || !prompt.trim()}
                            className="px-4 py-2.5 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 disabled:opacity-50 text-white rounded-xl text-xs font-semibold flex items-center gap-1.5 transition-all shadow-xs shrink-0 cursor-pointer"
                        >
                            {loading ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Send className="w-3.5 h-3.5" />}
                            <span>Send</span>
                        </button>
                    </form>
                </div>
            </div>
        </Modal>
    );
}
