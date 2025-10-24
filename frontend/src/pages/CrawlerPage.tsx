import React, { useState, useEffect, useRef } from 'react';
import {
    Row,
    Col,
    Card,
    Table,
    Button,
    Input,
    Select,
    Space,
    Typography,
    Progress,
    Tag,
    Statistic,
    Timeline,
    Modal,
    Form,
    message,
    Tooltip,
    Badge,
    Spin
} from 'antd';
import {
    PlayCircleOutlined,
    PauseCircleOutlined,
    StopOutlined,
    ReloadOutlined,
    EyeOutlined,
    DownloadOutlined,
    DeleteOutlined,
    ExclamationCircleOutlined,
    PlusOutlined,
    SearchOutlined,
    FolderOutlined,
    FileImageOutlined,
    FileOutlined,
    FilePdfOutlined,
    FileWordOutlined,
    FileExcelOutlined,
    FileZipOutlined,
    FileTextOutlined,
    FilePptOutlined,
    FileGifOutlined,
    FileJpgOutlined,
    CodeOutlined
} from '@ant-design/icons';

// 导入API服务和类型
import { apiService } from '@/services/api';
import { wsManager } from '@/services/websocket';
import { Task, PixivImage, CrawlRequest, TaskStatus, CrawlType, Order, Mode } from '@/services/appState';

const { Text } = Typography;
const { Option } = Select;
const { Search } = Input;

// 文件树节点类型
interface FileTreeNode {
    key: string;
    title: string;
    icon?: React.ReactNode;
    children?: FileTreeNode[];
    isLeaf?: boolean;
    filePath?: string;
    fileSize?: number;
    fileType?: string;
    isExpanded?: boolean;
    level?: number;
}

const CrawlerPage: React.FC = () => {
    // 添加CSS样式
    React.useEffect(() => {
        const style = document.createElement('style');
        style.textContent = `
            .file-tree-node:hover {
                background: linear-gradient(135deg, #e6f7ff 0%, #f0f9ff 100%) !important;
                box-shadow: 0 2px 8px rgba(24, 144, 255, 0.15);
                transform: translateX(2px);
            }
            
            .file-tree-node:active {
                background: linear-gradient(135deg, #bae7ff 0%, #e6f7ff 100%) !important;
                transform: translateX(1px);
            }
            
            .ant-tree .ant-tree-node-content-wrapper {
                padding: 0 !important;
                border-radius: 4px;
                transition: all 0.2s ease;
            }
            
            .ant-tree .ant-tree-node-content-wrapper:hover {
                background: transparent !important;
            }
            
            .ant-tree .ant-tree-node-content-wrapper.ant-tree-node-selected {
                background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%) !important;
                color: white !important;
                box-shadow: 0 2px 8px rgba(24, 144, 255, 0.3);
            }
            
            .ant-tree .ant-tree-node-content-wrapper.ant-tree-node-selected .file-tree-node {
                color: white !important;
            }
            
            .ant-tree .ant-tree-node-content-wrapper.ant-tree-node-selected .file-tree-node span {
                color: white !important;
            }
            
            .ant-tree .ant-tree-switcher {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 20px;
                height: 20px;
                border-radius: 50%;
                background: #f0f0f0;
                transition: all 0.2s ease;
            }
            
            .ant-tree .ant-tree-switcher:hover {
                background: #d9d9d9;
                transform: scale(1.1);
            }
            
            .ant-tree .ant-tree-switcher-noop {
                width: 20px;
            }
            
            .ant-tree .ant-tree-iconEle {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 20px;
                height: 20px;
            }
            
            .ant-tree .ant-tree-treenode {
                padding: 2px 0;
            }
            
            .ant-tree .ant-tree-treenode:hover {
                background: transparent;
            }
            
            .ant-tree .ant-tree-treenode-selected {
                background: transparent;
            }
            
            .ant-tree .ant-tree-treenode-selected .ant-tree-node-content-wrapper {
                background: transparent !important;
            }
        `;
        document.head.appendChild(style);

        return () => {
            document.head.removeChild(style);
        };
    }, []);
    const [tasks, setTasks] = useState<Task[]>([]);
    const [results, setResults] = useState<PixivImage[]>([]);
    const [isModalVisible, setIsModalVisible] = useState(false);
    const [isDetailModalVisible, setIsDetailModalVisible] = useState(false);
    const [selectedTask, setSelectedTask] = useState<Task | null>(null);
    const [form] = Form.useForm();
    const [taskLogs, setTaskLogs] = useState<Record<string, Array<{ level: string, message: string, time: string }>>>({});
    const [globalLogs, setGlobalLogs] = useState<Array<{ level: string, message: string, time: string }>>([]);
    const [loading, setLoading] = useState(false);
    const [crawlType, setCrawlType] = useState<string>('tag');
    const [proxyEnabled, setProxyEnabled] = useState<boolean>(false);
    const [proxyUrl, setProxyUrl] = useState<string>('127.0.0.1:7890');
    const [useCookie, setUseCookie] = useState<boolean>(false);
    const [useDefaultCookie, setUseDefaultCookie] = useState<boolean>(true);
    const [pixivCookie, setPixivCookie] = useState<string>('');
    const [isFailedUrlsModalVisible, setIsFailedUrlsModalVisible] = useState(false);
    const [failedUrls, setFailedUrls] = useState<Array<{ url: string, reason: string, filename: string }>>([]);
    const [currentTaskForRetry, setCurrentTaskForRetry] = useState<Task | null>(null);
    const [fileTreeData, setFileTreeData] = useState<FileTreeNode[]>([]);
    const [expandedKeys, setExpandedKeys] = useState<string[]>([]);
    const [selectedImage, setSelectedImage] = useState<string | null>(null);
    const [fileTreeLoading, setFileTreeLoading] = useState(false);
    const fileTreeRefreshTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // 优化的文件树状态
    const [fileTreeMap, setFileTreeMap] = useState<Map<string, FileTreeNode>>(new Map());
    const [fileTreeHeight] = useState(500);

    // 加载任务数据
    const loadTasks = async () => {
        try {
            setLoading(true);
            const response = await apiService.getTasks(1, 20, '', 'crawl');
            setTasks(response.tasks || []);
        } catch (error) {
            console.error('加载任务失败:', error);
            message.error('加载任务失败');
            setTasks([]);
        } finally {
            setLoading(false);
        }
    };


    // 加载文件树数据
    const loadFileTree = async (force = false) => {
        // 防止重复加载
        if (fileTreeLoading && !force) {
            console.log('文件树正在加载中，跳过重复请求');
            return;
        }

        try {
            setFileTreeLoading(true);
            console.log('开始加载文件树...');

            // 从后端API获取真实的文件树数据
            const response = await fetch('http://localhost:50052/api/filetree', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({})
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            const backendFileTree = data.data.fileTree;

            // 转换后端数据为前端格式
            const fileTreeData = convertBackendFileTreeToFrontend(backendFileTree);

            // 构建扁平化映射
            const treeMap = buildFileTreeMap(fileTreeData);

            setFileTreeData(fileTreeData);
            setFileTreeMap(treeMap);
            setExpandedKeys(['images']);
            console.log('文件树加载完成，数据:', fileTreeData);
        } catch (error) {
            console.error('加载文件树失败:', error);
            message.error('加载文件树失败');

            // 如果API失败，显示空状态
            const emptyTree: FileTreeNode[] = [{
                key: 'images',
                title: 'images',
                icon: <FolderOutlined style={{ color: '#1890ff', fontSize: '16px' }} />,
                level: 0,
                children: [{
                    key: 'empty',
                    title: '暂无图片文件',
                    icon: <FileOutlined style={{ color: '#8c8c8c', fontSize: '16px' }} />,
                    isLeaf: true,
                    level: 1
                }]
            }];
            setFileTreeData(emptyTree);
        } finally {
            setFileTreeLoading(false);
        }
    };

    // 转换后端文件树数据为前端格式（优化版）
    const convertBackendFileTreeToFrontend = (backendTree: any): FileTreeNode[] => {
        const convertNode = (node: any, level: number): FileTreeNode => {
            const isFolder = node.isFolder || false;
            const isLeaf = !isFolder;

            return {
                key: node.key,
                title: node.title,
                icon: isFolder ?
                    <FolderOutlined style={{ color: '#1890ff', fontSize: '16px' }} /> :
                    getFileIcon(node.title),
                isLeaf: isLeaf,
                level: level,
                filePath: node.filePath,
                fileSize: node.fileSize,
                fileType: node.fileType,
                children: node.children ? node.children.map((child: any) => convertNode(child, level + 1)) : undefined
            };
        };

        // 后端返回的已经是完整的树结构，直接转换
        return [convertNode(backendTree, 0)];
    };

    // 构建扁平化的文件树映射（优化性能）
    const buildFileTreeMap = (treeData: FileTreeNode[]): Map<string, FileTreeNode> => {
        const map = new Map<string, FileTreeNode>();

        const processNode = (node: FileTreeNode) => {
            map.set(node.key, node);
            if (node.children) {
                node.children.forEach(processNode);
            }
        };

        treeData.forEach(processNode);
        return map;
    };



    // 初始化数据
    useEffect(() => {
        loadTasks();
        loadFileTree();
        // 设置默认爬取类型
        setCrawlType('tag');

        // 监控WebSocket连接状态
        const checkConnection = () => {
            const status = wsManager.getConnectionStatus();
            if (!status.isConnected && !status.isConnecting) {
                console.warn('WebSocket连接已断开，尝试重连...');
                wsManager.reconnect();
            }
        };

        // 每60秒检查一次连接状态（减少检查频率）
        const interval = setInterval(checkConnection, 60000);

        // 监听WebSocket消息
        const handleTaskUpdate = (data: any) => {
            console.log('收到任务更新:', data);
            console.log('当前任务列表:', tasks?.map(t => ({ id: t.id, status: t.status, progress: t.progress })));

            // 更新任务状态、进度和图片数量
            setTasks(prev => {
                const updatedTasks = (prev || []).map(task => {
                    console.log('比较任务ID:', task.id, '===', data.task_id, '结果:', task.id === data.task_id);
                    if (task.id === data.task_id) {
                        const updatedTask = { ...task };
                        if (data.status !== undefined) {
                            updatedTask.status = data.status;
                        }
                        if (data.progress !== undefined) {
                            updatedTask.progress = data.progress;
                        }
                        if (data.images_found !== undefined) {
                            updatedTask.images_found = data.images_found;
                        }
                        if (data.images_downloaded !== undefined) {
                            updatedTask.images_downloaded = data.images_downloaded;
                        }
                        console.log('更新任务:', task.id, '状态:', data.status, '进度:', data.progress, '图片数量:', data.images_found);

                        // 如果任务完成，防抖刷新文件树（避免重复刷新）
                        if (data.status === 'completed') {
                            console.log('任务完成，防抖刷新文件树');

                            // 清除之前的定时器
                            if (fileTreeRefreshTimeoutRef.current) {
                                clearTimeout(fileTreeRefreshTimeoutRef.current);
                            }

                            // 设置新的定时器
                            fileTreeRefreshTimeoutRef.current = setTimeout(() => {
                                loadFileTree(true); // 强制刷新
                                fileTreeRefreshTimeoutRef.current = null;
                            }, 2000); // 延迟2秒
                        }

                        return updatedTask;
                    }
                    return task;
                });
                console.log('更新后的任务列表:', updatedTasks.map(t => ({ id: t.id, status: t.status, progress: t.progress, images_found: t.images_found })));
                return updatedTasks;
            });
        };



        const handleLogMessage = (data: any) => {
            console.log('收到日志消息:', data);
            // 检查消息类型
            if (data.type === 'log_message') {
                // 尝试从不同的位置获取数据
                let logData = data.data || data;
                const { task_id, level, message, time } = logData;
                console.log('处理日志消息:', { task_id, level, message, time });
                if (task_id && level && message) {
                    setTaskLogs(prev => ({
                        ...prev,
                        [task_id]: [...(prev[task_id] || []), { level, message, time }]
                    }));
                } else {
                    console.log('日志消息数据不完整:', logData);
                }
            } else if (data.type === 'global_log') {
                // 处理全局日志
                const { level, message, time } = data.data || data;
                console.log('处理全局日志:', { level, message, time });
                if (level && message) {
                    console.log('添加全局日志到状态:', { level, message, time });
                    setGlobalLogs(prev => {
                        const newLogs = [...prev, { level, message, time }];
                        console.log('更新后的全局日志数量:', newLogs.length);
                        return newLogs;
                    });
                } else {
                    console.log('全局日志数据不完整:', { level, message, time });
                }
            } else {
                console.log('未知的日志消息类型:', data.type);
            }
        };

        // 注册WebSocket事件监听器
        wsManager.on('taskUpdate', handleTaskUpdate);
        // wsManager.on('crawlResultsUpdate', handleCrawlResultsUpdate);
        wsManager.on('logMessage', handleLogMessage);

        // 清理函数
        return () => {
            clearInterval(interval);
            wsManager.off('taskUpdate', handleTaskUpdate);
            // wsManager.off('crawlResultsUpdate', handleCrawlResultsUpdate);
            wsManager.off('logMessage', handleLogMessage);

            // 清理文件树刷新定时器
            if (fileTreeRefreshTimeoutRef.current) {
                clearTimeout(fileTreeRefreshTimeoutRef.current);
                fileTreeRefreshTimeoutRef.current = null;
            }
        };
    }, []);

    // 监听文件树数据变化，用于调试
    useEffect(() => {
        console.log('文件树数据变化:', fileTreeData);
    }, [fileTreeData]);

    // 优化的文件树节点组件
    const FileTreeNodeComponent = React.memo(({ node, isVisible }: { node: FileTreeNode, isVisible: boolean }) => {
        const isExpanded = expandedKeys.includes(node.key);
        const isSelected = selectedImage === node.filePath;

        const handleToggle = () => {
            if (node.children && node.children.length > 0) {
                const newExpandedKeys = isExpanded
                    ? expandedKeys.filter(key => key !== node.key)
                    : [...expandedKeys, node.key];
                setExpandedKeys(newExpandedKeys);
            }
        };

        const handleSelect = () => {
            if (node.isLeaf && node.filePath) {
                setSelectedImage(node.filePath);
            }
        };

        if (!isVisible) return null;

        return (
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    padding: '4px 8px',
                    marginLeft: `${(node.level || 0) * 20}px`,
                    borderRadius: '4px',
                    cursor: node.isLeaf ? 'pointer' : 'default',
                    backgroundColor: isSelected ? '#e6f7ff' : 'transparent',
                    transition: 'all 0.2s ease',
                    height: '32px',
                    minHeight: '32px'
                }}
                onClick={handleSelect}
                onMouseEnter={(e) => {
                    if (!isSelected) {
                        e.currentTarget.style.backgroundColor = '#f5f5f5';
                    }
                }}
                onMouseLeave={(e) => {
                    if (!isSelected) {
                        e.currentTarget.style.backgroundColor = 'transparent';
                    }
                }}
            >
                {/* 展开/收起按钮 */}
                {node.children && node.children.length > 0 && (
                    <div
                        onClick={(e) => {
                            e.stopPropagation();
                            handleToggle();
                        }}
                        style={{
                            width: '16px',
                            height: '16px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            marginRight: '8px',
                            cursor: 'pointer',
                            borderRadius: '2px',
                            transition: 'all 0.2s ease'
                        }}
                    >
                        {isExpanded ? '▼' : '▶'}
                    </div>
                )}

                {/* 图标 */}
                <div style={{ marginRight: '8px', display: 'flex', alignItems: 'center' }}>
                    {node.icon}
                </div>

                {/* 标题和文件信息 */}
                <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <span style={{
                        fontSize: '14px',
                        color: node.isLeaf ? '#333' : '#1890ff',
                        fontWeight: node.isLeaf ? 'normal' : '500'
                    }}>
                        {node.title}
                    </span>

                    {node.isLeaf && node.fileSize && (
                        <span style={{
                            fontSize: '12px',
                            color: '#999',
                            marginLeft: '8px'
                        }}>
                            {formatFileSize(node.fileSize)}
                        </span>
                    )}
                </div>
            </div>
        );
    });

    // 简化的文件树容器
    const VirtualFileTree = React.memo(() => {
        console.log('VirtualFileTree 渲染:', {
            fileTreeData,
            fileTreeMap: Array.from(fileTreeMap.entries()),
            expandedKeys
        });

        // 直接渲染文件树数据，不使用虚拟滚动
        const renderTreeNodes = (nodes: FileTreeNode[]) => {
            return nodes.map((node) => (
                <div key={node.key}>
                    <FileTreeNodeComponent
                        node={node}
                        isVisible={true}
                    />
                    {expandedKeys.includes(node.key) && node.children && (
                        <div style={{ marginLeft: '20px' }}>
                            {renderTreeNodes(node.children)}
                        </div>
                    )}
                </div>
            ));
        };

        return (
            <div
                style={{
                    height: `${fileTreeHeight}px`,
                    overflow: 'auto',
                    border: '1px solid #d9d9d9',
                    borderRadius: '6px',
                    backgroundColor: '#fafafa'
                }}
            >
                <div style={{ padding: '8px' }}>
                    {fileTreeData.length > 0 ? renderTreeNodes(fileTreeData) : (
                        <div style={{ textAlign: 'center', color: '#999', padding: '20px' }}>
                            暂无文件
                        </div>
                    )}
                </div>
            </div>
        );
    });

    const taskColumns = [
        {
            title: '任务名称',
            dataIndex: 'name',
            key: 'name',
            render: (text: string, record: Task) => (
                <Space>
                    <Text strong>{text || `${record.type}任务-${record.id.slice(-8)}`}</Text>
                    {record.status === 'running' && <Badge status="processing" />}
                    {record.status === 'completed' && <Badge status="success" />}
                    {record.status === 'failed' && <Badge status="error" />}
                </Space>
            )
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status: TaskStatus, record: Task) => {
                // 检查是否为部分完成状态（获取数量 > 下载数量）
                const isPartialComplete = record.status === 'completed' &&
                    (record.images_found || 0) > 0 &&
                    (record.images_downloaded || 0) > 0 &&
                    (record.images_found || 0) > (record.images_downloaded || 0);

                const statusMap = {
                    pending: { color: 'default', text: '等待中' },
                    running: { color: 'processing', text: '运行中' },
                    completed: { color: isPartialComplete ? 'warning' : 'success', text: isPartialComplete ? '部分完成' : '已完成' },
                    failed: { color: 'error', text: '失败' },
                    cancelled: { color: 'warning', text: '已取消' }
                };
                const config = statusMap[status] || { color: 'default', text: status };
                return <Tag color={config.color}>{config.text}</Tag>;
            }
        },
        {
            title: '进度',
            dataIndex: 'progress',
            key: 'progress',
            render: (progress: number, record: Task) => (
                <div>
                    <Progress
                        percent={progress}
                        size="small"
                        status={record.status === 'failed' ? 'exception' : 'active'}
                    />
                </div>
            )
        },
        {
            title: '图片数量',
            key: 'images',
            render: (record: Task) => (
                <Space direction="vertical" size="small">
                    <div>
                        <Text type="secondary">获取: </Text>
                        <Text strong style={{ color: '#1890ff' }}>
                            {record.images_found || 0}
                        </Text>
                    </div>
                    <div>
                        <Text type="secondary">下载: </Text>
                        <Text strong style={{ color: '#52c41a' }}>
                            {record.images_downloaded || 0}
                        </Text>
                    </div>
                </Space>
            )
        },
        {
            title: '开始时间',
            dataIndex: 'created_at',
            key: 'created_at',
            render: (time: string) => new Date(time).toLocaleString()
        },
        {
            title: '操作',
            key: 'actions',
            render: (_: any, record: Task) => (
                <Space>
                    {record.status === 'running' && (
                        <Tooltip title="暂停">
                            <Button
                                type="text"
                                icon={<PauseCircleOutlined />}
                                onClick={() => handlePauseTask(record.id)}
                            />
                        </Tooltip>
                    )}
                    {record.status === 'cancelled' && (
                        <Tooltip title="继续">
                            <Button
                                type="text"
                                icon={<PlayCircleOutlined />}
                                onClick={() => handleResumeTask(record.id)}
                            />
                        </Tooltip>
                    )}
                    <Tooltip title="查看详情">
                        <Button
                            type="text"
                            icon={<EyeOutlined />}
                            onClick={() => handleViewDetail(record)}
                        />
                    </Tooltip>
                    {/* 部分完成状态：显示查看失败URL按钮 */}
                    {record.status === 'completed' &&
                        (record.images_found || 0) > 0 &&
                        (record.images_downloaded || 0) > 0 &&
                        (record.images_found || 0) > (record.images_downloaded || 0) && (
                            <Tooltip title="查看失败URL">
                                <Button
                                    type="text"
                                    icon={<ExclamationCircleOutlined />}
                                    onClick={() => handleViewFailedUrls(record)}
                                />
                            </Tooltip>
                        )}
                    {(record.status === 'completed' || record.status === 'failed') && (
                        <Tooltip title="重新运行">
                            <Button
                                type="text"
                                icon={<ReloadOutlined />}
                                onClick={() => handleRerunTask(record)}
                            />
                        </Tooltip>
                    )}
                    <Tooltip title="停止">
                        <Button
                            type="text"
                            danger
                            icon={<StopOutlined />}
                            onClick={() => handleStopTask(record.id)}
                        />
                    </Tooltip>
                </Space>
            )
        }
    ];


    const handlePauseTask = (taskId: string) => {
        setTasks(prev => (prev || []).map(task =>
            task.id === taskId ? { ...task, status: 'cancelled' as TaskStatus } : task
        ));
        message.success('任务已暂停');
    };

    const handleResumeTask = (taskId: string) => {
        setTasks(prev => (prev || []).map(task =>
            task.id === taskId ? { ...task, status: 'running' as TaskStatus } : task
        ));
        message.success('任务已继续');
    };

    const handleStopTask = async (taskId: string) => {
        try {
            // 调用后端API停止任务
            await apiService.stopTask(taskId);
            message.success('任务已停止');
        } catch (error) {
            console.error('停止任务失败:', error);
            message.error('停止任务失败');
        }
    };

    const handleViewDetail = (task: Task) => {
        setSelectedTask(task);
        setIsDetailModalVisible(true);
    };

    const handleRerunTask = async (task: Task) => {
        try {
            // 先更新本地状态，清除错误信息
            setTasks(prev => (prev || []).map(t =>
                t.id === task.id
                    ? { ...t, status: 'pending', error_message: undefined, progress: 0 }
                    : t
            ));

            // 直接重新启动原任务
            await apiService.startTask(task.id);
            message.success('任务重新运行成功');
        } catch (error) {
            console.error('重新运行任务失败:', error);
            message.error('重新运行任务失败');
        }
    };

    const handleViewFailedUrls = async (task: Task) => {
        try {
            // 模拟获取失败URL数据（实际应该从后端API获取）
            const mockFailedUrls = [
                { url: 'https://i.pximg.net/img-original/img/2025/01/18/18/55/57/126337059_p0.png', reason: '代理连接失败', filename: 'artworks_126337059_p01.png' },
                { url: 'https://i.pximg.net/img-original/img/2025/01/18/18/55/57/126337059_p1.png', reason: '代理连接失败', filename: 'artworks_126337059_p02.png' },
                { url: 'https://i.pximg.net/img-original/img/2024/12/29/10/44/31/125653483_p0.png', reason: '代理连接失败', filename: 'artworks_125653483_p01.png' },
                { url: 'https://i.pximg.net/img-original/img/2024/12/29/10/44/31/125653483_p1.png', reason: '代理连接失败', filename: 'artworks_125653483_p02.png' }
            ];

            setFailedUrls(mockFailedUrls);
            setCurrentTaskForRetry(task);
            setIsFailedUrlsModalVisible(true);
        } catch (error) {
            console.error('获取失败URL失败:', error);
            message.error('获取失败URL失败');
        }
    };

    const handleRetryFailedDownloads = async () => {
        if (!currentTaskForRetry) return;

        try {
            // 重新运行任务以下载失败的图片
            await apiService.startTask(currentTaskForRetry.id);
            message.success('正在重新下载失败的图片...');
            setIsFailedUrlsModalVisible(false);
        } catch (error) {
            console.error('重新下载失败:', error);
            message.error('重新下载失败');
        }
    };


    const formatFileSize = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    // 根据文件扩展名获取图标
    const getFileIcon = (filename: string, isFolder: boolean = false) => {
        if (isFolder) {
            return <FolderOutlined style={{ color: '#1890ff' }} />;
        }

        const ext = filename.split('.').pop()?.toLowerCase();
        const iconStyle = { fontSize: '14px', marginRight: '6px' };

        switch (ext) {
            case 'jpg':
            case 'jpeg':
                return <FileJpgOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
            case 'png':
                return <FileImageOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
            case 'gif':
                return <FileGifOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
            case 'svg':
                return <FileImageOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
            case 'webp':
                return <FileImageOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
            case 'pdf':
                return <FilePdfOutlined style={{ ...iconStyle, color: '#ff4d4f' }} />;
            case 'doc':
            case 'docx':
                return <FileWordOutlined style={{ ...iconStyle, color: '#1890ff' }} />;
            case 'xls':
            case 'xlsx':
                return <FileExcelOutlined style={{ ...iconStyle, color: '#52c41a' }} />;
            case 'ppt':
            case 'pptx':
                return <FilePptOutlined style={{ ...iconStyle, color: '#fa8c16' }} />;
            case 'txt':
            case 'md':
                return <FileTextOutlined style={{ ...iconStyle, color: '#722ed1' }} />;
            case 'zip':
            case 'rar':
            case '7z':
                return <FileZipOutlined style={{ ...iconStyle, color: '#faad14' }} />;
            case 'mp4':
            case 'avi':
            case 'mov':
                return <PlayCircleOutlined style={{ ...iconStyle, color: '#eb2f96' }} />;
            case 'mp3':
            case 'wav':
                return <PlayCircleOutlined style={{ ...iconStyle, color: '#13c2c2' }} />;
            case 'js':
            case 'ts':
            case 'jsx':
            case 'tsx':
            case 'html':
            case 'css':
                return <CodeOutlined style={{ ...iconStyle, color: '#722ed1' }} />;
            default:
                return <FileOutlined style={{ ...iconStyle, color: '#8c8c8c' }} />;
        }
    };


    const handleCreateTask = () => {
        setIsModalVisible(true);
    };


    const handleBatchExport = () => {
        try {
            console.log('导出数据 - 当前结果数量:', results?.length || 0);
            console.log('导出数据 - 当前结果:', results);

            if (!results || results.length === 0) {
                message.warning('没有数据可导出');
                return;
            }

            // 创建CSV数据
            const csvData = results.map(result => ({
                id: result.id,
                title: result.title,
                author: result.author,
                tags: result.tags?.join(',') || '',
                imageUrl: result.url || '',
                size: `${result.width}x${result.height}`,
                createdAt: result.created_at || ''
            }));

            // 转换为CSV格式
            const headers = ['ID', '标题', '作者', '标签', '图片链接', '大小', '创建时间'];
            const csvContent = [
                headers.join(','),
                ...csvData.map(row => [
                    row.id,
                    `"${row.title}"`,
                    `"${row.author}"`,
                    `"${row.tags}"`,
                    `"${row.imageUrl}"`,
                    `"${row.size}"`,
                    `"${row.createdAt}"`
                ].join(','))
            ].join('\n');

            // 创建下载链接
            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = `pixiv_images_${new Date().toISOString().split('T')[0]}.csv`;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);

            // 清理URL对象
            URL.revokeObjectURL(link.href);

            message.success(`数据导出成功！共导出 ${results.length} 条记录`);
            console.log('导出完成，文件大小:', blob.size, 'bytes');
        } catch (error) {
            console.error('导出数据失败:', error);
            message.error('导出数据失败');
        }
    };

    const handleBatchDelete = () => {
        console.log('批量删除 - 当前结果数量:', results?.length || 0);
        console.log('批量删除 - 当前结果:', results);

        Modal.confirm({
            title: '批量删除确认',
            content: `确定要删除所有 ${results?.length || 0} 张图片吗？此操作不可撤销。`,
            okText: '确定删除',
            cancelText: '取消',
            okType: 'danger',
            onOk: () => {
                try {
                    console.log('执行批量删除...');
                    setResults([]);
                    message.success('所有图片已删除');
                    console.log('批量删除完成');
                } catch (error) {
                    console.error('批量删除失败:', error);
                    message.error('批量删除失败');
                }
            }
        });
    };

    const handleCleanupTasks = () => {
        const completedTasks = tasks?.filter(task => task.status === 'completed').length || 0;
        const failedTasks = tasks?.filter(task => task.status === 'failed').length || 0;
        const totalTasks = tasks?.length || 0;

        const cleanupModal = Modal.confirm({
            title: '清理任务',
            content: (
                <div>
                    <p>选择要清理的任务类型：</p>
                    <p>• 已完成任务: {completedTasks} 个</p>
                    <p>• 失败任务: {failedTasks} 个</p>
                    <p>• 所有任务: {totalTasks} 个</p>
                </div>
            ),
            okText: '清理已完成',
            cancelText: '取消',
            okType: 'danger',
            onOk: async () => {
                try {
                    // 调用后端API清理已完成的任务
                    const result = await apiService.cleanupTasks('completed');
                    // 重新加载任务列表
                    await loadTasks();
                    message.success(`已清理 ${result.cleaned_count} 个已完成的任务`);
                    // 关闭模态框
                    cleanupModal.destroy();
                } catch (error) {
                    console.error('清理任务失败:', error);
                    message.error('清理任务失败');
                }
            },
            footer: [
                <Button key="cancel" onClick={() => cleanupModal.destroy()}>
                    取消
                </Button>,
                <Button
                    key="failed"
                    danger
                    onClick={async () => {
                        try {
                            // 调用后端API清理失败的任务
                            const result = await apiService.cleanupTasks('failed');
                            // 重新加载任务列表
                            await loadTasks();
                            message.success(`已清理 ${result.cleaned_count} 个失败的任务`);
                            // 关闭模态框
                            cleanupModal.destroy();
                        } catch (error) {
                            console.error('清理失败任务失败:', error);
                            message.error('清理失败任务失败');
                        }
                    }}
                >
                    清理失败任务
                </Button>,
                <Button
                    key="all"
                    danger
                    onClick={() => {
                        const confirmModal = Modal.confirm({
                            title: '确认清理所有任务',
                            content: '确定要清理所有任务吗？此操作不可撤销。',
                            okText: '确定清理',
                            cancelText: '取消',
                            okType: 'danger',
                            onOk: async () => {
                                try {
                                    // 调用后端API清理所有任务
                                    const result = await apiService.cleanupTasks('all');
                                    // 重新加载任务列表
                                    await loadTasks();
                                    message.success(`已清理所有 ${result.cleaned_count} 个任务`);
                                    // 关闭确认模态框
                                    confirmModal.destroy();
                                    // 关闭主模态框
                                    cleanupModal.destroy();
                                } catch (error) {
                                    console.error('清理所有任务失败:', error);
                                    message.error('清理所有任务失败');
                                }
                            }
                        });
                    }}
                >
                    清理所有任务
                </Button>
            ]
        });
    };

    const handleModalOk = async () => {
        try {
            const values = await form.validateFields();

            // 调试：检查数据类型
            console.log('表单值:', values);
            console.log('limit类型:', typeof values.limit, '值:', values.limit);
            console.log('delay类型:', typeof values.delay, '值:', values.delay);

            // 确保数据类型正确 - 强制转换为数字
            const limit = parseInt(String(values.limit)) || 100;
            const delay = parseInt(String(values.delay)) || 1;

            console.log('转换后 - limit类型:', typeof limit, '值:', limit);
            console.log('转换后 - delay类型:', typeof delay, '值:', delay);

            // 创建爬虫请求
            const crawlRequest: CrawlRequest = {
                type: values.type as CrawlType,
                query: values.query || values.user_id?.toString() || values.illust_id?.toString() || '',
                user_id: values.user_id ? parseInt(String(values.user_id)) : undefined,
                illust_id: values.illust_id ? parseInt(String(values.illust_id)) : undefined,
                order: values.order as Order,
                mode: values.mode as Mode,
                limit: limit,
                delay: delay,
                proxy_enabled: proxyEnabled,
                proxy_url: proxyEnabled ? `http://${proxyUrl}` : undefined,
                cookie: useCookie ? (useDefaultCookie ? 'default' : pixivCookie) : undefined
            };

            // 调试：打印请求数据
            console.log('发送的爬虫请求:', crawlRequest);

            // 调用API创建任务
            const newTask = await apiService.createCrawlTask(crawlRequest);

            // 更新本地状态
            setTasks(prev => [newTask, ...(prev || [])]);
            setIsModalVisible(false);
            form.resetFields();
            setProxyEnabled(false);
            setProxyUrl('127.0.0.1:7890');
            message.success('任务创建成功');
        } catch (error) {
            console.error('创建任务失败:', error);
            message.error('创建任务失败');
        }
    };

    const runningTasks = tasks?.filter(task => task.status === 'running').length || 0;
    const completedTasks = tasks?.filter(task => task.status === 'completed').length || 0;
    const totalCrawled = results?.length || 0;

    return (
        <div>
            {/* 统计卡片 */}
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col xs={24} sm={8}>
                    <Card>
                        <Statistic
                            title="运行中任务"
                            value={runningTasks}
                            valueStyle={{ color: '#1890ff' }}
                            prefix={<PlayCircleOutlined />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={8}>
                    <Card>
                        <Statistic
                            title="已完成任务"
                            value={completedTasks}
                            valueStyle={{ color: '#52c41a' }}
                            prefix={<ReloadOutlined />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={8}>
                    <Card>
                        <Statistic
                            title="总爬取数量"
                            value={totalCrawled}
                            valueStyle={{ color: '#faad14' }}
                            prefix={<DownloadOutlined />}
                        />
                    </Card>
                </Col>
            </Row>

            <Row gutter={[24, 24]} style={{ height: '500px' }}>
                {/* 任务管理 */}
                <Col xs={24} lg={14} style={{ height: '100%' }}>
                    <Card
                        title="📋 爬取任务"
                        extra={
                            <Space>
                                <Button
                                    type="primary"
                                    icon={<PlusOutlined />}
                                    onClick={handleCreateTask}
                                >
                                    新建任务
                                </Button>
                                <Button
                                    icon={<ReloadOutlined />}
                                    onClick={loadTasks}
                                    loading={loading}
                                >
                                    刷新
                                </Button>
                                <Button
                                    icon={<DeleteOutlined />}
                                    onClick={handleCleanupTasks}
                                    danger
                                >
                                    清理任务
                                </Button>
                            </Space>
                        }
                        style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
                        bodyStyle={{ flex: 1, padding: 0, overflow: 'hidden' }}
                    >
                        <Spin spinning={loading} style={{ height: '100%' }}>
                            <div style={{ height: '100%', overflow: 'auto' }}>
                                <Table
                                    columns={taskColumns}
                                    dataSource={tasks || []}
                                    rowKey="id"
                                    pagination={false}
                                    size="small"
                                    scroll={{ y: 400 }}
                                />
                            </div>
                        </Spin>
                    </Card>
                </Col>

                {/* 实时日志 */}
                <Col xs={24} lg={10} style={{ height: '100%' }}>
                    <Card
                        title="📊 实时日志"
                        style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
                        bodyStyle={{ flex: 1, padding: '16px', overflow: 'hidden' }}
                    >
                        <div style={{ height: '100%', overflow: 'auto' }}>
                            <Timeline
                                items={globalLogs.slice(-50).map((log, index) => ({
                                    key: index,
                                    color: log.level === 'error' ? 'red' : log.level === 'warning' ? 'orange' : 'blue',
                                    children: (
                                        <div>
                                            <Text type="secondary" style={{ fontSize: '12px' }}>
                                                {log.time}
                                            </Text>
                                            <br />
                                            <Text
                                                type={log.level === 'error' ? 'danger' : undefined}
                                                style={{
                                                    color: log.level === 'error' ? '#ff4d4f' :
                                                        log.level === 'warning' ? '#faad14' : '#1890ff'
                                                }}
                                            >
                                                [{log.level.toUpperCase()}] {log.message}
                                            </Text>
                                        </div>
                                    )
                                }))}
                            />
                        </div>
                    </Card>
                </Col>
            </Row>

            {/* 爬取结果 - 文件树视图 */}
            <Card
                title="🖼️ 爬取结果"
                style={{ marginTop: 24 }}
                extra={
                    <Space>
                        <Search
                            placeholder="搜索文件..."
                            style={{ width: 200 }}
                            prefix={<SearchOutlined />}
                        />
                        <Button
                            type="primary"
                            icon={<DownloadOutlined />}
                            onClick={handleBatchExport}
                        >
                            导出数据
                        </Button>
                        <Button
                            danger
                            icon={<DeleteOutlined />}
                            onClick={handleBatchDelete}
                        >
                            批量删除
                        </Button>
                    </Space>
                }
            >
                <Row gutter={16}>
                    <Col span={8}>
                        <div style={{
                            border: '1px solid #d9d9d9',
                            borderRadius: '6px',
                            padding: '8px',
                            height: '500px',
                            overflow: 'auto',
                            background: 'linear-gradient(135deg, #fafafa 0%, #f5f5f5 100%)'
                        }}>
                            <div style={{
                                padding: '8px 12px',
                                background: '#fafafa',
                                borderBottom: '1px solid #d9d9d9',
                                fontWeight: 'bold',
                                fontSize: '14px'
                            }}>
                                📁 文件树
                            </div>
                            <VirtualFileTree />
                        </div>
                    </Col>
                    <Col span={16}>
                        <div style={{
                            border: '1px solid #d9d9d9',
                            borderRadius: '6px',
                            padding: '16px',
                            height: '500px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            background: 'linear-gradient(135deg, #fafafa 0%, #f0f9ff 100%)',
                            position: 'relative',
                            overflow: 'hidden'
                        }}>
                            {selectedImage ? (
                                <div style={{
                                    textAlign: 'center',
                                    width: '100%',
                                    height: '100%',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}>
                                    <div style={{
                                        position: 'relative',
                                        borderRadius: '12px',
                                        overflow: 'hidden',
                                        boxShadow: '0 8px 32px rgba(0,0,0,0.12)',
                                        background: 'white',
                                        padding: '8px',
                                        maxWidth: '90%',
                                        maxHeight: '80%'
                                    }}>
                                        <img
                                            src={`http://localhost:50052/api/images/${selectedImage}`}
                                            alt="预览"
                                            style={{
                                                maxWidth: '100%',
                                                maxHeight: '400px',
                                                objectFit: 'contain',
                                                borderRadius: '8px',
                                                transition: 'all 0.3s ease',
                                                cursor: 'pointer'
                                            }}
                                            onError={(e) => {
                                                const target = e.target as HTMLImageElement;
                                                target.style.display = 'none';
                                                const parent = target.parentElement;
                                                if (parent) {
                                                    parent.innerHTML = `
                                                        <div style="
                                                            color: #999; 
                                                            font-size: 14px; 
                                                            padding: 40px 20px;
                                                            background: linear-gradient(135deg, #f5f5f5 0%, #e8e8e8 100%);
                                                            border-radius: 8px;
                                                            border: 2px dashed #d9d9d9;
                                                        ">
                                                            <div style="font-size: 48px; margin-bottom: 16px;">🖼️</div>
                                                            <div style="font-weight: 500; margin-bottom: 8px;">无法加载图片预览</div>
                                                            <div style="font-size: 12px; color: #bfbfbf; word-break: break-all;">${selectedImage}</div>
                                                        </div>
                                                    `;
                                                }
                                            }}
                                            onClick={() => {
                                                window.open(`http://localhost:50052/api/images/${selectedImage}`, '_blank');
                                            }}
                                        />
                                    </div>
                                    <div style={{
                                        marginTop: '20px',
                                        color: '#666',
                                        fontSize: '14px',
                                        background: 'rgba(255,255,255,0.8)',
                                        padding: '8px 16px',
                                        borderRadius: '20px',
                                        backdropFilter: 'blur(10px)',
                                        border: '1px solid rgba(255,255,255,0.2)'
                                    }}>
                                        <div style={{ fontWeight: '500', marginBottom: '4px' }}>
                                            {selectedImage.split('/').pop()}
                                        </div>
                                        <div style={{ fontSize: '12px', color: '#999' }}>
                                            点击图片可在新窗口中查看
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <div style={{
                                    color: '#999',
                                    textAlign: 'center',
                                    width: '100%',
                                    height: '100%',
                                    display: 'flex',
                                    flexDirection: 'column',
                                    alignItems: 'center',
                                    justifyContent: 'center'
                                }}>
                                    <div style={{
                                        background: 'linear-gradient(135deg, #f0f9ff 0%, #e6f7ff 100%)',
                                        borderRadius: '50%',
                                        width: '120px',
                                        height: '120px',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        marginBottom: '24px',
                                        boxShadow: '0 4px 16px rgba(24, 144, 255, 0.1)'
                                    }}>
                                        <div style={{ fontSize: '48px' }}>📁</div>
                                    </div>
                                    <div style={{
                                        fontSize: '16px',
                                        fontWeight: '500',
                                        marginBottom: '8px',
                                        color: '#666'
                                    }}>
                                        请选择左侧文件树中的图片进行预览
                                    </div>
                                    <div style={{
                                        fontSize: '12px',
                                        color: '#bfbfbf',
                                        background: 'rgba(255,255,255,0.6)',
                                        padding: '4px 12px',
                                        borderRadius: '12px',
                                        backdropFilter: 'blur(10px)'
                                    }}>
                                        支持 PNG, JPG, GIF, SVG 等格式
                                    </div>
                                </div>
                            )}
                        </div>
                    </Col>
                </Row>
            </Card>

            {/* 创建任务模态框 */}
            <Modal
                title="创建爬取任务"
                open={isModalVisible}
                onOk={handleModalOk}
                onCancel={() => setIsModalVisible(false)}
                width={600}
            >
                <Form form={form} layout="vertical">
                    <Form.Item
                        name="type"
                        label="爬取类型"
                        rules={[{ required: true, message: '请选择爬取类型' }]}
                        initialValue="tag"
                    >
                        <Select onChange={(value) => setCrawlType(value)}>
                            <Option value="tag">标签爬取</Option>
                            <Option value="user">用户爬取</Option>
                            <Option value="illust">插画爬取</Option>
                        </Select>
                    </Form.Item>

                    {crawlType === 'tag' && (
                        <Form.Item
                            name="query"
                            label="标签名"
                            rules={[{ required: true, message: '请输入标签名' }]}
                        >
                            <Input
                                placeholder="请输入标签名，如：1girl, anime, landscape"
                                type="text"
                            />
                        </Form.Item>
                    )}

                    {crawlType === 'user' && (
                        <Form.Item
                            name="user_id"
                            label="用户ID"
                            rules={[
                                { required: true, message: '用户爬取时用户ID不能为空' }
                            ]}
                            normalize={(value) => value ? parseInt(value) : undefined}
                        >
                            <Input type="number" placeholder="请输入用户ID，如：107022296" />
                        </Form.Item>
                    )}

                    {crawlType === 'illust' && (
                        <Form.Item
                            name="illust_id"
                            label="插画ID"
                            rules={[
                                { required: true, message: '插画爬取时插画ID不能为空' }
                            ]}
                            normalize={(value) => value ? parseInt(value) : undefined}
                        >
                            <Input type="number" placeholder="请输入插画ID，如：12345678" />
                        </Form.Item>
                    )}

                    <Row gutter={16}>
                        <Col span={12}>
                            <Form.Item
                                name="order"
                                label="排序方式"
                                initialValue="date_d"
                            >
                                <Select>
                                    <Option value="date_d">按日期降序</Option>
                                    <Option value="popular_d">按热度降序</Option>
                                </Select>
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item
                                name="mode"
                                label="内容模式"
                                initialValue="all"
                            >
                                <Select>
                                    <Option value="safe">安全模式</Option>
                                    <Option value="r18">R18模式</Option>
                                    <Option value="all">全部模式</Option>
                                </Select>
                            </Form.Item>
                        </Col>
                    </Row>

                    <Row gutter={16}>
                        <Col span={12}>
                            <Form.Item
                                name="limit"
                                label="爬取数量"
                                initialValue={1000}
                                normalize={(value) => value ? parseInt(value) : 1000}
                            >
                                <Input type="number" placeholder="1000" />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item
                                name="delay"
                                label="延迟(秒)"
                                initialValue={2}
                                normalize={(value) => value ? parseInt(value) : 2}
                            >
                                <Input type="number" placeholder="2" />
                            </Form.Item>
                        </Col>
                    </Row>

                    {/* 代理配置 */}
                    <Row gutter={16}>
                        <Col span={24}>
                            <Form.Item
                                label="代理设置"
                            >
                                <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
                                    <div>
                                        <input
                                            type="checkbox"
                                            checked={proxyEnabled}
                                            onChange={(e) => setProxyEnabled(e.target.checked)}
                                            style={{ marginRight: '8px' }}
                                        />
                                        <span>启用代理</span>
                                    </div>
                                    {proxyEnabled && (
                                        <div style={{ flex: 1, maxWidth: '300px' }}>
                                            <Input
                                                placeholder="127.0.0.1:7890"
                                                value={proxyUrl}
                                                onChange={(e) => setProxyUrl(e.target.value)}
                                                addonBefore="http://"
                                                size="small"
                                            />
                                        </div>
                                    )}
                                </div>
                            </Form.Item>
                        </Col>
                    </Row>

                    {/* Cookie配置 */}
                    <Row gutter={16}>
                        <Col span={24}>
                            <Form.Item
                                label="Pixiv Cookie 配置"
                                help="配置Pixiv网站的Cookie用于身份验证"
                            >
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                                    {/* 是否使用Cookie */}
                                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                        <input
                                            type="checkbox"
                                            checked={useCookie}
                                            onChange={(e) => {
                                                setUseCookie(e.target.checked);
                                                if (!e.target.checked) {
                                                    setUseDefaultCookie(true);
                                                    setPixivCookie('');
                                                }
                                            }}
                                            style={{ marginRight: '8px' }}
                                        />
                                        <span>启用Cookie身份验证</span>
                                    </div>

                                    {/* Cookie配置选项 */}
                                    {useCookie && (
                                        <div style={{ marginLeft: '20px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                            {/* 是否使用默认Cookie */}
                                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                                <input
                                                    type="checkbox"
                                                    checked={useDefaultCookie}
                                                    onChange={(e) => {
                                                        setUseDefaultCookie(e.target.checked);
                                                        if (e.target.checked) {
                                                            setPixivCookie('');
                                                        }
                                                    }}
                                                    style={{ marginRight: '8px' }}
                                                />
                                                <span>使用默认Cookie（来自配置文件）</span>
                                            </div>

                                            {/* 自定义Cookie输入框 */}
                                            {!useDefaultCookie && (
                                                <div>
                                                    <Input.TextArea
                                                        placeholder="PHPSESSID=xxx; __utma=xxx; ..."
                                                        value={pixivCookie}
                                                        onChange={(e) => setPixivCookie(e.target.value)}
                                                        rows={3}
                                                        style={{
                                                            fontFamily: 'monospace',
                                                            fontSize: '12px',
                                                            marginTop: '8px'
                                                        }}
                                                    />
                                                    <div style={{
                                                        fontSize: '12px',
                                                        color: '#666',
                                                        marginTop: '4px'
                                                    }}>
                                                        从浏览器开发者工具中复制Pixiv网站的Cookie
                                                    </div>
                                                </div>
                                            )}

                                            {/* 默认Cookie说明 */}
                                            {useDefaultCookie && (
                                                <div style={{
                                                    fontSize: '12px',
                                                    color: '#52c41a',
                                                    background: '#f6ffed',
                                                    padding: '8px 12px',
                                                    borderRadius: '4px',
                                                    border: '1px solid #b7eb8f'
                                                }}>
                                                    ✓ 将使用配置文件中的默认Cookie进行身份验证
                                                </div>
                                            )}
                                        </div>
                                    )}

                                    {/* 未启用Cookie的说明 */}
                                    {!useCookie && (
                                        <div style={{
                                            fontSize: '12px',
                                            color: '#faad14',
                                            background: '#fffbe6',
                                            padding: '8px 12px',
                                            borderRadius: '4px',
                                            border: '1px solid #ffe58f'
                                        }}>
                                            ⚠️ 未启用Cookie可能导致无法访问Pixiv内容
                                        </div>
                                    )}
                                </div>
                            </Form.Item>
                        </Col>
                    </Row>
                </Form>
            </Modal>

            {/* 任务详情模态框 */}
            <Modal
                title="任务详情"
                open={isDetailModalVisible}
                onCancel={() => setIsDetailModalVisible(false)}
                footer={[
                    <Button key="close" onClick={() => setIsDetailModalVisible(false)}>
                        关闭
                    </Button>
                ]}
                width={800}
            >
                {selectedTask && (
                    <div>
                        <Row gutter={[16, 16]}>
                            <Col span={12}>
                                <div>
                                    <Text strong>任务ID:</Text>
                                    <br />
                                    <Text>{selectedTask.id}</Text>
                                </div>
                            </Col>
                            <Col span={12}>
                                <div>
                                    <Text strong>任务类型:</Text>
                                    <br />
                                    <Tag color="blue">{selectedTask.type}</Tag>
                                </div>
                            </Col>
                        </Row>

                        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
                            <Col span={12}>
                                <div>
                                    <Text strong>任务状态:</Text>
                                    <br />
                                    <Tag color={
                                        selectedTask.status === 'completed' ? 'success' :
                                            selectedTask.status === 'running' ? 'processing' :
                                                selectedTask.status === 'failed' ? 'error' :
                                                    selectedTask.status === 'pending' ? 'default' : 'warning'
                                    }>
                                        {selectedTask.status === 'completed' ? '已完成' :
                                            selectedTask.status === 'running' ? '运行中' :
                                                selectedTask.status === 'failed' ? '失败' :
                                                    selectedTask.status === 'pending' ? '等待中' : '已取消'}
                                    </Tag>
                                </div>
                            </Col>
                            <Col span={12}>
                                <div>
                                    <Text strong>进度:</Text>
                                    <br />
                                    <Progress percent={selectedTask.progress || 0} />
                                </div>
                            </Col>
                        </Row>

                        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
                            <Col span={12}>
                                <div>
                                    <Text strong>创建时间:</Text>
                                    <br />
                                    <Text>{new Date(selectedTask.created_at).toLocaleString()}</Text>
                                </div>
                            </Col>
                            <Col span={12}>
                                <div>
                                    <Text strong>更新时间:</Text>
                                    <br />
                                    <Text>{new Date(selectedTask.updated_at).toLocaleString()}</Text>
                                </div>
                            </Col>
                        </Row>

                        <div style={{ marginTop: 16 }}>
                            <Text strong>任务配置:</Text>
                            <br />
                            <pre style={{
                                background: '#f5f5f5',
                                padding: '12px',
                                borderRadius: '4px',
                                marginTop: '8px',
                                maxHeight: '200px',
                                overflow: 'auto'
                            }}>
                                {selectedTask.config ? JSON.stringify(JSON.parse(selectedTask.config), null, 2) : '{}'}
                            </pre>
                        </div>

                        {selectedTask.error_message && (
                            <div style={{ marginTop: 16 }}>
                                <Text strong>错误信息:</Text>
                                <br />
                                <Text type="danger">{selectedTask.error_message}</Text>
                            </div>
                        )}

                        {/* 实时日志显示 */}
                        <div style={{ marginTop: 16 }}>
                            <Text strong>实时日志:</Text>
                            <div style={{
                                background: '#f5f5f5',
                                padding: '12px',
                                borderRadius: '4px',
                                marginTop: '8px',
                                maxHeight: '300px',
                                overflow: 'auto',
                                fontFamily: 'monospace',
                                fontSize: '12px'
                            }}>
                                {taskLogs[selectedTask.id] && taskLogs[selectedTask.id].length > 0 ? (
                                    taskLogs[selectedTask.id].map((log, index) => (
                                        <div key={index} style={{
                                            marginBottom: '4px',
                                            color: log.level === 'error' ? '#ff4d4f' :
                                                log.level === 'warn' ? '#faad14' :
                                                    log.level === 'info' ? '#1890ff' : '#666'
                                        }}>
                                            <span style={{ color: '#999' }}>[{log.time}]</span>
                                            <span style={{
                                                marginLeft: '8px',
                                                fontWeight: 'bold',
                                                textTransform: 'uppercase'
                                            }}>[{log.level}]</span>
                                            <span style={{ marginLeft: '8px' }}>{log.message}</span>
                                        </div>
                                    ))
                                ) : (
                                    <div style={{ color: '#999', fontStyle: 'italic' }}>
                                        暂无日志信息
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                )}
            </Modal>

            {/* 失败URL弹窗 */}
            <Modal
                title="下载失败的URL"
                open={isFailedUrlsModalVisible}
                onCancel={() => setIsFailedUrlsModalVisible(false)}
                width={800}
                footer={[
                    <Button key="close" onClick={() => setIsFailedUrlsModalVisible(false)}>
                        关闭
                    </Button>,
                    <Button key="retry" type="primary" onClick={handleRetryFailedDownloads}>
                        重新下载
                    </Button>
                ]}
            >
                <div style={{ marginBottom: 16 }}>
                    <Text type="secondary">
                        共 {failedUrls.length} 个下载失败的URL，点击"重新下载"按钮可以重新尝试下载这些图片。
                    </Text>
                </div>

                <Table
                    dataSource={failedUrls}
                    columns={[
                        {
                            title: '文件名',
                            dataIndex: 'filename',
                            key: 'filename',
                            width: 200,
                            render: (filename: string) => (
                                <Text code>{filename}</Text>
                            )
                        },
                        {
                            title: 'URL',
                            dataIndex: 'url',
                            key: 'url',
                            ellipsis: true,
                            render: (url: string) => (
                                <Text copyable={{ text: url }} style={{ fontSize: '12px' }}>
                                    {url}
                                </Text>
                            )
                        },
                        {
                            title: '失败原因',
                            dataIndex: 'reason',
                            key: 'reason',
                            width: 150,
                            render: (reason: string) => (
                                <Tag color="red">{reason}</Tag>
                            )
                        }
                    ]}
                    pagination={false}
                    size="small"
                    scroll={{ y: 400 }}
                />
            </Modal>
        </div>
    );
};

export default CrawlerPage;
