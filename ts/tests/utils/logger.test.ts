import { Logger } from '../../src/utils/logger';
import path from 'path';
import fs from 'fs';

jest.mock('fs');

describe('Logger', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        Logger.initialize(path.join('test', 'log.txt'));
    });

    it('should initialize logger with correct path', () => {
        expect(fs.mkdirSync).toHaveBeenCalledWith(path.join('test'), { recursive: true });
    });

    it('should log messages correctly', () => {
        const logger = Logger.getInstance();
        const spy = jest.spyOn(logger, 'log');

        logger.log('info', 'test message', { data: 'test' });

        expect(spy).toHaveBeenCalledWith('info', 'test message', { data: 'test' });
    });
});
