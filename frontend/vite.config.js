import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
            'react': path.resolve(__dirname, './node_modules/react'),
            'react-dom': path.resolve(__dirname, './node_modules/react-dom'),
        },
        dedupe: ['react', 'react-dom'],
    },
    server: {
        port: 5173,
        proxy: {
            '/api': {
                target: 'http://localhost:8100',
                changeOrigin: true,
            },
        },
    },
    test: {
        globals: true,
        environment: 'jsdom',
        include: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
        setupFiles: ['./src/test-setup.ts'],
        coverage: {
            provider: 'v8',
            exclude: [
                'node_modules/',
                'src/test-setup.ts',
                'src/**/*.test.ts',
                'src/**/*.test.tsx',
                'src/**/*.d.ts',
                'src/pages/**/*.tsx',
                'src/components/Layout/index.tsx',
            ],
        },
        deps: {
            optimizer: {
                web: {
                    include: ['react', 'react-dom', '@testing-library/react'],
                },
            },
        },
    },
});
