import { defineConfig } from 'vite'
import { idePlugin } from '@antennae-space/plugin-sdk/vite'

// web/src/index.tsx → web/index.js, one ES module as the IDE loads it; one Solid for the
// plugin and the SDK (a linked SDK has its own).
export default defineConfig({ plugins: [idePlugin()], resolve: { dedupe: ['solid-js'] } })
