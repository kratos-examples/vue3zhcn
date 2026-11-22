import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
    plugins: [vue()],
    server: {
        proxy: {
            '/demo1kratos-base': {
                target: 'http://127.0.0.1:8001',
                // The duplex LookupStudents method goes through a WebSocket. Kratos accepts a WebSocket
                // handshake only when the page Origin matches the Host, so the Host stays the page's own here.
                changeOrigin: false,
                ws: true,
                rewrite: (path) => path.replace(/^\/demo1kratos-base/, ''),
            },
            '/demo2kratos-base': {
                target: 'http://127.0.0.1:8002',
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/demo2kratos-base/, ''),
            },
        },
    },
})
