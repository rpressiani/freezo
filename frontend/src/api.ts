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

const BASE_URL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080/api' : '/api');

export const api = {
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
};
