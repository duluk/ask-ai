module.exports = {
    preset: 'ts-jest',
    testEnvironment: 'node',
    moduleDirectories: ['node_modules', 'src'],
    roots: ['<rootDir>/tests'],
    transform: {
        '^.+\\.tsx?$': ['ts-jest', {
            tsconfig: 'tsconfig.json'
        }]
    },
    moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
    },
    collectCoverageFrom: ['<rootDir>/src/**/*.ts'],
    testMatch: ['**/tests/**/*.test.ts'],
};
