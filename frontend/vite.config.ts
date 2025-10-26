import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react({
      // React 插件配置
      jsxImportSource: 'react',
    }),
  ],
  
  // 路径解析配置
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@services': path.resolve(__dirname, './src/services'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@types': path.resolve(__dirname, './src/types'),
      '@constants': path.resolve(__dirname, './src/constants'),
      '@config': path.resolve(__dirname, './src/config'),
      '@styles': path.resolve(__dirname, './src/styles'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@contexts': path.resolve(__dirname, './src/contexts'),
    },
  },
  
  // 开发服务器配置
  server: {
    port: 3000,
    host: '0.0.0.0', // 监听所有网络接口
    open: false, // 禁用自动打开浏览器，由启动脚本控制
    cors: true, // 启用 CORS
    proxy: {
      // API 代理配置
      '/api': {
        target: 'http://localhost:50052',
        changeOrigin: true,
        secure: false,
      },
      // WebSocket 代理配置
      '/ws': {
        target: 'ws://localhost:50052',
        ws: true,
        changeOrigin: true,
      },
    },
  },
  
  // 构建配置
  build: {
    outDir: 'dist',
    sourcemap: true,
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // 生产环境移除 console
        drop_debugger: true, // 生产环境移除 debugger
      },
    },
    rollupOptions: {
      output: {
        manualChunks: {
          // 第三方库分包
          vendor: ['react', 'react-dom'],
          antd: ['antd'],
          router: ['react-router-dom'],
          // 工具库分包
          utils: ['lodash', 'dayjs'],
        },
        // 资源文件命名
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
      },
    },
    // 构建大小限制
    chunkSizeWarningLimit: 1000,
  },
  
  // 环境变量定义
  define: {
    __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '1.0.0'),
    __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
  },
  
  // CSS 配置
  css: {
    preprocessorOptions: {
      less: {
        javascriptEnabled: true,
      },
    },
  },
  
  // 优化配置
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'antd',
      'react-router-dom',
      'lodash',
      'dayjs',
    ],
  },
  
  // 预览配置
  preview: {
    port: 4173,
    host: '0.0.0.0',
    open: false,
  },
});