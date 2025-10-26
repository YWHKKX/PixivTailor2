import {
  formatFileSize,
  formatTime,
  formatDuration,
  formatRelativeTime,
  debounce,
  throttle,
  delay,
  generateId,
  generateUUID,
  truncateString,
  capitalize,
  toCamelCase,
  toKebabCase,
  deepClone,
  merge,
  filterObject,
  unique,
  groupBy,
  sortBy,
  paginate,
  isValidEmail,
  isValidUrl,
  isValidPhone,
  storage,
  downloadFile,
  downloadBlob,
  random,
  randomString,
  isMobile,
  isDevelopment,
  isProduction
} from '../../utils/index';

describe('Utils', () => {
  describe('formatFileSize', () => {
    it('should format file size correctly', () => {
      expect(formatFileSize(0)).toBe('0 B');
      expect(formatFileSize(1024)).toBe('1 KB');
      expect(formatFileSize(1024 * 1024)).toBe('1 MB');
      expect(formatFileSize(1024 * 1024 * 1024)).toBe('1 GB');
      expect(formatFileSize(1024 * 1024 * 1024 * 1024)).toBe('1 TB');
    });

    it('should handle decimal places correctly', () => {
      expect(formatFileSize(1536)).toBe('1.5 KB');
      expect(formatFileSize(1024 * 1.5)).toBe('1.5 MB');
    });
  });

  describe('formatTime', () => {
    it('should format timestamp correctly', () => {
      const timestamp = 1672531200000; // 2023-01-01T00:00:00Z
      const result = formatTime(timestamp);
      
      expect(result).toContain('2023');
      expect(result).toContain('01');
    });

    it('should format date string correctly', () => {
      const dateString = '2023-01-01T00:00:00Z';
      const result = formatTime(dateString);
      
      expect(result).toContain('2023');
      expect(result).toContain('01');
    });
  });

  describe('formatDuration', () => {
    it('should format duration correctly', () => {
      expect(formatDuration(30)).toBe('30秒');
      expect(formatDuration(90)).toBe('2分钟');
      expect(formatDuration(3660)).toBe('1小时');
    });
  });

  describe('formatRelativeTime', () => {
    it('should format relative time correctly', () => {
      const now = new Date().getTime();
      const oneMinuteAgo = now - 60 * 1000;
      const oneHourAgo = now - 60 * 60 * 1000;
      const oneDayAgo = now - 24 * 60 * 60 * 1000;

      expect(formatRelativeTime(oneMinuteAgo)).toBe('刚刚');
      expect(formatRelativeTime(oneHourAgo)).toBe('1小时前');
      expect(formatRelativeTime(oneDayAgo)).toBe('1天前');
    });
  });

  describe('debounce', () => {
    it('should debounce function calls', (done) => {
      const mockFn = jest.fn();
      const debouncedFn = debounce(mockFn, 100);

      debouncedFn();
      debouncedFn();
      debouncedFn();

      setTimeout(() => {
        expect(mockFn).toHaveBeenCalledTimes(1);
        done();
      }, 150);
    });
  });

  describe('throttle', () => {
    it('should throttle function calls', (done) => {
      const mockFn = jest.fn();
      const throttledFn = throttle(mockFn, 100);

      throttledFn();
      throttledFn();
      throttledFn();

      setTimeout(() => {
        expect(mockFn).toHaveBeenCalledTimes(1);
        done();
      }, 50);
    });
  });

  describe('delay', () => {
    it('should delay execution', async () => {
      const start = Date.now();
      await delay(100);
      const end = Date.now();
      
      expect(end - start).toBeGreaterThanOrEqual(100);
    });
  });

  describe('generateId', () => {
    it('should generate unique IDs', () => {
      const id1 = generateId();
      const id2 = generateId();

      expect(typeof id1).toBe('string');
      expect(id1.length).toBe(9);
      expect(id1).not.toBe(id2);
    });
  });

  describe('generateUUID', () => {
    it('should generate valid UUIDs', () => {
      const uuid1 = generateUUID();
      const uuid2 = generateUUID();

      expect(typeof uuid1).toBe('string');
      expect(uuid1).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i);
      expect(uuid1).not.toBe(uuid2);
    });
  });

  describe('truncateString', () => {
    it('should truncate string correctly', () => {
      expect(truncateString('Hello World', 5)).toBe('Hello...');
      expect(truncateString('Hello World', 11)).toBe('Hello World');
      expect(truncateString('Hello World', 5, '***')).toBe('He***');
    });
  });

  describe('capitalize', () => {
    it('should capitalize first letter', () => {
      expect(capitalize('hello')).toBe('Hello');
      expect(capitalize('HELLO')).toBe('HELLO');
      expect(capitalize('')).toBe('');
    });
  });

  describe('toCamelCase', () => {
    it('should convert to camel case', () => {
      expect(toCamelCase('hello-world')).toBe('helloWorld');
      expect(toCamelCase('hello-world-test')).toBe('helloWorldTest');
      expect(toCamelCase('hello')).toBe('hello');
    });
  });

  describe('toKebabCase', () => {
    it('should convert to kebab case', () => {
      expect(toKebabCase('helloWorld')).toBe('hello-world');
      expect(toKebabCase('helloWorldTest')).toBe('hello-world-test');
      expect(toKebabCase('hello')).toBe('hello');
    });
  });

  describe('deepClone', () => {
    it('should deep clone objects', () => {
      const original = { a: 1, b: { c: 2 } };
      const cloned = deepClone(original);

      expect(cloned).toEqual(original);
      expect(cloned).not.toBe(original);
      expect(cloned.b).not.toBe(original.b);
    });

    it('should handle arrays', () => {
      const original = [1, 2, { a: 3 }];
      const cloned = deepClone(original);

      expect(cloned).toEqual(original);
      expect(cloned).not.toBe(original);
      expect(cloned[2]).not.toBe(original[2]);
    });

    it('should handle primitives', () => {
      expect(deepClone(42)).toBe(42);
      expect(deepClone('hello')).toBe('hello');
      expect(deepClone(true)).toBe(true);
      expect(deepClone(null)).toBe(null);
    });
  });

  describe('merge', () => {
    it('should merge objects correctly', () => {
      const target = { a: 1, b: { c: 2 } };
      const source = { b: { d: 3 }, e: 4 };
      const result = merge(target, source);

      expect(result).toEqual({ a: 1, b: { c: 2, d: 3 }, e: 4 });
    });
  });

  describe('filterObject', () => {
    it('should filter object properties', () => {
      const obj = { a: 1, b: 2, c: 3 };
      const result = filterObject(obj, (value) => value > 1);

      expect(result).toEqual({ b: 2, c: 3 });
    });
  });

  describe('unique', () => {
    it('should remove duplicates from array', () => {
      expect(unique([1, 2, 2, 3, 3, 3])).toEqual([1, 2, 3]);
      expect(unique(['a', 'b', 'a', 'c'])).toEqual(['a', 'b', 'c']);
    });
  });

  describe('groupBy', () => {
    it('should group array by key', () => {
      const items = [
        { category: 'a', value: 1 },
        { category: 'b', value: 2 },
        { category: 'a', value: 3 }
      ];
      const result = groupBy(items, 'category');

      expect(result).toEqual({
        a: [{ category: 'a', value: 1 }, { category: 'a', value: 3 }],
        b: [{ category: 'b', value: 2 }]
      });
    });
  });

  describe('sortBy', () => {
    it('should sort array by key', () => {
      const items = [
        { name: 'c', value: 3 },
        { name: 'a', value: 1 },
        { name: 'b', value: 2 }
      ];
      const result = sortBy(items, 'name');

      expect(result).toEqual([
        { name: 'a', value: 1 },
        { name: 'b', value: 2 },
        { name: 'c', value: 3 }
      ]);
    });

    it('should sort in descending order', () => {
      const items = [
        { name: 'a', value: 1 },
        { name: 'b', value: 2 },
        { name: 'c', value: 3 }
      ];
      const result = sortBy(items, 'value', 'desc');

      expect(result).toEqual([
        { name: 'c', value: 3 },
        { name: 'b', value: 2 },
        { name: 'a', value: 1 }
      ]);
    });
  });

  describe('paginate', () => {
    it('should paginate array correctly', () => {
      const items = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
      
      expect(paginate(items, 1, 3)).toEqual([1, 2, 3]);
      expect(paginate(items, 2, 3)).toEqual([4, 5, 6]);
      expect(paginate(items, 3, 3)).toEqual([7, 8, 9]);
      expect(paginate(items, 4, 3)).toEqual([10]);
    });
  });

  describe('isValidEmail', () => {
    it('should validate email addresses', () => {
      expect(isValidEmail('test@example.com')).toBe(true);
      expect(isValidEmail('user.name@domain.co.uk')).toBe(true);
      expect(isValidEmail('invalid-email')).toBe(false);
      expect(isValidEmail('@example.com')).toBe(false);
      expect(isValidEmail('test@')).toBe(false);
    });
  });

  describe('isValidUrl', () => {
    it('should validate URLs', () => {
      expect(isValidUrl('https://example.com')).toBe(true);
      expect(isValidUrl('http://localhost:3000')).toBe(true);
      expect(isValidUrl('ftp://example.com')).toBe(true);
      expect(isValidUrl('invalid-url')).toBe(false);
      expect(isValidUrl('')).toBe(false);
    });
  });

  describe('isValidPhone', () => {
    it('should validate phone numbers', () => {
      expect(isValidPhone('13800138000')).toBe(true);
      expect(isValidPhone('15900159000')).toBe(true);
      expect(isValidPhone('12345678901')).toBe(false);
      expect(isValidPhone('1380013800')).toBe(false);
      expect(isValidPhone('')).toBe(false);
    });
  });

  describe('storage', () => {
    beforeEach(() => {
      localStorage.clear();
    });

    it('should store and retrieve data', () => {
      const data = { key: 'value' };
      storage.set('test', data);
      const result = storage.get('test');
      
      expect(result).toEqual(data);
    });

    it('should return default value for missing key', () => {
      const result = storage.get('missing', 'default');
      
      expect(result).toBe('default');
    });

    it('should remove data', () => {
      storage.set('test', 'value');
      storage.remove('test');
      const result = storage.get('test');
      
      expect(result).toBeNull();
    });

    it('should clear all data', () => {
      storage.set('test1', 'value1');
      storage.set('test2', 'value2');
      storage.clear();
      
      expect(storage.get('test1')).toBeNull();
      expect(storage.get('test2')).toBeNull();
    });
  });

  describe('downloadFile', () => {
    it('should create download link', () => {
      const createElementSpy = jest.spyOn(document, 'createElement');
      const appendChildSpy = jest.spyOn(document.body, 'appendChild');
      const removeChildSpy = jest.spyOn(document.body, 'removeChild');
      const clickSpy = jest.fn();

      createElementSpy.mockReturnValue({
        href: '',
        download: '',
        click: clickSpy
      } as any);

      downloadFile('http://example.com/file.pdf', 'test.pdf');

      expect(createElementSpy).toHaveBeenCalledWith('a');
      expect(clickSpy).toHaveBeenCalled();
      expect(appendChildSpy).toHaveBeenCalled();
      expect(removeChildSpy).toHaveBeenCalled();

      createElementSpy.mockRestore();
      appendChildSpy.mockRestore();
      removeChildSpy.mockRestore();
    });
  });

  describe('downloadBlob', () => {
    it('should download blob', () => {
      const createObjectURLSpy = jest.spyOn(URL, 'createObjectURL');
      const revokeObjectURLSpy = jest.spyOn(URL, 'revokeObjectURL');
      const downloadFileSpy = jest.spyOn(require('../../utils/index'), 'downloadFile');

      createObjectURLSpy.mockReturnValue('blob:url');
      downloadFileSpy.mockImplementation(() => {});

      const blob = new Blob(['test'], { type: 'text/plain' });
      downloadBlob(blob, 'test.txt');

      expect(createObjectURLSpy).toHaveBeenCalledWith(blob);
      expect(downloadFileSpy).toHaveBeenCalledWith('blob:url', 'test.txt');
      expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:url');

      createObjectURLSpy.mockRestore();
      revokeObjectURLSpy.mockRestore();
      downloadFileSpy.mockRestore();
    });
  });

  describe('random', () => {
    it('should generate random number in range', () => {
      const result = random(1, 10);
      
      expect(result).toBeGreaterThanOrEqual(1);
      expect(result).toBeLessThanOrEqual(10);
      expect(Number.isInteger(result)).toBe(true);
    });
  });

  describe('randomString', () => {
    it('should generate random string', () => {
      const result = randomString(10);
      
      expect(typeof result).toBe('string');
      expect(result.length).toBe(10);
    });
  });

  describe('isMobile', () => {
    it('should detect mobile devices', () => {
      // Mock navigator.userAgent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)',
        writable: true
      });

      expect(isMobile()).toBe(true);
    });

    it('should detect desktop devices', () => {
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        writable: true
      });

      expect(isMobile()).toBe(false);
    });
  });

  describe('isDevelopment', () => {
    it('should detect development environment', () => {
      // Mock import.meta.env.DEV
      Object.defineProperty(import.meta, 'env', {
        value: { DEV: true },
        writable: true
      });

      expect(isDevelopment()).toBe(true);
    });
  });

  describe('isProduction', () => {
    it('should detect production environment', () => {
      // Mock import.meta.env.PROD
      Object.defineProperty(import.meta, 'env', {
        value: { PROD: true },
        writable: true
      });

      expect(isProduction()).toBe(true);
    });
  });
});
