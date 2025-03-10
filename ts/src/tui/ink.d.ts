// Type declarations for ink v4.4.1
declare module 'ink' {
    import { FC, ReactNode, Component } from 'react';

    export interface BoxProps {
        flexDirection?: 'row' | 'column';
        flexGrow?: number;
        flexShrink?: number;
        flexBasis?: string | number;
        alignItems?: 'flex-start' | 'center' | 'flex-end';
        alignSelf?: 'flex-start' | 'center' | 'flex-end';
        justifyContent?: 'flex-start' | 'center' | 'flex-end' | 'space-between' | 'space-around';
        marginTop?: number;
        marginBottom?: number;
        marginLeft?: number;
        marginRight?: number;
        paddingTop?: number;
        paddingBottom?: number;
        paddingLeft?: number;
        paddingRight?: number;
        padding?: number;
        borderStyle?: 'single' | 'double' | 'round' | 'bold';
        borderColor?: string;
        height?: number | string;
        width?: number | string;
        minHeight?: number;
        overflow?: 'visible' | 'hidden' | 'scroll' | 'auto';
        backgroundColor?: string;
        children?: ReactNode;
    }

    export interface TextProps {
        color?: string;
        backgroundColor?: string;
        dimColor?: boolean;
        inverse?: boolean;
        bold?: boolean;
        italic?: boolean;
        underline?: boolean;
        strikethrough?: boolean;
        width?: number | string;
        wrap?: 'wrap' | 'truncate' | 'truncate-start' | 'truncate-middle' | 'truncate-end';
        children?: ReactNode;
        padding?: number;
        paddingLeft?: number;
        paddingRight?: number;
    }

    export const Box: FC<BoxProps>;
    export const Text: FC<TextProps>;
    export const Static: FC<{ items?: unknown[]; children: ReactNode }>;

    export interface KeyboardInputOptions {
        escape?: boolean;
        return?: boolean;
        ctrl?: boolean;
        meta?: boolean;
        shift?: boolean;
        tab?: boolean;
        backspace?: boolean;
        delete?: boolean;
        upArrow?: boolean;
        downArrow?: boolean;
        leftArrow?: boolean;
        rightArrow?: boolean;
        pageDown?: boolean;
        pageUp?: boolean;
        name?: string;
    }

    export function useInput(
        callback: (input: string, key: KeyboardInputOptions) => void,
        options?: { isActive?: boolean }
    ): void;

    export interface AppContext {
        exit: (code?: number) => void;
    }

    export function useApp(): AppContext;

    export interface StdoutContext {
        write: (data: string) => void;
        columns: number;
        rows: number;
    }

    export function useStdout(): StdoutContext;

    export interface StdinContext {
        isRawModeSupported: boolean;
        setRawMode: (isRawMode: boolean) => void;
        setEncoding: (encoding: string) => void;
    }

    export function useStdin(): StdinContext;

    export interface RenderOptions {
        stdout?: NodeJS.WriteStream;
        stdin?: NodeJS.ReadStream;
        debug?: boolean;
        exitOnCtrlC?: boolean;
        patchConsole?: boolean;
    }

    export interface Instance {
        rerender: (tree: ReactNode) => void;
        unmount: () => void;
        waitUntilExit: () => Promise<void>;
        clear: () => void;
    }

    export function render(
        element: React.ReactElement,
        options?: RenderOptions
    ): Instance;
}
