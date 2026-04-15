import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = typeof __dirname !== 'undefined' ? __dirname : path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist-export-reader',
    assetsDir: 'static',
    cssCodeSplit: false,
    rollupOptions: {
      input: path.resolve(dirname, 'export-reader.html'),
      output: {
        inlineDynamicImports: true,
        entryFileNames: 'static/export-reader.js',
        chunkFileNames: 'static/export-reader-[name].js',
        assetFileNames: 'static/[name][extname]',
      },
    },
  },
});
