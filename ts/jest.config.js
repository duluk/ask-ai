export default {
    preset: 'ts-jest/presets/default-esm',
    extensionsToTreatAsEsm: ['.ts'],
    testEnvironment: 'node',
    moduleDirectories: ['node_modules', 'src'],
    roots: ['<rootDir>/tests'],
    transform: {
        '^.+\\.tsx?$': ['ts-jest', {
            tsconfig: 'tsconfig.json',
            useESM: true
        }]
    },
    moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
        '^@/(.*)\\.(js|ts|tsx)$': '<rootDir>/src/$1.$2',
    },
    collectCoverageFrom: ['<rootDir>/src/**/*.ts'],
    testMatch: ['**/tests/**/*.test.ts'],
};