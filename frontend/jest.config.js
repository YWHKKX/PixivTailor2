export default {
  // 测试环境
  testEnvironment: 'jsdom',
  
  // 测试设置文件
  setupFilesAfterEnv: ['<rootDir>/src/__tests__/setupTests.ts'],
  
  // 模块名映射
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@components/(.*)$': '<rootDir>/src/components/$1',
    '^@pages/(.*)$': '<rootDir>/src/pages/$1',
    '^@services/(.*)$': '<rootDir>/src/services/$1',
    '^@utils/(.*)$': '<rootDir>/src/utils/$1',
    '^@types/(.*)$': '<rootDir>/src/types/$1',
    '^@constants/(.*)$': '<rootDir>/src/constants/$1',
    '^@config/(.*)$': '<rootDir>/src/config/$1',
    '^@styles/(.*)$': '<rootDir>/src/styles/$1',
    '^@hooks/(.*)$': '<rootDir>/src/hooks/$1',
    '^@contexts/(.*)$': '<rootDir>/src/contexts/$1',
  },
  
  // 转换配置
  transform: {
    '^.+\\.(ts|tsx)$': 'ts-jest',
  },
  
  // 测试文件匹配模式
  testMatch: [
    '<rootDir>/src/**/__tests__/**/*.(ts|tsx)',
    '<rootDir>/src/**/*.(test|spec).(ts|tsx)',
  ],
  
  // 覆盖率收集配置
  collectCoverageFrom: [
    'src/**/*.(ts|tsx)',
    '!src/**/*.d.ts',
    '!src/setupTests.ts',
    '!src/__tests__/**',
    '!src/**/*.stories.(ts|tsx)',
    '!src/**/*.config.(ts|tsx)',
    '!src/main.tsx',
    '!src/vite-env.d.ts',
  ],
  
  // 覆盖率报告器
  coverageReporters: ['text', 'lcov', 'html', 'json'],
  
  // 覆盖率目录
  coverageDirectory: 'coverage',
  
  // 覆盖率阈值
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70,
    },
  },
  
  // 模块文件扩展名
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json'],
  
  // 测试路径忽略模式
  testPathIgnorePatterns: ['/node_modules/', '/dist/', '/build/'],
  
  // 转换忽略模式
  transformIgnorePatterns: [
    'node_modules/(?!(antd|@ant-design|rc-.*)/)',
  ],
  
  // 测试超时时间
  testTimeout: 10000,
  
  // 详细输出
  verbose: true,
  
  // 清理模拟
  clearMocks: true,
  
  // 恢复模拟
  restoreMocks: true,
  
  // 模拟模块
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@components/(.*)$': '<rootDir>/src/components/$1',
    '^@pages/(.*)$': '<rootDir>/src/pages/$1',
    '^@services/(.*)$': '<rootDir>/src/services/$1',
    '^@utils/(.*)$': '<rootDir>/src/utils/$1',
    '^@types/(.*)$': '<rootDir>/src/types/$1',
    '^@constants/(.*)$': '<rootDir>/src/constants/$1',
    '^@config/(.*)$': '<rootDir>/src/config/$1',
    '^@styles/(.*)$': '<rootDir>/src/styles/$1',
    '^@hooks/(.*)$': '<rootDir>/src/hooks/$1',
    '^@contexts/(.*)$': '<rootDir>/src/contexts/$1',
  },
  
  // 全局设置
  globals: {
    'ts-jest': {
      tsconfig: 'tsconfig.json',
    },
  },
  
  // 测试环境选项
  testEnvironmentOptions: {
    customExportConditions: ['node'],
  },
  
  // 模块解析
  moduleDirectories: ['node_modules'],
  
  // 报告器
  reporters: ['default'],
};