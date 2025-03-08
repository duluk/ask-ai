import { Database } from '../../src/db/sqlite.js';
import { Logger } from '../../src/utils/logger.js';
import { loadConfig } from '../../src/config/index.js';

jest.mock('../../src/db/sqlite.js');
jest.mock('../../src/utils/logger.js');
jest.mock('../../src/config/index.js');

describe('CLI', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('should initialize logger correctly', () => {
        const mockLogger = {
            log: jest.fn()
        };
        (Logger.getInstance as jest.Mock).mockReturnValue(mockLogger);
        (loadConfig as jest.Mock).mockReturnValue({
            historyFile: ':memory:'
        });

        // Simulate CLI initialization
        require('../../src/cli/index');

        expect(Logger.initialize).toHaveBeenCalled();
        expect(Database).toHaveBeenCalledWith(':memory:');
    });
});
