#!/usr/bin/env node

/**
 * This is a standalone launcher for the TUI that works around the
 * top-level await issue in dependencies.
 * 
 * Run this directly with: node tui-standalone.js
 */

const path = require('path');
const { exec } = require('child_process');

// Get the absolute path to the current directory
const cwd = process.cwd();
const entryPath = path.resolve(cwd, 'dist/src/tui-entry.js');

// Generate a temporary script that uses dynamic import with absolute path
const scriptContent = `
// Use absolute path to the entry file
import('${entryPath.replace(/\\/g, '\\\\')}').catch(error => {
  console.error('Error loading TUI:', error);
  process.exit(1);
});
`;

// Write the script to a temporary file
const fs = require('fs');
const os = require('os');
const tempDir = os.tmpdir();
const tempFile = path.join(tempDir, `ask-ai-tui-${Date.now()}.mjs`);

try {
    console.log(`Creating temporary launcher at: ${tempFile}`);
    fs.writeFileSync(tempFile, scriptContent);

    // Execute the temporary script with Node.js
    console.log(`Launching TUI from: ${entryPath}`);
    const child = exec(`node --experimental-modules ${tempFile}`, {
        cwd: cwd,
        env: process.env
    });

    // Forward stdout and stderr
    child.stdout.pipe(process.stdout);
    child.stderr.pipe(process.stderr);

    // Handle process exit
    child.on('exit', (code) => {
        // Clean up the temporary file
        try {
            fs.unlinkSync(tempFile);
        } catch (err) {
            // Ignore cleanup errors
        }
        process.exit(code || 0);
    });

    // Handle process errors
    child.on('error', (error) => {
        console.error('Failed to start TUI:', error);
        try {
            fs.unlinkSync(tempFile);
        } catch (err) {
            // Ignore cleanup errors
        }
        process.exit(1);
    });

} catch (error) {
    console.error('Error setting up TUI:', error);
    process.exit(1);
}
