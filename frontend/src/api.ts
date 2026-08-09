export interface Freezer {
    id: number;
    name: string;
}

export interface Item {
    id: number;
    name: string;
    category_id: number;
    freezer_id: number;
    weight?: string;
    frozen_date?: string;
}
export interface Category {
    id: number;
    name: string;
    icon?: string;
}

const BASE_URL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080/api' : '/api');

export const api = {
    getCategories: async (): Promise<Category[]> => {
        const res = await fetch(`${BASE_URL}/categories`);
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to fetch categories');
        }
        return res.json();
    },

    createCategory: async (name: string, icon?: string): Promise<Category> => {
        const res = await fetch(`${BASE_URL}/categories`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, icon }),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to create category');
        }
        return res.json();
    },

    deleteCategory: async (id: number): Promise<void> => {
        const res = await fetch(`${BASE_URL}/categories/${id}`, { method: 'DELETE' });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to delete category');
        }
    },

    getFreezers: async (): Promise<Freezer[]> => {
        const res = await fetch(`${BASE_URL}/freezers`);
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to fetch freezers');
        }
        return res.json();
    },

    createFreezer: async (name: string): Promise<Freezer> => {
        const res = await fetch(`${BASE_URL}/freezers`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name }),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to create freezer');
        }
        return res.json();
    },

    deleteFreezer: async (id: number): Promise<void> => {
        const res = await fetch(`${BASE_URL}/freezers/${id}`, { method: 'DELETE' });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to delete freezer');
        }
    },

    getItems: async (): Promise<Item[]> => {
        const res = await fetch(`${BASE_URL}/items`);
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to fetch items');
        }
        return res.json();
    },

    createItem: async (item: Omit<Item, 'id'>): Promise<Item> => {
        const res = await fetch(`${BASE_URL}/items`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to create item');
        }
        return res.json();
    },

    createItemsBatch: async (items: Omit<Item, 'id'>[]): Promise<Item[]> => {
        const res = await fetch(`${BASE_URL}/items/batch`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(items),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to create items batch');
        }
        return res.json();
    },

    updateItem: async (id: number, item: Partial<Item>): Promise<Item> => {
        const res = await fetch(`${BASE_URL}/items/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to update item');
        }
        return res.json();
    },

    consumeItemsBatch: async (deleteIds: number[]): Promise<void> => {
        const res = await fetch(`${BASE_URL}/items/consume`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ delete_ids: deleteIds }),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to consume items batch');
        }
    },

    deleteItem: async (id: number): Promise<void> => {
        const res = await fetch(`${BASE_URL}/items/${id}`, { method: 'DELETE' });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to delete item');
        }
    },

    moveItems: async (itemIds: number[], newFreezerId: number): Promise<void> => {
        const res = await fetch(`${BASE_URL}/items/move`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ item_ids: itemIds, new_freezer_id: newFreezerId }),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to move items');
        }
    },

    updateItemsCategory: async (itemIds: number[], categoryId: number): Promise<void> => {
        const res = await fetch(`${BASE_URL}/items/category`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ item_ids: itemIds, category_id: categoryId }),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to update items category');
        }
    },

    exportDatabase: async (): Promise<void> => {
        const res = await fetch(`${BASE_URL}/database/export`);
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to export database');
        }
        const blob = await res.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        const datetime = new Date().toISOString().replace(/[:.]/g, '-');
        a.download = `freezo-${datetime}.db`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    },

    importDatabase: async (file: File): Promise<void> => {
        const formData = new FormData();
        formData.append('database', file);

        const res = await fetch(`${BASE_URL}/database/import`, {
            method: 'POST',
            body: formData,
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to import database');
        }
    },

    resetDatabase: async (): Promise<void> => {
        const res = await fetch(`${BASE_URL}/database/reset`, {
            method: 'POST',
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to reset database');
        }
    },

    processAIPrompt: async (prompt: string): Promise<AIProcessResponse> => {
        const res = await fetch(`${BASE_URL}/ai/process`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ prompt }),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to process AI prompt');
        }
        return res.json();
    },

    executeAIActions: async (actions: AIActionResult[]): Promise<AIProcessResponse> => {
        const res = await fetch(`${BASE_URL}/ai/execute`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ actions }),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to execute AI actions');
        }
        return res.json();
    },

    getAIConfig: async (): Promise<AIConfig> => {
        const res = await fetch(`${BASE_URL}/ai/config`);
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to get AI config');
        }
        return res.json();
    },

    updateAIConfig: async (config: AIConfig): Promise<AIConfig> => {
        const res = await fetch(`${BASE_URL}/ai/config`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config),
        });
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Failed to update AI config');
        }
        return res.json();
    },
};

export interface AIConfig {
    base_url: string;
    api_key: string;
    model: string;
}

export interface AIActionResult {
    tool_name: string;
    args: any;
    summary?: string;
    output?: string;
}

export interface AIProcessResponse {
    message: string;
    actions: AIActionResult[];
    pending_confirmation?: boolean;
}


