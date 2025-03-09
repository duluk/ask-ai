import React from 'react';
import { render } from 'ink';
import { Database } from '../db/sqlite.js';
import { AskAITUI } from './components.js';

/**
 * Entry point for the Ink app that handles JSX rendering
 */
export async function startInkApp(
    config: any, 
    db: Database, 
    modelName: string, 
    logger: any
): Promise<void> {
    try {
        // Render the Ink app with Ink 4.x API
        const { waitUntilExit } = render(
            <AskAITUI 
                config={config} 
                db={db} 
                modelName={modelName} 
                logger={logger} 
            />
        );
        
        // Wait for the app to exit
        await waitUntilExit();
    } catch (error) {
        console.error('Error rendering Ink app:', error);
        throw error;
    }
}