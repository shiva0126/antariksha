// Library build of the design system (web/src/ds) for /design-sync. Output in
// dist-ds is a self-contained package: index.js, astrisk-ds.css, types/.
import { readFileSync, writeFileSync } from 'node:fs';
import { defineConfig, type Plugin } from 'vite';
import react from '@vitejs/plugin-react';

// The app loads its fonts from Google Fonts in index.html; the library
// stylesheet imports the same families so designs render in brand type.
const fonts = "@import url('https://fonts.googleapis.com/css2?family=DM+Sans:opsz,wght@9..40,300;9..40,400;9..40,500;9..40,600&family=Marcellus&family=Noto+Serif+Devanagari:wght@500&display=swap');\n";

const packageFiles: Plugin = {
  name: 'astrisk-ds-package',
  writeBundle() {
    const css = 'dist-ds/astrisk-ds.css';
    writeFileSync(css, fonts + readFileSync(css, 'utf8'));
    writeFileSync('dist-ds/package.json', JSON.stringify({
      name: '@astrisk/ds', version: '1.0.0', type: 'module', private: true,
      module: 'index.js', types: 'types/index.d.ts', style: 'astrisk-ds.css',
      peerDependencies: { react: '^18', 'react-dom': '^18' },
    }, null, 2) + '\n');
  },
};

export default defineConfig({
  plugins: [react(), packageFiles],
  publicDir: false,
  build: {
    outDir: 'dist-ds',
    emptyOutDir: true,
    lib: { entry: 'src/ds/index.ts', formats: ['es'], fileName: () => 'index.js' },
    rollupOptions: { external: ['react', 'react-dom', 'react/jsx-runtime'], output: { assetFileNames: 'astrisk-ds[extname]' } },
  },
});
