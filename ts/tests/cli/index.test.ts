import { Database } from '../../src/db/sqlite';
import { Logger } from '../../src/utils/logger';
import { loadConfig } from '../../src/config';

jest.mock('../../src/db/sqlite');
jest.mock('../../src/utils/logger');
jest.mock('../../src/config');

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
