#!/usr/bin/env node

/**
 * Simple TUI wrapper that displays a message about the issue
 * without using any Ink/React components
 */

// Create a simple console-based TUI
function showSimpleTUI() {
  console.clear();
  
  console.log('\n\x1b[1m=== Ask-AI Terminal UI ===\x1b[0m\n');
  console.log('The TUI is currently unavailable due to a module dependency issue.');
  console.log('\n\x1b[33mIssue:\x1b[0m');
  console.log('The error "ERR_REQUIRE_ASYNC_MODULE" is occurring because the Ink 5.x library');
  console.log('depends on yoga-wasm-web, which uses top-level await. This creates conflicts');
  console.log('between CommonJS and ESM module systems in the project.\n');
  
  console.log('\x1b[32mRecommended fixes:\x1b[0m');
  console.log('1. Downgrade to Ink 4.x (which doesn\'t have this dependency issue)');
  console.log('2. Migrate the entire project to native ESM by:');
  console.log('   - Setting "type": "module" in package.json');
  console.log('   - Changing module config in tsconfig.json to "NodeNext"');
  console.log('   - Adding file extensions to all imports');
  console.log('   - Updating package scripts to handle ESM\n');
  
  console.log('\x1b[36mIn the meantime, you can use the non-TUI mode:\x1b[0m');
  console.log('npm run start -- [your query]\n');
  
  console.log('\x1b[36mDetailed information:\x1b[0m');
  console.log('See TUI_ISSUE.md for detailed explanation and permanent fix options.\n');
  
  console.log('Press Ctrl+C to exit.');
  
  // Keep the process running
  process.stdin.resume();
  
  // Handle exit with Ctrl+C
  process.on('SIGINT', () => {
    console.log('\nExiting Ask-AI TUI');
    process.exit(0);
  });
}

// Run the simple TUI
showSimpleTUI();