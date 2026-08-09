# Freezo Instructions

Welcome to the Freezo instructions guide! Here you will find everything you need to know to get the most out of your self-hosted food inventory tracker.

## Overview
Freezo allows you to:
- **Manage Multiple Freezers:** Create separate virtual freezers to track items in different physical spaces (e.g., Garage Freezer, Kitchen Freezer).
- **Categories & Custom Icons:** Categorize items into pre-seeded categories (*Beef, Pork, Poultry, Seafood, Bread, Uncategorized*) or create custom categories with tailored icons.
- **Filter Inventory:** Filter items dynamically by freezer or category.
- **Track Inventory & Frozen Dates:** Log items with quantity, weight, category, and frozen date (which defaults automatically to today's date). Identical items in the same freezer are automatically grouped.
- **Move & Change Category:** Seamlessly move items between freezers (by clicking freezer badges or using the Move button) and change item categories on the fly.
- **Consume & Deduct Items:** Remove items completely or deduct specific quantities when using food from your freezer.
- **Backup & Restore:** Export timestamped SQLite backups (`freezo_backup_YYYY-MM-DD_HH-MM-SS.db`), restore snapshots, or reset your data.

## Getting Started

1. **Create a Freezer:** 
   Head to the **Settings** menu and select **Add Freezer** (or click **Add Freezer** on the main view if no freezers exist). Name it according to your physical setup.
   
2. **Manage Categories:**
   In **Settings**, under **Category Management**, you can add new custom categories with your choice of icons (piggy bank, fish, beef, drumstick, croissant, tag, pizza, apple, etc.) or delete custom categories.

3. **Add Items:** 
   Click the `+` button in the bottom right corner (or top menu action). Select the target freezer, choose an item name (utilize autocomplete suggestions), select a category, specify weight, frozen date (defaults to today), and set the quantity.

4. **Filter Inventory:**
   Use the top filter bar on the inventory screen to filter items by specific freezers or categories for quick access.

5. **Moving Items Between Freezers:**
   To move items to a different freezer:
   - Click directly on the **Freezer Badge** attached to any item card, or
   - Click the **Move** action button on the item card.
   Choose the destination freezer and the quantity to transfer.

6. **Updating Item Category:**
   Click the **Category Badge** on an item card (or click the category update icon) to select a new category for the item or group.

7. **Consuming & Deducting Items:** 
   Click the **Consume** button on an item card to deduct a specific quantity or remove the item entirely from inventory.

## Backup and Database Management

Backups are natively supported inside Freezo:

1. Go to the **Settings** menu.
2. Under **Backup & Restore**, click **Export Backup** to generate a local `.db` file containing your full inventory snapshot (named `freezo_backup_YYYY-MM-DD_HH-MM-SS.db`).
3. **Restore Backup:** If you need to restore your inventory, click **Restore Backup**, select your `.db` backup file, and confirm. This will safely replace your active database state.
4. **Reset Database (Danger Zone):** To wipe all tables and return to clean default categories, select **Reset Database** and type `RESET` when prompted.

