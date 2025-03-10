#!/usr/bin/env node

import React from 'react';
import { render } from 'ink';
import { App } from './tui';

// Allow top-level await by wrapping in an async IIFE
(async () => {
  const { waitUntilExit } = render(React.createElement(App));
  await waitUntilExit();
})();