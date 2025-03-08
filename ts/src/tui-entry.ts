#!/usr/bin/env node

// This file serves as a direct entry point for the TUI
import { startTUI } from './tui/index.js';

// Create an async function to wrap our code and run it immediately
(async function runTUI() {
    try {
        // Start the TUI with default options
        await startTUI();
    } catch (error) {
        console.error('Error in TUI:', error);
        process.exit(1);
    }
})();
