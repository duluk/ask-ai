import { loadConfig, Config } from '../../src/config/index.js';
import fs from 'fs';
import yaml from 'yaml';

jest.mock('fs');
jest.mock('yaml');

describe('Config Management', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('should load config from file', () => {
        const mockConfig: Partial<Config> = {
            defaultModel: 'chatgpt-4o-latest',
            maxTokens: 2000,
            temperature: 0.8
        };

        (fs.existsSync as jest.Mock).mockReturnValue(true);
        (fs.readFileSync as jest.Mock).mockReturnValue('dummy yaml content');
        (yaml.parse as jest.Mock).mockReturnValue(mockConfig);

        const config = loadConfig();
        expect(config).toMatchObject(mockConfig);
    });

    it('should return default config if file not found', () => {
        (fs.existsSync as jest.Mock).mockReturnValue(false);

        const config = loadConfig();
        expect(config).toMatchObject({
            defaultModel: 'chatgpt-4o-latest',
            maxTokens: 1024,
            temperature: 0.7,
            stream: false
        });
    });

    it('should create config directory if it does not exist', () => {
        (fs.existsSync as jest.Mock)
            .mockReturnValueOnce(false)  // config dir doesn't exist
            .mockReturnValueOnce(false); // config file doesn't exist

        const mkdirSpy = jest.spyOn(fs, 'mkdirSync');

        loadConfig();

        expect(mkdirSpy).toHaveBeenCalledWith(expect.stringContaining('ask-ai'), { recursive: true });
    });

    it('should handle yaml parse errors gracefully', () => {
        (fs.existsSync as jest.Mock).mockReturnValue(true);
        (fs.readFileSync as jest.Mock).mockReturnValue('invalid: yaml: content');
        (yaml.parse as jest.Mock).mockImplementation(() => {
            throw new Error('Invalid YAML');
        });

        const config = loadConfig();
        expect(config).toMatchObject({
            defaultModel: 'chatgpt-4o-latest',
            maxTokens: 1024,
            temperature: 0.7,
            stream: false
        });
    });
});

