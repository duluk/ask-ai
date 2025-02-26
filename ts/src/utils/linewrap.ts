/**
 * Line wrapping utility functions
 */

export function wrapText(text: string, maxLength: number, tabWidth: number = 4): string {
  if (!text) return '';
  
  // Replace tabs with spaces for consistent measurement
  const expandedText = text.replace(/\t/g, ' '.repeat(tabWidth));
  
  // Split on newlines first to preserve paragraph breaks
  const paragraphs = expandedText.split(/\r?\n/);
  
  // Process each paragraph
  const wrappedParagraphs = paragraphs.map(paragraph => {
    if (paragraph.length <= maxLength) {
      return paragraph; // No wrapping needed
    }
    
    // Wrap this paragraph
    const words = paragraph.split(' ');
    const lines: string[] = [];
    let currentLine = '';
    
    for (const word of words) {
      // Check if adding this word would exceed the max length
      if (currentLine.length + word.length + 1 > maxLength) {
        // Line would be too long, start a new line
        if (currentLine) {
          lines.push(currentLine);
          currentLine = word;
        } else {
          // The word itself is longer than the max length
          lines.push(word);
          currentLine = '';
        }
      } else {
        // Add word to current line
        currentLine = currentLine 
          ? currentLine + ' ' + word 
          : word;
      }
    }
    
    // Add the last line if there's anything left
    if (currentLine) {
      lines.push(currentLine);
    }
    
    return lines.join('\n');
  });
  
  return wrappedParagraphs.join('\n');
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