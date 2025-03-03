export class LineWrapper {
    private currWidth: number = 0;

    constructor(
        private maxWidth: number,
        private tabWidth: number = 4
    ) { }

    wrap(text: string): string {
        if (!text) return '';

        let result = '';
        let i = 0;

        while (i < text.length) {
            const char = text[i];

            switch (char) {
                case '\n':
                    result += '\n';
                    this.currWidth = 0;
                    break;

                case '\t':
                    const spaces = ' '.repeat(this.tabWidth);
                    result += spaces;
                    this.currWidth += this.tabWidth;
                    break;

                case ' ':
                    // Don't write a space at the beginning of a line
                    if (this.currWidth !== 0) {
                        result += ' ';
                    }

                    // Look ahead for next space to determine word length
                    let nextWord = '';
                    let j = i + 1;
                    while (j < text.length && text[j] !== ' ' && text[j] !== '\n') {
                        nextWord += text[j];
                        j++;
                    }

                    // If adding this word would exceed maxWidth, add newline instead
                    if (this.currWidth + nextWord.length + 1 >= this.maxWidth) {
                        result += '\n';
                        this.currWidth = 0;
                    } else {
                        this.currWidth++;
                    }
                    break;

                default:
                    result += char;
                    this.currWidth++;

                    // Add newline if we've reached maxWidth and next char is a space
                    if (this.currWidth >= this.maxWidth &&
                        i + 1 < text.length &&
                        text[i + 1] === ' ') {
                        result += '\n';
                        this.currWidth = 0;
                    }
            }
            i++;
        }

        return result;
    }
}

// For backwards compatibility, maintain the wrapText function
export function wrapText(text: string, maxLength: number, tabWidth: number = 4): string {
    const wrapper = new LineWrapper(maxLength, tabWidth);
    return wrapper.wrap(text);
}

export function getTerminalWidth(): number {
    // Default width if we can't determine actual width
    const DEFAULT_WIDTH = 80;

    try {
        const width = process.stdout.columns;
        return width || DEFAULT_WIDTH;
    } catch (error) {
        return DEFAULT_WIDTH;
    }
}
