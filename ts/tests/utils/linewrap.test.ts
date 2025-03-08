import { wrapText, getTerminalWidth } from '../../src/utils/linewrap.js';

describe('Line Wrapping', () => {
    it('should wrap text at specified width', () => {
        const text = 'This is a long line that should be wrapped at a specific width';
        const wrapped = wrapText(text, 20);

        expect(wrapped.split('\n').every(line => line.length <= 20)).toBe(true);
    });

    it('should handle empty strings', () => {
        expect(wrapText('', 80)).toBe('');
    });

    it('should get terminal width', () => {
        const width = getTerminalWidth();
        expect(typeof width).toBe('number');
        expect(width).toBeGreaterThan(0);
    });
});
