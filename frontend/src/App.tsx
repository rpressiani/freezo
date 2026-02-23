import { useEffect, useState } from 'react';
import { formatDistanceToNow, isToday, isYesterday, parse } from 'date-fns';
import { api, type Freezer, type Item } from './api';
import { Plus, Trash2, Snowflake, AlertCircle, ChevronDown, ChevronUp, Package, Scale, Calendar, Settings, ArrowLeft } from 'lucide-react';
import { Modal } from './components/Modal';

interface DateGroup {
  date: string; // YYYY-MM-DD or 'No Date'
  items: Item[];
  totalQuantity: number;
  freezerId: number;
}

interface ItemGroup {
  name: string;
  totalQuantity: number;
  dateGroups: DateGroup[];
}

type WeightMode = 'none' | 'same' | 'individual';

function getRelativeTime(dateStr: string): string {
  // Parse YYYY-MM-DD as local date to ensure isToday/isYesterday work correctly with local timeframe
  const created = parse(dateStr, 'yyyy-MM-dd', new Date());

  if (isToday(created)) return 'Today';
  if (isYesterday(created)) return 'Yesterday';

  return formatDistanceToNow(created, { addSuffix: true });
}

function App() {
  const [currentView, setCurrentView] = useState<'items' | 'settings'>('items');
  const [freezers, setFreezers] = useState<Freezer[]>([]);
  const [items, setItems] = useState<Item[]>([]);
  const [selectedFreezerFilter, setSelectedFreezerFilter] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set());

  // Modal States
  const [isAddFreezerOpen, setIsAddFreezerOpen] = useState(false);
  const [isAddItemOpen, setIsAddItemOpen] = useState(false);

  // Delete Confirmation State
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [freezerToDelete, setFreezerToDelete] = useState<Freezer | null>(null);

  // No Freezer Warning State
  const [isNoFreezerWarningOpen, setIsNoFreezerWarningOpen] = useState(false);

  // Consume Modal State
  const [consumeModalOpen, setConsumeModalOpen] = useState(false);
  const [consumeContext, setConsumeContext] = useState<{
    items: Item[]; // All items in the selected date group
    selectedWeight: string | null; // User selected weight (or null if uniform/none)
    step: 'select-weight' | 'select-quantity';
  } | null>(null);
  const [consumeQty, setConsumeQty] = useState(1);

  // Move Modal State
  const [moveModalOpen, setMoveModalOpen] = useState(false);
  const [moveContext, setMoveContext] = useState<{
    items: Item[];
    selectedWeight: string | null;
    step: 'select-weight' | 'select-details' | 'confirm-move-all';
  } | null>(null);
  const [moveQty, setMoveQty] = useState(1);
  const [moveTargetFreezerId, setMoveTargetFreezerId] = useState<number>(0);

  // Backup / Restore Modal States
  const [isRestoreConfirmOpen, setIsRestoreConfirmOpen] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const [restoreFile, setRestoreFile] = useState<File | null>(null);

  const [isResetConfirmOpen, setIsResetConfirmOpen] = useState(false);
  const [isResetting, setIsResetting] = useState(false);
  const [resetConfirmText, setResetConfirmText] = useState('');

  const [notificationModal, setNotificationModal] = useState({ isOpen: false, title: '', message: '', type: 'success' as 'success' | 'error' });

  // Form States
  const [freezerName, setFreezerName] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [isNameLocked, setIsNameLocked] = useState(false);
  const [itemForm, setItemForm] = useState({
    name: '',
    quantity: 1 as number | string,
    freezerId: 0,
    frozenDate: '',
    weightMode: 'none' as WeightMode,
    commonWeight: '',
    individualWeights: [] as string[],
  });

  const loadData = async () => {
    try {
      setLoading(true);
      const [f, i] = await Promise.all([api.getFreezers(), api.getItems()]);
      setFreezers(f);
      setItems(i);
      setError(null);
    } catch (err) {
      console.error(err);
      setError('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  // Group items by Name -> Expiry Date
  const filteredItems = selectedFreezerFilter
    ? items.filter(i => i.freezer_id === selectedFreezerFilter)
    : items;

  const itemGroups: ItemGroup[] = Object.values(filteredItems.reduce((acc, item) => {
    if (!acc[item.name]) {
      acc[item.name] = { name: item.name, totalQuantity: 0, dateGroups: [] };
    }

    const dateKey = item.frozen_date ? item.frozen_date.split('T')[0] : 'No Date';
    let dateGroup = acc[item.name].dateGroups.find(g => g.date === dateKey && g.freezerId === item.freezer_id);

    if (!dateGroup) {
      dateGroup = { date: dateKey, items: [], totalQuantity: 0, freezerId: item.freezer_id };
      acc[item.name].dateGroups.push(dateGroup);
    }

    dateGroup.items.push(item);
    dateGroup.totalQuantity += 1;
    acc[item.name].totalQuantity += 1;

    return acc;
  }, {} as Record<string, ItemGroup>)).map(group => ({
    ...group,
    dateGroups: group.dateGroups.sort((a, b) => {
      const dateDiff = a.date.localeCompare(b.date);
      if (dateDiff !== 0) return dateDiff;
      const f1 = freezers.find(f => f.id === a.freezerId)?.name || '';
      const f2 = freezers.find(f => f.id === b.freezerId)?.name || '';
      return f1.localeCompare(f2);
    })
  }));

  const toggleGroup = (name: string) => {
    const newExpanded = new Set(expandedGroups);
    if (newExpanded.has(name)) {
      newExpanded.delete(name);
    } else {
      newExpanded.add(name);
    }
    setExpandedGroups(newExpanded);
  };

  const initiateDeleteFreezer = (freezer: Freezer) => {
    setFreezerToDelete(freezer);
    setIsDeleteModalOpen(true);
  };

  const confirmDeleteFreezer = async () => {
    if (!freezerToDelete) return;
    try {
      await api.deleteFreezer(freezerToDelete.id);
      setIsDeleteModalOpen(false);
      setFreezerToDelete(null);
      loadData();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const openAddItemModal = (prefillName?: string) => {
    if (freezers.length === 0) {
      if (freezers.length === 0) {
        setIsNoFreezerWarningOpen(true);
        return;
      }
    }
    setItemForm({
      name: prefillName || '',
      quantity: 1,
      freezerId: freezers[0].id,
      frozenDate: '',
      weightMode: 'none',
      commonWeight: '',
      individualWeights: [''],
    });
    setFormError(null);
    setSuggestions([]);
    setShowSuggestions(false);
    setIsNameLocked(!!prefillName);
    setIsAddItemOpen(true);
  };

  const handleCreateFreezer = async () => {
    if (!freezerName) return;
    try {
      await api.createFreezer(freezerName);
      setIsAddFreezerOpen(false);
      setFreezerName('');
      loadData();
    } catch (err) {
      alert('Failed to create freezer');
    }
  };

  const handleExport = async () => {
    try {
      await api.exportDatabase();
    } catch (err: any) {
      alert(err.message || 'Failed to export database');
    }
  };

  const handleImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setRestoreFile(file);
    setIsRestoreConfirmOpen(true);
    e.target.value = ''; // Reset input
  };

  const confirmRestore = async () => {
    if (!restoreFile) return;
    setIsRestoring(true);

    try {
      await api.importDatabase(restoreFile);
      setIsRestoreConfirmOpen(false);
      setNotificationModal({
        isOpen: true,
        title: 'Success',
        message: 'Database restored successfully',
        type: 'success'
      });
      await loadData();
    } catch (err: any) {
      setIsRestoreConfirmOpen(false);
      setNotificationModal({
        isOpen: true,
        title: 'Restore Failed',
        message: err.message || 'Failed to import database',
        type: 'error'
      });
    } finally {
      setIsRestoring(false);
      setRestoreFile(null);
    }
  };

  const confirmReset = async () => {
    if (resetConfirmText !== 'DELETE') return;
    setIsResetting(true);

    try {
      await api.resetDatabase();
      setIsResetConfirmOpen(false);
      setNotificationModal({
        isOpen: true,
        title: 'Success',
        message: 'Database reset successfully',
        type: 'success'
      });
      await loadData();
    } catch (err: any) {
      setIsResetConfirmOpen(false);
      setNotificationModal({
        isOpen: true,
        title: 'Reset Failed',
        message: err.message || 'Failed to reset database',
        type: 'error'
      });
    } finally {
      setIsResetting(false);
      setResetConfirmText('');
    }
  };

  const handleCreateItem = async () => {
    if (!itemForm.name || !itemForm.freezerId) return;

    try {
      const itemsToCreate = [];
      const qty = typeof itemForm.quantity === 'string' ? parseInt(itemForm.quantity) || 0 : itemForm.quantity;

      if (qty < 1) {
        setFormError('Quantity must be at least 1');
        return;
      }

      for (let i = 0; i < qty; i++) {
        let weight: string | undefined;
        if (itemForm.weightMode === 'same') weight = itemForm.commonWeight;
        if (itemForm.weightMode === 'individual') weight = itemForm.individualWeights[i];

        itemsToCreate.push({
          name: itemForm.name,
          category_id: 1,
          freezer_id: itemForm.freezerId,
          weight: weight || undefined,
          frozen_date: itemForm.frozenDate ? new Date(itemForm.frozenDate).toISOString() : undefined,
        });
      }

      await api.createItemsBatch(itemsToCreate);
      setIsAddItemOpen(false);
      loadData();
    } catch (err: any) {
      alert(err.message || 'Failed to create items');
    }
  };

  const openConsumeModal = (dateGroup: DateGroup) => {
    // Check distinct weights
    const weights = Array.from(new Set(dateGroup.items.map(i => i.weight || 'No Weight')));

    if (weights.length > 1) {
      // Mixed weights -> Step 1: Select Weight
      setConsumeContext({
        items: dateGroup.items,
        selectedWeight: null,
        step: 'select-weight'
      });
    } else {
      // Uniform weights -> Step 2: Select Quantity
      setConsumeContext({
        items: dateGroup.items,
        selectedWeight: weights[0] === 'No Weight' ? null : weights[0],
        step: 'select-quantity'
      });
      setConsumeQty(1);
    }
    setConsumeModalOpen(true);
  };

  const handleConsumeSubmit = async () => {
    if (!consumeContext) return;

    // Filter items matching the selected weight (if any)
    const targetItems = consumeContext.items.filter(i =>
      (consumeContext.selectedWeight === null && !i.weight) ||
      i.weight === consumeContext.selectedWeight
    );

    if (consumeQty > targetItems.length) {
      alert('Cannot consume more than available');
      return;
    }

    // Take the first N items to delete
    const itemsToDelete = targetItems.slice(0, consumeQty);
    const deleteIds = itemsToDelete.map(i => i.id);

    try {
      await api.consumeItemsBatch(deleteIds);
      setConsumeModalOpen(false);
      setConsumeContext(null);
      loadData();
    } catch (err: any) {
      alert(err.message || 'Failed to consume items');
    }
  };

  const openMoveModal = (items: Item[], isMoveAll: boolean = false) => {
    // Check distinct freezers for these items
    const sourceFreezerIds = new Set(items.map(i => i.freezer_id));

    // Set default target to first freezer available that is NOT in the source list (if possible)
    const validTargets = freezers.filter(f => !sourceFreezerIds.has(f.id));
    if (validTargets.length > 0) {
      setMoveTargetFreezerId(validTargets[0].id);
    } else if (freezers.length > 0) {
      // Fallback if somehow we can't find a different one (e.g. only 1 freezer exists)
      setMoveTargetFreezerId(freezers[0].id);
    }

    if (isMoveAll) {
      setMoveContext({
        items,
        selectedWeight: null,
        step: 'confirm-move-all'
      });
      setMoveQty(items.length); // Default to all
      setMoveModalOpen(true);
      return;
    }

    // Check distinct weights
    const weights = Array.from(new Set(items.map(i => i.weight || 'No Weight')));

    if (weights.length > 1) {
      setMoveContext({
        items,
        selectedWeight: null,
        step: 'select-weight'
      });
    } else {
      setMoveContext({
        items,
        selectedWeight: weights[0] === 'No Weight' ? null : weights[0],
        step: 'select-details'
      });
      setMoveQty(1);
    }
    setMoveModalOpen(true);
  };

  const handleMoveSubmit = async () => {
    if (!moveContext || !moveTargetFreezerId) return;

    let itemsToMove: Item[] = [];

    if (moveContext.step === 'confirm-move-all') {
      itemsToMove = moveContext.items;
      // In move-all mode, moveQty is implicitly items.length, but we just take all items in context
    } else {
      const specificItems = moveContext.items.filter(i =>
        (moveContext.selectedWeight === null && !i.weight) || i.weight === moveContext.selectedWeight
      );

      if (moveQty > specificItems.length) {
        alert('Cannot move more than available');
        return;
      }
      itemsToMove = specificItems.slice(0, moveQty);
    }

    const moveIds = itemsToMove.map(i => i.id);

    try {
      await api.moveItems(moveIds, moveTargetFreezerId);
      setMoveModalOpen(false);
      setMoveContext(null);
      loadData();
    } catch (err: any) {
      alert(err.message || 'Failed to move items');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 font-sans pb-24">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between sticky top-0 z-10 transition-colors">
        <div className="flex items-center gap-2">
          {currentView === 'settings' && (
            <button
              onClick={() => setCurrentView('items')}
              className="mr-2 p-1 hover:bg-gray-100 rounded-full transition-colors"
            >
              <ArrowLeft className="w-6 h-6 text-gray-600" />
            </button>
          )}
          <div className="bg-cyan-100 p-2 rounded-lg">
            <Snowflake className="w-6 h-6 text-cyan-600" />
          </div>
          <h1 className="text-xl font-bold text-gray-800">
            {currentView === 'items' ? 'Freezo' : 'Settings'}
          </h1>
        </div>

        {currentView === 'items' && (
          <button
            onClick={() => setCurrentView('settings')}
            className="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <Settings className="w-6 h-6" />
          </button>
        )}
      </header>

      {/* Main Content */}
      <div className="max-w-5xl mx-auto mt-6 px-4">
        {loading ? (
          <div className="text-center py-12 text-gray-500">Loading...</div>
        ) : error ? (
          <div className="text-center py-12 text-red-500 flex flex-col items-center gap-2">
            <AlertCircle className="w-8 h-8" />
            {error}
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-4">
            {currentView === 'settings' ? (
              // Settings View (Freezer Management)
              <div className="space-y-6">
                <div className="flex justify-between items-center">
                  <h2 className="text-lg font-semibold text-gray-700">Manage Freezers</h2>
                  <button
                    onClick={() => setIsAddFreezerOpen(true)}
                    className="bg-cyan-600 hover:bg-cyan-700 text-white px-4 py-2 rounded-lg flex items-center gap-2 transition-colors"
                  >
                    <Plus className="w-4 h-4" />
                    Add Freezer
                  </button>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {freezers.map(f => {
                    const itemCount = items.filter(i => i.freezer_id === f.id).length;
                    return (
                      <div
                        key={f.id}
                        className="bg-white p-4 rounded-xl border border-gray-200 shadow-sm hover:shadow-md transition-shadow group cursor-pointer"
                        onClick={() => {
                          setSelectedFreezerFilter(f.id);
                          setCurrentView('items');
                        }}
                      >
                        <div className="flex justify-between items-center">
                          <div className="flex gap-4 items-center">
                            <div className="bg-cyan-50 p-3 rounded-xl">
                              <Snowflake className="w-6 h-6 text-cyan-600" />
                            </div>
                            <div>
                              <h3 className="font-semibold text-gray-900">{f.name}</h3>
                              <p className="text-sm text-gray-400 mt-0.5">{itemCount} items</p>
                            </div>
                          </div>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              initiateDeleteFreezer(f);
                            }}
                            disabled={itemCount > 0}
                            className={`transition-colors p-2 rounded-lg ${itemCount > 0
                              ? 'text-gray-300 cursor-not-allowed'
                              : 'text-gray-400 hover:text-red-600 hover:bg-red-50 opacity-0 group-hover:opacity-100'}`}
                            title={itemCount > 0 ? "Cannot delete freezer with items" : "Delete freezer"}
                          >
                            <Trash2 className="w-5 h-5" />
                          </button>
                        </div>
                      </div>
                    )
                  })}
                  {freezers.length === 0 && (
                    <div className="col-span-full text-center py-8 text-gray-500 bg-white rounded-xl border border-dashed border-gray-300">
                      No freezers found. Create one to get started.
                    </div>
                  )}
                </div>

                <div className="pt-8 border-t border-gray-200 mt-8">
                  <h2 className="text-lg font-semibold text-gray-700 mb-4">Backup & Restore</h2>
                  <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div>
                      <h3 className="font-medium text-gray-900">Database Backup</h3>
                      <p className="text-sm text-gray-500 mt-1 max-w-lg">Download a full backup of your freezer inventory, or restore from a previously exported backup file.</p>
                    </div>
                    <div className="flex flex-col sm:flex-row gap-3">
                      <button
                        onClick={handleExport}
                        className="bg-cyan-600 hover:bg-cyan-700 text-white px-4 py-2 rounded-lg transition-colors font-medium whitespace-nowrap shadow-sm text-sm"
                      >
                        Export Backup
                      </button>
                      <label className="bg-white hover:bg-gray-50 text-gray-700 border border-gray-300 px-4 py-2 rounded-lg transition-colors font-medium whitespace-nowrap shadow-sm cursor-pointer text-sm text-center">
                        Restore Backup
                        <input
                          type="file"
                          className="hidden"
                          accept=".db,.sqlite"
                          onChange={handleImport}
                        />
                      </label>
                    </div>
                  </div>

                  <div className="bg-white p-6 rounded-xl border border-red-200 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-4 mt-6">
                    <div>
                      <h3 className="font-medium text-red-700">Danger Zone</h3>
                      <p className="text-sm text-gray-500 mt-1 max-w-lg">Permanently delete all freezers, categories, and items. This will completely wipe your database.</p>
                    </div>
                    <div className="flex flex-col sm:flex-row gap-3">
                      <button
                        onClick={() => {
                          setResetConfirmText('');
                          setIsResetConfirmOpen(true);
                        }}
                        className="bg-red-50 hover:bg-red-100 text-red-600 border border-red-200 px-4 py-2 rounded-lg transition-colors font-medium whitespace-nowrap shadow-sm text-sm"
                      >
                        Reset Database
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              // Items View
              <div className="space-y-4">
                {selectedFreezerFilter && (
                  <div className="bg-blue-50 border border-blue-100 rounded-xl p-4 flex items-center justify-between animate-in fade-in slide-in-from-top-2">
                    <div className="flex items-center gap-3 text-blue-800">
                      <Snowflake className="w-5 h-5" />
                      <span className="font-medium">
                        Viewing items in <strong>{freezers.find(f => f.id === selectedFreezerFilter)?.name}</strong>
                      </span>
                    </div>
                    <button
                      onClick={() => setSelectedFreezerFilter(null)}
                      className="text-sm bg-white hover:bg-blue-100 text-blue-600 px-3 py-1.5 rounded-lg border border-blue-200 transition-colors font-medium"
                    >
                      Clear Filter
                    </button>
                  </div>
                )}
                {itemGroups.map(group => (
                  <div key={group.name} className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                    <div
                      className="p-4 flex items-center justify-between cursor-pointer hover:bg-gray-50 transition-colors"
                      onClick={() => toggleGroup(group.name)}
                    >
                      <div className="flex items-center gap-4">
                        <div className="bg-orange-100 p-2 rounded-lg">
                          <Package className="w-6 h-6 text-orange-600" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-lg">{group.name}</h3>
                          <p className="text-gray-500 text-sm">{group.totalQuantity} total items</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          onClick={(e) => { e.stopPropagation(); openMoveModal(group.dateGroups.flatMap(d => d.items), true); }}
                          className="text-sm bg-blue-50 text-blue-700 px-3 py-1.5 rounded-md hover:bg-blue-100 transition-colors font-medium mr-2"
                        >
                          Move All
                        </button>
                        <button
                          onClick={(e) => { e.stopPropagation(); openAddItemModal(group.name); }}
                          className="text-sm bg-cyan-50 text-cyan-700 px-3 py-1.5 rounded-md hover:bg-cyan-100 transition-colors font-medium"
                        >
                          + Add Batch
                        </button>
                        {expandedGroups.has(group.name) ? <ChevronUp className="w-5 h-5 text-gray-400" /> : <ChevronDown className="w-5 h-5 text-gray-400" />}
                      </div>
                    </div>

                    {expandedGroups.has(group.name) && (
                      <div className="border-t border-gray-100 bg-gray-50/50">
                        {group.dateGroups.map(dateGroup => (
                          <div key={dateGroup.date} className="p-4 border-b border-gray-100 last:border-0 hover:bg-white transition-colors">
                            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                              {/* Left Side: Date & Details */}
                              <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-6 flex-1">

                                {/* Date Section */}
                                <div className="flex items-center gap-2 text-sm text-gray-600 sm:w-48 shrink-0">
                                  <div className="bg-gray-100 p-1.5 rounded-md">
                                    <Calendar className="w-4 h-4 text-gray-500" />
                                  </div>
                                  <div className="flex flex-col">
                                    <span className="font-medium text-gray-900">{dateGroup.date}</span>
                                    {dateGroup.date !== 'No Date' && (
                                      <span className="text-xs text-gray-500">
                                        {getRelativeTime(dateGroup.date)}
                                      </span>
                                    )}
                                  </div>
                                </div>

                                {/* Details Row (Qty, Weight, Freezer) */}
                                <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-gray-600">

                                  {/* Quantity */}
                                  <div className="flex items-center gap-2">
                                    <span className="font-bold text-gray-900 text-base">{dateGroup.totalQuantity}x</span>
                                    <span className="text-xs font-medium text-gray-400 uppercase tracking-wide">Qty</span>
                                  </div>

                                  {/* Weight */}
                                  <div className="flex items-center gap-2">
                                    <Scale className="w-4 h-4 text-gray-400" />
                                    <span className="text-gray-700">
                                      {Array.from(new Set(dateGroup.items.map(i => i.weight || 'No Weight'))).join(', ')}
                                    </span>
                                  </div>

                                  <button
                                    onClick={() => openMoveModal(dateGroup.items)}
                                    className="flex items-center gap-2 px-2.5 py-1.5 -mx-2 rounded-lg bg-cyan-50/50 border border-transparent hover:bg-cyan-100 hover:border-cyan-200 text-cyan-700 transition-all cursor-pointer group shadow-sm hover:shadow"
                                    title="Move items to another freezer"
                                  >
                                    <Snowflake className="w-4 h-4 text-cyan-500 group-hover:scale-110 transition-transform" />
                                    <span
                                      className="font-medium truncate max-w-[150px]"
                                    >
                                      {freezers.find(f => f.id === dateGroup.freezerId)?.name || 'Unknown'}
                                    </span>
                                  </button>
                                </div>
                              </div>

                              {/* Right Side: Action Button */}
                              <div className="flex justify-end sm:block pt-2 sm:pt-0 border-t sm:border-t-0 border-gray-50 sm:border-none">
                                <button
                                  onClick={() => openConsumeModal(dateGroup)}
                                  className="text-sm font-medium text-red-600 bg-red-50 hover:bg-red-100 px-4 py-2 rounded-lg transition-colors w-full sm:w-auto text-center"
                                >
                                  Consume
                                </button>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
                {filteredItems.length === 0 && (
                  <div className="text-center py-12 text-gray-500">
                    {selectedFreezerFilter
                      ? "No items in this freezer."
                      : "No items found. Tap the + button to add one!"}
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Floating Action Button (FAB) */}
      {
        currentView === 'items' && (
          <button
            onClick={() => openAddItemModal()}
            className="fixed bottom-6 left-6 bg-cyan-600 hover:bg-cyan-700 text-white p-4 rounded-full shadow-lg transition-transform hover:scale-110 active:scale-95 z-50 flex items-center justify-center"
            aria-label="Add Item"
          >
            <Plus className="w-8 h-8" />
          </button>
        )
      }

      {/* Add Freezer Modal (Only needed for logic, UI moved to Settings page but logic is shared or can be triggered from there) */}
      <Modal
        isOpen={isAddFreezerOpen}
        onClose={() => setIsAddFreezerOpen(false)}
        title="Add New Freezer"
        footer={
          <>
            <button onClick={() => setIsAddFreezerOpen(false)} className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">Cancel</button>
            <button onClick={handleCreateFreezer} className="px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700">Create Freezer</button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 outline-none"
              value={freezerName}
              onChange={e => setFreezerName(e.target.value)}
              placeholder="e.g. Garage Freezer"
              autoFocus
            />
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        title="Delete Freezer"
        footer={
          <>
            <button onClick={() => setIsDeleteModalOpen(false)} className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">Cancel</button>
            <button onClick={confirmDeleteFreezer} className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">Delete</button>
          </>
        }
      >
        <p className="text-gray-600">
          Are you sure you want to delete <strong>{freezerToDelete?.name}</strong>?
          <br /><br />
          This action cannot be undone.
        </p>
      </Modal>

      {/* Restore Confirmation Modal */}
      <Modal
        isOpen={isRestoreConfirmOpen}
        onClose={() => !isRestoring && setIsRestoreConfirmOpen(false)}
        title="Restore Backup"
        footer={
          isRestoring ? (
            <div className="flex justify-center w-full py-2">
              <span className="text-gray-500 font-medium animate-pulse">Restoring database...</span>
            </div>
          ) : (
            <>
              <button
                onClick={() => { setIsRestoreConfirmOpen(false); setRestoreFile(null); }}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                Cancel
              </button>
              <button
                onClick={confirmRestore}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Restore
              </button>
            </>
          )
        }
      >
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-3 text-red-600 bg-red-50 p-3 rounded-lg border border-red-100">
            <AlertCircle className="w-6 h-6 shrink-0" />
            <span className="font-semibold text-sm">Warning: Destructive Action</span>
          </div>
          <p className="text-gray-600">
            Are you sure you want to restore this backup? <strong>This will completely overwrite your current database and CANNOT be undone.</strong>
          </p>
          {restoreFile && (
            <p className="text-sm text-gray-500">File selected: {restoreFile.name}</p>
          )}
        </div>
      </Modal>

      {/* Reset Confirmation Modal */}
      <Modal
        isOpen={isResetConfirmOpen}
        onClose={() => !isResetting && setIsResetConfirmOpen(false)}
        title="Reset Database"
        footer={
          isResetting ? (
            <div className="flex justify-center w-full py-2">
              <span className="text-gray-500 font-medium animate-pulse">Resetting database...</span>
            </div>
          ) : (
            <>
              <button
                onClick={() => { setIsResetConfirmOpen(false); setResetConfirmText(''); }}
                className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
              >
                Cancel
              </button>
              <button
                onClick={confirmReset}
                disabled={resetConfirmText !== 'DELETE'}
                className={`px-4 py-2 text-white rounded-lg transition-colors ${resetConfirmText === 'DELETE' ? 'bg-red-600 hover:bg-red-700' : 'bg-red-300 cursor-not-allowed'}`}
              >
                Confirm Reset
              </button>
            </>
          )
        }
      >
        <div className="flex flex-col gap-4">
          <div className="flex items-center gap-3 text-red-600 bg-red-50 p-3 rounded-lg border border-red-100">
            <AlertCircle className="w-6 h-6 shrink-0" />
            <span className="font-semibold text-sm">Warning: Destructive Action</span>
          </div>
          <p className="text-gray-600">
            Are you sure you want to reset your database? <strong>This will permanently delete all freezers, categories, and items. This CANNOT be undone.</strong>
          </p>
          <div className="pt-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Type <strong>DELETE</strong> below to confirm:
            </label>
            <input
              type="text"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 outline-none"
              value={resetConfirmText}
              onChange={(e) => setResetConfirmText(e.target.value)}
              placeholder="DELETE"
              autoComplete="off"
            />
          </div>
        </div>
      </Modal>

      {/* Notification Modal */}
      <Modal
        isOpen={notificationModal.isOpen}
        onClose={() => setNotificationModal(prev => ({ ...prev, isOpen: false }))}
        title={notificationModal.title}
        footer={
          <button
            onClick={() => setNotificationModal(prev => ({ ...prev, isOpen: false }))}
            className="px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700 w-full sm:w-auto"
          >
            OK
          </button>
        }
      >
        <div className="flex items-center gap-3">
          {notificationModal.type === 'error' ? (
            <AlertCircle className="w-6 h-6 text-red-500 shrink-0" />
          ) : (
            <div className="w-6 h-6 rounded-full bg-green-100 flex items-center justify-center shrink-0">
              <span className="text-green-600 font-bold text-xs">✓</span>
            </div>
          )}
          <p className="text-gray-600">{notificationModal.message}</p>
        </div>
      </Modal>

      {/* Add Item Modal */}
      <Modal
        isOpen={isAddItemOpen}
        onClose={() => setIsAddItemOpen(false)}
        title="Add Items"
        footer={
          <>
            <button onClick={() => setIsAddItemOpen(false)} className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">Cancel</button>
            <button onClick={handleCreateItem} className="px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700">
              Add {itemForm.quantity || 0} Item{(itemForm.quantity && Number(itemForm.quantity) > 1) ? 's' : ''}
            </button>
          </>
        }
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Item Name</label>
            <div className="relative">
              <input
                type="text"
                className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 outline-none ${isNameLocked ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : ''}`}
                value={itemForm.name}
                disabled={isNameLocked}
                onChange={e => {
                  if (isNameLocked) return;
                  const val = e.target.value;
                  setItemForm({ ...itemForm, name: val });
                  if (val.length > 0) {
                    const uniqueNames = Array.from(new Set(items.map(i => i.name)));
                    const matches = uniqueNames
                      .filter(name => name.toLowerCase().includes(val.toLowerCase()) && name !== val)
                      .sort()
                      .slice(0, 5);
                    setSuggestions(matches);
                    setShowSuggestions(matches.length > 0);
                  } else {
                    setShowSuggestions(false);
                  }
                }}
                onFocus={() => {
                  if (itemForm.name.length > 0) {
                    const uniqueNames = Array.from(new Set(items.map(i => i.name)));
                    const matches = uniqueNames
                      .filter(name => name.toLowerCase().includes(itemForm.name.toLowerCase()) && name !== itemForm.name)
                      .sort()
                      .slice(0, 5);
                    setSuggestions(matches);
                    setShowSuggestions(matches.length > 0);
                  }
                }}
                onBlur={() => {
                  // Slight delay to allow click on suggestion to register
                  setTimeout(() => setShowSuggestions(false), 200);
                }}
                placeholder="e.g. Pulled Pork"
                autoFocus
              />
              {showSuggestions && (
                <div className="absolute z-10 w-full bg-white mt-1 border border-gray-200 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                  {suggestions.map((suggestion, index) => (
                    <div
                      key={index}
                      className="px-4 py-2 hover:bg-cyan-50 cursor-pointer text-sm text-gray-700 hover:text-cyan-700 transition-colors"
                      onMouseDown={(e) => {
                        e.preventDefault(); // Prevent blur before click fires
                        setItemForm({ ...itemForm, name: suggestion });
                        setShowSuggestions(false);
                      }}
                    >
                      {suggestion}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Quantity</label>
              <input
                type="number"
                min="1"
                className={`w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 outline-none ${formError ? 'border-red-500' : 'border-gray-300'}`}
                value={itemForm.quantity}
                onChange={e => {
                  setFormError(null);
                  const val = e.target.value;
                  if (val === '') {
                    setItemForm({
                      ...itemForm,
                      quantity: '',
                      individualWeights: []
                    });
                  } else {
                    const qty = parseInt(val);
                    if (!isNaN(qty)) {
                      setItemForm({
                        ...itemForm,
                        quantity: qty,
                        individualWeights: Array(qty).fill('')
                      });
                    }
                  }
                }}
              />
              {formError && <p className="text-red-500 text-xs mt-1">{formError}</p>}
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Freezer</label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 outline-none"
                value={itemForm.freezerId}
                onChange={e => setItemForm({ ...itemForm, freezerId: parseInt(e.target.value) })}
              >
                {freezers.map(f => <option key={f.id} value={f.id}>{f.name}</option>)}
              </select>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Date Frozen</label>
            <input
              type="date"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 outline-none"
              value={itemForm.frozenDate}
              onChange={e => setItemForm({ ...itemForm, frozenDate: e.target.value })}
            />
          </div>

          <div className="border-t border-gray-100 pt-4">
            <div className="flex items-center justify-between mb-3">
              <label className="block text-sm font-medium text-gray-700">Weight Options</label>
              <button
                onClick={() => {
                  if (itemForm.weightMode === 'none') {
                    setItemForm({ ...itemForm, weightMode: 'individual' });
                  } else {
                    setItemForm({ ...itemForm, weightMode: 'none' });
                  }
                }}
                className={`flex items-center gap-2 px-3 py-1.5 text-sm rounded-md border transition-colors ${itemForm.weightMode !== 'none'
                  ? 'bg-cyan-50 border-cyan-200 text-cyan-700 font-medium'
                  : 'border-gray-200 text-gray-600 hover:bg-gray-50'
                  }`}
              >
                <Scale className="w-4 h-4" />
                {itemForm.weightMode !== 'none' ? 'Remove Weight' : 'Add Weight'}
              </button>
            </div>

            {itemForm.weightMode !== 'none' && (
              <div className="animate-in fade-in slide-in-from-top-2 duration-200">
                <div className="flex justify-end mb-2">
                  <button
                    onClick={() => setItemForm({
                      ...itemForm,
                      weightMode: itemForm.weightMode === 'same' ? 'individual' : 'same'
                    })}
                    className="text-xs text-cyan-600 hover:text-cyan-700 underline"
                  >
                    {itemForm.weightMode === 'same' ? 'Switch to Individual Weights' : 'Switch to Same Weight for All'}
                  </button>
                </div>

                {itemForm.weightMode === 'same' ? (
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 outline-none"
                    placeholder="Weight (e.g. 500g)"
                    value={itemForm.commonWeight}
                    onChange={e => setItemForm({ ...itemForm, commonWeight: e.target.value })}
                    autoFocus
                  />
                ) : (
                  <div className="space-y-2 max-h-40 overflow-y-auto p-1 pr-2">
                    {Array.from({ length: Number(itemForm.quantity) || 0 }).map((_, idx) => (
                      <div key={idx} className="flex items-center gap-2">
                        <span className="text-sm text-gray-500 w-6">#{idx + 1}</span>
                        <input
                          type="text"
                          className="flex-1 px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-cyan-500 focus:border-cyan-500 outline-none"
                          placeholder={`Weight for item ${idx + 1}`}
                          value={itemForm.individualWeights[idx] || ''}
                          onChange={e => {
                            const newWeights = [...itemForm.individualWeights];
                            newWeights[idx] = e.target.value;
                            setItemForm({ ...itemForm, individualWeights: newWeights });
                          }}
                        />
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </Modal>

      {/* Consume Modal */}
      <Modal
        isOpen={consumeModalOpen}
        onClose={() => setConsumeModalOpen(false)}
        title="Consume Item"
        footer={
          <>
            <button onClick={() => setConsumeModalOpen(false)} className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">Cancel</button>
            {consumeContext?.step === 'select-quantity' && (
              <button onClick={handleConsumeSubmit} className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">Consume</button>
            )}
          </>
        }
      >
        <div className="space-y-4">
          {consumeContext?.step === 'select-weight' && (
            <>
              <p className="text-gray-600">Select which weight batch to consume from:</p>
              <div className="grid grid-cols-2 gap-3">
                {Array.from(new Set(consumeContext.items.map(i => i.weight || 'No Weight'))).map(weight => (
                  <button
                    key={weight}
                    onClick={() => {
                      setConsumeContext({
                        ...consumeContext,
                        selectedWeight: weight === 'No Weight' ? null : weight,
                        step: 'select-quantity'
                      });
                      setConsumeQty(1);
                    }}
                    className="p-3 border border-gray-200 rounded-lg hover:bg-cyan-50 hover:border-cyan-200 transition-colors text-left"
                  >
                    <div className="font-medium text-gray-900">{weight}</div>
                    <div className="text-sm text-gray-500">
                      {consumeContext.items.filter(i => (i.weight || 'No Weight') === weight).length} items
                    </div>
                  </button>
                ))}
              </div>
            </>
          )}

          {consumeContext?.step === 'select-quantity' && (
            <>
              <p className="text-gray-600">
                How many <strong>{consumeContext.items[0].name}</strong> ({consumeContext.selectedWeight || 'No Weight'}) did you consume?
              </p>
              <div className="flex items-center gap-4">
                <button
                  onClick={() => setConsumeQty(Math.max(1, consumeQty - 1))}
                  className="p-2 border border-gray-200 rounded-lg hover:bg-gray-50"
                >
                  <ChevronDown className="w-5 h-5" />
                </button>
                <span className="text-2xl font-bold w-12 text-center">{consumeQty}</span>
                <button
                  onClick={() => {
                    const maxQty = consumeContext!.items
                      .filter(i => (consumeContext!.selectedWeight === null && !i.weight) || i.weight === consumeContext!.selectedWeight).length;
                    setConsumeQty(Math.min(maxQty, consumeQty + 1));
                  }}
                  className="p-2 border border-gray-200 rounded-lg hover:bg-gray-50"
                >
                  <ChevronUp className="w-5 h-5" />
                </button>
              </div>
              <p className="text-sm text-gray-500">
                Available: {consumeContext!.items
                  .filter(i => (consumeContext!.selectedWeight === null && !i.weight) || i.weight === consumeContext!.selectedWeight).length}
              </p>
            </>
          )}
        </div>
      </Modal>

      {/* Move Modal */}
      <Modal
        isOpen={moveModalOpen}
        onClose={() => setMoveModalOpen(false)}
        title="Move Items"
        footer={
          <>
            <button onClick={() => setMoveModalOpen(false)} className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg">Cancel</button>
            {(moveContext?.step === 'select-details' || moveContext?.step === 'confirm-move-all') && (
              <button onClick={handleMoveSubmit} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">Move Items</button>
            )}
          </>
        }
      >
        <div className="space-y-4">
          {moveContext?.step === 'select-weight' && (
            <>
              <p className="text-gray-600">Select which weight batch to move:</p>
              <div className="grid grid-cols-2 gap-3">
                {Array.from(new Set(moveContext.items.map(i => i.weight || 'No Weight'))).map(weight => (
                  <button
                    key={weight}
                    onClick={() => {
                      setMoveContext({
                        ...moveContext!,
                        selectedWeight: weight === 'No Weight' ? null : weight,
                        step: 'select-details'
                      });
                      setMoveQty(1);
                    }}
                    className="p-3 border border-gray-200 rounded-lg hover:bg-cyan-50 hover:border-cyan-200 transition-colors text-left"
                  >
                    <div className="font-medium text-gray-900">{weight}</div>
                    <div className="text-sm text-gray-500">
                      {moveContext.items.filter(i => (i.weight || 'No Weight') === weight).length} items
                    </div>
                  </button>
                ))}
              </div>
            </>
          )}

          {moveContext?.step === 'select-details' && (
            <div className="space-y-4">
              <div>
                <p className="text-gray-600 mb-2">How many items to move?</p>
                <div className="flex items-center gap-4">
                  <button
                    onClick={() => setMoveQty(Math.max(1, moveQty - 1))}
                    className="p-2 bg-gray-100 rounded-lg hover:bg-gray-200"
                  >
                    <ChevronDown className="w-5 h-5" />
                  </button>
                  <span className="text-xl font-bold w-12 text-center">{moveQty}</span>
                  <button
                    onClick={() => setMoveQty(Math.min(
                      moveContext!.items.filter(i => (moveContext!.selectedWeight === null && !i.weight) || i.weight === moveContext!.selectedWeight).length,
                      moveQty + 1
                    ))}
                    className="p-2 bg-gray-100 rounded-lg hover:bg-gray-200"
                  >
                    <ChevronUp className="w-5 h-5" />
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Destination Freezer</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                  value={moveTargetFreezerId || ''}
                  onChange={e => setMoveTargetFreezerId(parseInt(e.target.value))}
                >
                  {freezers
                    .filter(f => !moveContext!.items.every(i => i.freezer_id === f.id))
                    .map(f => <option key={f.id} value={f.id}>{f.name}</option>)}
                </select>
              </div>
            </div>
          )}

          {moveContext?.step === 'confirm-move-all' && (
            <div className="space-y-4">
              <div>
                <p className="text-gray-600 mb-2">
                  Moving <strong>{moveContext.items.length} items</strong>.
                </p>
                <p className="text-sm text-gray-500">
                  This includes all weights and dates in this group.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Destination Freezer</label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
                  value={moveTargetFreezerId || ''}
                  onChange={e => setMoveTargetFreezerId(parseInt(e.target.value))}
                >
                  {freezers
                    .filter(f => !moveContext!.items.every(i => i.freezer_id === f.id))
                    .map(f => <option key={f.id} value={f.id}>{f.name}</option>)}
                </select>
              </div>
            </div>
          )}
        </div>
      </Modal>

      {/* No Freezer Warning Modal */}
      <Modal
        isOpen={isNoFreezerWarningOpen}
        onClose={() => setIsNoFreezerWarningOpen(false)}
        title="No Freezers Found"
        footer={
          <>
            <button
              onClick={() => setIsNoFreezerWarningOpen(false)}
              className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
            >
              Cancel
            </button>
            <button
              onClick={() => {
                setIsNoFreezerWarningOpen(false);
                setIsAddFreezerOpen(true);
              }}
              className="px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700"
            >
              Create Freezer
            </button>
          </>
        }
      >
        <p className="text-gray-600">
          You need to create a freezer before you can add items.
        </p>
      </Modal>
    </div>
  );
}

export default App;
