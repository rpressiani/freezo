# Freezo Instructions

Welcome to the Freezo instructions guide! Here you will find everything you need to know to get the most out of your self-hosted food inventory tracker.

## Overview
Freezo allows you to:
- **Manage Multiple Freezers:** Create separate virtual freezers to track items in different physical spaces (e.g., Garage Freezer, Kitchen Freezer).
- **Track Inventory:** Log items with their quantity, weight and an optional expiration date to avoid food waste.
- **Move Items:** Seamlessly move items back and forth between your different freezers.
- **Backup & Restore:** Easily save and load database backups so your inventory is never lost.

## Getting Started

1. **Create a Freezer:** 
   Head to the Settings page and select "Add Freezer". Name it whatever helps you identify it physically.
   
2. **Add Items:** 
   Click the `+` button in the bottom right corner. Select the freezer, name your item, and define the quantity.

3. **Consuming Items:** 
   When you use an item, easily click the **Consume** button to remove it completely or deduct a specific amount from the total quantity.

## Backup and Database Management
Backups are directly supported natively inside Freezo. 

1. Go to the **Settings** menu.
2. Under **Backup & Restore**, click **Export Backup** to generate a local `.db` file containing your full inventory snapshot.
3. If you ever need to restore, click **Restore Backup** and upload your previously saved file. This will safely overwrite your current application state.
4. **Danger Zone:** Use this option sparingly to completely wipe all tables in the database and start entirely fresh.
