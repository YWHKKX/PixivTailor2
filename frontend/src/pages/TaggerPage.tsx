import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
    Row,
    Col,
    Card,
    Input,
    Button,
    Select,
    Progress,
    Space,
    Typography,
    message,
    Tag,
    Table,
    Tooltip,
    Badge,
    Spin,
    Modal,
    Form,
    InputNumber,
    Divider
} from 'antd';
import {
    PlayCircleOutlined,
    StopOutlined,
    UploadOutlined,
    ClearOutlined,
    EyeOutlined,
    DeleteOutlined,
    ReloadOutlined,
    FileImageOutlined,
    FolderOutlined,
    FileTextOutlined
} from '@ant-design/icons';
import { apiService } from '../services/api';
import { wsManager } from '../services/websocket';
import { Task, TaskStatus } from '../services/appState';

const { Text } = Typography;
const { Option } = Select;

// 标签请求接口
interface TagRequest {
    input_dir: string | string[]; // 支持单个路径或多个路径
    output_dir: string;
    analyzer: string;
    model?: string;  // 新增模型字段
    skip_tags: string[];
    extend_tags: string[];
    tag_order: string;
    save_type: string;
    limit: number;
}

// 已标签图像接口
interface TaggedImage {
    id: number;
    image_path: string;
    tags: Array<{
        name: string;
        score: number;
        category: string;
        is_general: boolean;
    }>;
    analyzer: string;
    metadata: Record<string, string>;
    created_at: string;
    updated_at: string;
}

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
    level?: number;
}

const TaggerPage: React.FC = () => {
    // ==================== 状态管理 ====================
    const [isTagging, setIsTagging] = useState(false);
    const [progress, setProgress] = useState(0);
    const [status, setStatus] = useState('就绪');
    const [taggedImages, setTaggedImages] = useState<TaggedImage[]>([]);

    // 任务管理
    const [tasks, setTasks] = useState<Task[]>([]);
    const [taskLoading, setTaskLoading] = useState(false);
    const [showTaskManagement, setShowTaskManagement] = useState(false);

    // WebUI 管理
    const [webUIStatus, setWebUIStatus] = useState<string>('stopped');
    const [logStream, setLogStream] = useState<EventSource | null>(null);

    // 文件管理
    const [fileTreeData, setFileTreeData] = useState<FileTreeNode[]>([]);

    // 配置管理
    const [tagRequest, setTagRequest] = useState<TagRequest>({
        input_dir: [],  // 改为数组，支持多路径
        output_dir: '',  // 后端会自动处理输出目录
        analyzer: 'wd14tagger',
        model: 'wd14-convnext-v2',  // 默认模型
        skip_tags: [],
        extend_tags: [],
        tag_order: 'score',
        save_type: 'json',
        limit: 100
    });
    const [inputPaths, setInputPaths] = useState<string[]>([]); // 新增：管理多个输入路径的状态

    // 模态框状态
    const [isDetailModalVisible, setIsDetailModalVisible] = useState(false);
    const [isDirSelectModalVisible, setIsDirSelectModalVisible] = useState(false);
    const [isCharacterExtractModalVisible, setIsCharacterExtractModalVisible] = useState(false);
    const [characterExtractPath, setCharacterExtractPath] = useState('D:\\GolangProject\\PixivTailor2\\backend\\data\\tags');
    const [extractedTags, setExtractedTags] = useState<any>(null);
    const [isExtracting, setIsExtracting] = useState(false);
    const [isCreatingProfile, setIsCreatingProfile] = useState(false);
    const [profileName, setProfileName] = useState('');
    const [profileDescription, setProfileDescription] = useState('');
    const [coreTags, setCoreTags] = useState('');
    const [selectedTask, setSelectedTask] = useState<Task | null>(null);
    const directoryInputRef = useRef<HTMLInputElement>(null);
    const [taskLogs, setTaskLogs] = useState<Record<string, Array<{ level: string, message: string, time: string }>>>({});

    // 表单引用
    const [form] = Form.useForm();

    // ==================== 文件树管理 ====================
    const loadFileTree = useCallback(async () => {
        try {
            const response = await fetch('/api/filetree', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({})
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            const backendFileTree = data.data.fileTree;
            const fileTreeData = convertBackendFileTreeToFrontend(backendFileTree);

            setFileTreeData(fileTreeData);
        } catch (error) {
            console.error('加载文件树失败:', error);
            message.error('加载文件树失败');
            setFileTreeData([]);
        }
    }, []);

    const convertBackendFileTreeToFrontend = useCallback((backendTree: any): FileTreeNode[] => {
        const convertNode = (node: any, level: number): FileTreeNode => {
            const hasChildren = node.children && node.children.length > 0;
            const isFolder = node.isFolder || hasChildren;

            // isLeaf: 没有子节点的才是叶子节点
            const isLeaf = !isFolder && !hasChildren;

            return {
                key: node.key,
                title: node.title,
                icon: isFolder ? <FolderOutlined style={{ color: '#1890ff', fontSize: '16px' }} /> : <FileImageOutlined style={{ color: '#999', fontSize: '16px' }} />,
                isLeaf: isLeaf,
                level: level,
                filePath: node.filePath,
                fileSize: node.fileSize,
                fileType: node.fileType,
                children: node.children && node.children.length > 0 ? node.children.map((child: any) => convertNode(child, level + 1)) : undefined
            };
        };

        // 后端返回的是单个对象，包含children数组
        if (backendTree.children && Array.isArray(backendTree.children)) {
            // 直接返回children数组的转换结果
            return backendTree.children.map((node: any) => convertNode(node, 0));
        } else if (Array.isArray(backendTree)) {
            return backendTree.map(node => convertNode(node, 0));
        } else {
            return [convertNode(backendTree, 0)];
        }
    }, []);

    // ==================== 任务管理 ====================
    const loadTasks = useCallback(async () => {
        try {
            setTaskLoading(true);
            const tasks = await apiService.getTasks();
            const taggerTasks = tasks.filter((task: any) => task.type === 'tag');
            setTasks(taggerTasks);
        } catch (error) {
            console.error('加载任务失败:', error);
            message.error('加载任务失败');
            setTasks([]);
        } finally {
            setTaskLoading(false);
        }
    }, []);

    // 加载标签结果
    const loadTaggedImages = useCallback(async () => {
        try {
            const images = await apiService.getTaggedImages();
            setTaggedImages(images);
            console.log('加载到', images.length, '个标签结果');
        } catch (error) {
            console.error('加载标签结果失败:', error);
        }
    }, []);

    const handleStartTagging = useCallback(async () => {
        // 检查输入目录（支持字符串或数组）
        const inputDirs = Array.isArray(tagRequest.input_dir) ? tagRequest.input_dir : [tagRequest.input_dir];
        if (inputDirs.length === 0 || (inputDirs.length === 1 && !inputDirs[0].trim())) {
            message.warning('请选择至少一个输入目录');
            return;
        }

        // 检查 WebUI 状态
        if (webUIStatus !== 'running' && webUIStatus !== 'external') {
            message.warning('请先启动 WebUI');
            return;
        }

        try {
            setIsTagging(true);
            setProgress(0);
            setStatus('正在生成标签...');
            setTaggedImages([]);

            // 创建标签任务
            const newTask = await apiService.createTagTask(tagRequest);
            setTasks(prev => [newTask, ...prev]);

            message.success('标签生成任务已创建');
            loadTasks();
        } catch (error) {
            console.error('创建标签任务失败:', error);
            message.error('创建标签任务失败');
        } finally {
            setIsTagging(false);
            setProgress(0);
            setStatus('就绪');
        }
    }, [tagRequest, webUIStatus]);

    const handleStopTagging = useCallback(async () => {
        try {
            setIsTagging(false);
            setProgress(0);
            setStatus('已停止');
            message.info('标签生成已停止');
        } catch (error) {
            console.error('停止失败:', error);
        }
    }, []);

    // ==================== 任务管理函数 ====================
    const handleCancelTask = useCallback(async (taskId: string) => {
        try {
            await apiService.cancelTask(taskId);
            message.success('任务已取消');
            loadTasks();
        } catch (error) {
            console.error('取消任务失败:', error);
            message.error('取消任务失败');
        }
    }, []);

    const handleDeleteTask = useCallback(async (taskId: string) => {
        Modal.confirm({
            title: '确认删除',
            content: '确定要删除这个任务吗？此操作不可撤销。',
            okText: '确定删除',
            cancelText: '取消',
            okType: 'danger',
            onOk: async () => {
                try {
                    await apiService.deleteTask(taskId);
                    message.success('任务已删除');
                    loadTasks();
                } catch (error) {
                    console.error('删除任务失败:', error);
                    message.error('删除任务失败');
                }
            }
        });
    }, []);

    const handleBatchDelete = useCallback(() => {
        const completedTasks = tasks?.filter(task => task.status === 'completed').length || 0;
        const failedTasks = tasks?.filter(task => task.status === 'failed').length || 0;

        Modal.info({
            title: '批量删除任务',
            content: (
                <div>
                    <p>选择要删除的任务类型：</p>
                    <div style={{ marginTop: 16 }}>
                        <Button
                            type="primary"
                            danger
                            style={{ marginRight: 8, marginBottom: 8 }}
                            onClick={async () => {
                                try {
                                    const result = await apiService.cleanupTasks('completed');
                                    await loadTasks();
                                    message.success(`已删除 ${result.cleaned_count} 个已完成的任务`);
                                    Modal.destroyAll();
                                } catch (error) {
                                    console.error('批量删除失败:', error);
                                    message.error('批量删除失败');
                                }
                            }}
                            disabled={completedTasks === 0}
                        >
                            删除已完成任务 ({completedTasks} 个)
                        </Button>
                        <Button
                            type="primary"
                            danger
                            onClick={async () => {
                                try {
                                    const result = await apiService.cleanupTasks('failed');
                                    await loadTasks();
                                    message.success(`已删除 ${result.cleaned_count} 个失败的任务`);
                                    Modal.destroyAll();
                                } catch (error) {
                                    console.error('批量删除失败:', error);
                                    message.error('批量删除失败');
                                }
                            }}
                            disabled={failedTasks === 0}
                        >
                            删除失败任务 ({failedTasks} 个)
                        </Button>
                    </div>
                </div>
            ),
            okText: '关闭',
            okButtonProps: { type: 'default' }
        });
    }, [tasks]);

    // ==================== 目录选择 ====================
    const handleOpenDirSelect = useCallback(() => {
        // 打开目录选择模态框
        setIsDirSelectModalVisible(true);
    }, []);

    const handleDirectorySelect = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
        const files = event.target.files;
        if (files && files.length > 0) {
            // 使用第一个文件的 webkitRelativePath 来获取目录路径
            const firstFile = files[0];
            const relativePath = (firstFile as any).webkitRelativePath;
            if (relativePath) {
                // 提取目录路径（移除文件名）
                const dirPath = relativePath.substring(0, relativePath.lastIndexOf('/'));
                // 添加到输入目录列表
                setInputPaths(prev => {
                    if (prev.includes(dirPath)) {
                        message.warning('该目录已存在');
                        return prev;
                    }
                    const newPaths = [...prev, dirPath];
                    setTagRequest(prev => ({ ...prev, input_dir: newPaths }));
                    return newPaths;
                });
                message.success(`已添加目录: ${dirPath}`);
            }
        }
        // 清空input以便可以再次选择
        event.target.value = '';
    }, []);

    const handleSelectDirectory = useCallback((dirPath: string) => {
        console.log(`选择目录: ${dirPath}`);

        // 添加目录到路径列表（如果不存在）
        setInputPaths(prev => {
            if (prev.includes(dirPath)) {
                message.warning('该目录已存在');
                return prev;
            }
            const newPaths = [...prev, dirPath];
            setTagRequest(prev => ({ ...prev, input_dir: newPaths }));
            message.success(`已添加目录: ${dirPath}`);
            return newPaths;
        });
        setIsDirSelectModalVisible(false);
    }, []);

    const handleCloseDirSelect = useCallback(() => {
        setIsDirSelectModalVisible(false);
    }, []);

    // ==================== 角色特征提取 ====================
    const handleOpenCharacterExtract = useCallback(async () => {
        // 默认使用tags目录
        setCharacterExtractPath('D:\\GolangProject\\PixivTailor2\\backend\\data\\tags');
        setIsCharacterExtractModalVisible(true);

        // 自动提取标签
        try {
            setIsExtracting(true);
            const result = await apiService.extractCharacterTags({
                source_dir: 'D:\\GolangProject\\PixivTailor2\\backend\\data\\tags',
                min_frequency: 2,
                min_weight: 0.3,
                max_tags: 30
            });
            setExtractedTags(result);
        } catch (error) {
            console.error('提取标签失败:', error);
            message.error('提取标签失败');
        } finally {
            setIsExtracting(false);
        }
    }, []);

    const handleCloseCharacterExtract = useCallback(() => {
        setIsCharacterExtractModalVisible(false);
        setExtractedTags(null);
        setProfileName('');
        setProfileDescription('');
        setCoreTags('');
    }, []);

    const handleExtractTags = useCallback(async () => {
        if (!characterExtractPath.trim()) {
            message.warning('请选择包含tag文件的目录');
            return;
        }

        try {
            setIsExtracting(true);
            const result = await apiService.extractCharacterTags({
                source_dir: characterExtractPath,
                min_frequency: 2,
                min_weight: 0.3,
                max_tags: 30
            });
            setExtractedTags(result);
            message.success('标签提取成功');
        } catch (error) {
            console.error('提取标签失败:', error);
            message.error('提取标签失败');
        } finally {
            setIsExtracting(false);
        }
    }, [characterExtractPath]);

    const handleCreateCharacterProfile = useCallback(async () => {
        if (!profileName.trim()) {
            message.warning('请输入角色名称');
            return;
        }

        // 解析核心标签（逗号分隔）
        const coreTagsList = coreTags
            .split(',')
            .map(tag => tag.trim())
            .filter(tag => tag.length > 0);

        try {
            setIsCreatingProfile(true);
            await apiService.createCharacterProfile({
                name: profileName,
                description: profileDescription,
                source_dir: characterExtractPath,
                core_tags: coreTagsList,
                min_frequency: 2,
                min_weight: 0.3,
                max_tags: 30
            });
            message.success('角色配置文件已创建');
            handleCloseCharacterExtract();
        } catch (error) {
            console.error('创建角色配置失败:', error);
            message.error('创建角色配置失败');
        } finally {
            setIsCreatingProfile(false);
        }
    }, [characterExtractPath, profileName, profileDescription, coreTags, handleCloseCharacterExtract]);

    // ==================== WebUI 管理 ====================
    const startLogStream = useCallback(() => {
        if (logStream) {
            logStream.close();
        }

        const stream = apiService.createWebUILogStream();
        setLogStream(stream);

        stream.onopen = () => {
            // 连接已打开，静默处理
        };

        stream.onmessage = (event) => {
            // 静默处理日志消息
            event.data; // 避免未使用变量警告
        };

        stream.onerror = () => {
            // 静默处理错误，不输出日志
            if (stream.readyState === EventSource.CLOSED) {
                setLogStream(null);
            }
        };
    }, []);

    const handleStartWebUI = useCallback(async () => {
        try {
            const response = await fetch('http://localhost:50052/api/webui/start-external', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (response.ok) {
                setWebUIStatus('starting');
                message.success('正在启动 WebUI，请查看弹出的命令行窗口');
                // 调用状态监控
                const checkStatus = async () => {
                    try {
                        const status = await apiService.getWebUIStatus();
                        setWebUIStatus(status.status);

                        if (status.status === 'running' || status.status === 'external') {
                            if (!logStream) {
                                startLogStream();
                            }
                        }

                        // 继续检查
                        setTimeout(checkStatus, 3000);
                    } catch (error) {
                        setTimeout(checkStatus, 5000);
                    }
                };
                checkStatus();
            } else {
                message.error('启动 WebUI 失败');
            }
        } catch (error) {
            console.error('启动 WebUI 失败:', error);
            message.error('启动 WebUI 失败');
        }
    }, [logStream, startLogStream]);

    const startWebUIStatusMonitoring = useCallback(async () => {
        const checkStatus = async () => {
            try {
                const status = await apiService.getWebUIStatus();
                setWebUIStatus(status.status);

                if (status.status === 'running' || status.status === 'external') {
                    if (!logStream) {
                        startLogStream();
                    }
                } else if (status.status === 'stopped') {
                    if (logStream) {
                        logStream.close();
                        setLogStream(null);
                    }
                }

                // 无论什么状态都继续检查，实现实时更新
                setTimeout(checkStatus, 3000);
            } catch (error) {
                // 静默处理错误
                setTimeout(checkStatus, 5000);
            }
        };

        checkStatus();
    }, [logStream, startLogStream]); // 添加依赖项

    const handleStopWebUI = useCallback(async () => {
        try {
            const result = await apiService.stopWebUI();
            setWebUIStatus('stopped');
            message.success(result.message);

            if (logStream) {
                logStream.close();
                setLogStream(null);
            }
        } catch (error) {
            console.error('停止 WebUI 失败:', error);
            message.error('停止 WebUI 失败');
        }
    }, [logStream]);

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'running':
            case 'external':
                return '#52c41a';
            case 'stopped':
                return '#ff4d4f';
            default:
                return '#999';
        }
    };

    const getStatusText = (status: string) => {
        switch (status) {
            case 'running':
                return '运行中';
            case 'external':
                return '外部 WebUI';
            case 'stopped':
                return '已停止';
            default:
                return status;
        }
    };

    // ==================== 生命周期 ====================
    useEffect(() => {
        loadFileTree();
        loadTasks();
        loadTaggedImages();
        startWebUIStatusMonitoring();

        // WebSocket监听
        const handleTaskUpdate = (data: any) => {
            console.log('收到任务更新:', data);

            setTasks(prev => {
                const updatedTasks = (prev || []).map(task => {
                    if (task.id === data.task_id) {
                        const updatedTask: any = { ...task };
                        const wasCompleted = task.status === 'completed';

                        if (data.status !== undefined) {
                            updatedTask.status = data.status;
                        }
                        if (data.progress !== undefined) {
                            updatedTask.progress = data.progress;
                        }
                        // 更新 stage 字段（如果存在）
                        if (data.stage !== undefined) {
                            updatedTask.stage = data.stage;
                        }
                        // 更新其他可能的字段
                        if (data.error !== undefined) {
                            updatedTask.error = data.error;
                        }
                        if (data.error_message !== undefined) {
                            updatedTask.error_message = data.error_message;
                        }

                        // 如果任务完成，加载标签结果
                        if (data.status === 'completed' && !wasCompleted) {
                            loadTaggedImages();
                        }

                        return updatedTask;
                    }
                    return task;
                });
                return updatedTasks;
            });
        };

        const handleLogMessage = (data: any) => {
            // data 已经是 log_message 类型的消息体，不需要再检查 type
            const { task_id, level, message, time } = data;
            if (task_id && level && message) {
                console.log('收到日志:', task_id, level, message);
                setTaskLogs(prev => ({
                    ...prev,
                    [task_id]: [...(prev[task_id] || []), { level, message, time }]
                }));
            }
        };

        wsManager.on('task_update', handleTaskUpdate);
        wsManager.on('log_message', handleLogMessage);

        return () => {
            wsManager.off('task_update', handleTaskUpdate);
            wsManager.off('log_message', handleLogMessage);
        };
    }, [loadTasks, loadFileTree, loadTaggedImages, startWebUIStatusMonitoring]);

    // ==================== 渲染函数 ====================
    const renderWebUIControls = () => (
        <Space>
            <Text strong>WebUI状态:</Text>
            <Text style={{ color: getStatusColor(webUIStatus) }}>
                {getStatusText(webUIStatus)}
            </Text>
            {webUIStatus === 'stopped' ? (
                <Button
                    type="primary"
                    icon={<UploadOutlined />}
                    onClick={handleStartWebUI}
                    size="small"
                >
                    启动 WebUI
                </Button>
            ) : (
                <Button
                    danger
                    icon={<StopOutlined />}
                    onClick={handleStopWebUI}
                    size="small"
                >
                    停止 WebUI
                </Button>
            )}
        </Space>
    );

    const renderConfigPanel = () => (
        <Card title="🏷️ 标签生成配置" size="small" extra={renderWebUIControls()}>
            <Form form={form} layout="vertical">
                <Row gutter={[16, 16]}>
                    <Col span={24}>
                        <Form.Item
                            label="输入目录"
                            required
                            extra={
                                <div style={{ fontSize: '12px', color: '#999', marginTop: '4px' }}>
                                    支持多路径选择，可添加多个目录进行批量处理
                                </div>
                            }
                        >
                            <Space direction="vertical" style={{ width: '100%' }} size="small">
                                {/* 显示已选择的路径 */}
                                {inputPaths.length > 0 && (
                                    <div style={{
                                        minHeight: '60px',
                                        padding: '8px',
                                        border: '1px solid #d9d9d9',
                                        borderRadius: '4px',
                                        backgroundColor: '#fafafa'
                                    }}>
                                        {inputPaths.map((path, index) => (
                                            <Tag
                                                key={index}
                                                closable
                                                onClose={() => {
                                                    const newPaths = inputPaths.filter((_, i) => i !== index);
                                                    setInputPaths(newPaths);
                                                    setTagRequest(prev => ({ ...prev, input_dir: newPaths }));
                                                    message.info(`已移除目录: ${path}`);
                                                }}
                                                color="blue"
                                                style={{ marginBottom: '4px' }}
                                            >
                                                {path}
                                            </Tag>
                                        ))}
                                    </div>
                                )}

                                {/* 添加路径按钮 */}
                                <Button
                                    type="default"
                                    icon={<FolderOutlined />}
                                    onClick={handleOpenDirSelect}
                                    style={{ width: '100%' }}
                                >
                                    {inputPaths.length === 0 ? '选择目录' : '添加更多目录'}
                                </Button>
                            </Space>
                        </Form.Item>
                    </Col>
                </Row>

                <Row gutter={[16, 16]}>
                    <Col span={8}>
                        <Form.Item label="分析器">
                            <Select
                                value={tagRequest.analyzer}
                                onChange={(value) => {
                                    setTagRequest(prev => ({ ...prev, analyzer: value }));
                                    // 根据分析器自动调整模型
                                    if (value === 'deepbooru') {
                                        setTagRequest(prev => ({ ...prev, model: '' }));
                                    } else if (value === 'wd14tagger') {
                                        setTagRequest(prev => ({ ...prev, model: 'wd14-convnext-v2' }));
                                    }
                                }}
                            >
                                <Option value="wd14tagger">WD14 Tagger</Option>
                                <Option value="deepbooru">DeepBooru</Option>
                            </Select>
                        </Form.Item>
                    </Col>
                    <Col span={8}>
                        <Form.Item label="模型" tooltip={tagRequest.analyzer === 'wd14tagger' ? "选择 WD14 Tagger 模型版本" : "DeepBooru 使用默认模型"}>
                            <Select
                                value={tagRequest.model || ''}
                                onChange={(value) => setTagRequest(prev => ({ ...prev, model: value }))}
                                disabled={tagRequest.analyzer === 'deepbooru'}
                            >
                                <Option value="wd14-convnext-v2">
                                    ConvNeXt V2 (推荐，CPU兼容)
                                </Option>
                                <Option value="wd14-vit-v2">ViT V2 (需要GPU)</Option>
                                <Option value="wd14-swinv2-v2">Swin V2 (需要GPU)</Option>
                            </Select>
                        </Form.Item>
                    </Col>
                    <Col span={8}>
                        <Form.Item label="保存格式">
                            <Select
                                value={tagRequest.save_type}
                                onChange={(value) => setTagRequest(prev => ({ ...prev, save_type: value }))}
                            >
                                <Option value="txt">TXT文件</Option>
                                <Option value="json">JSON文件</Option>
                            </Select>
                        </Form.Item>
                    </Col>
                </Row>

                <Row gutter={[16, 16]}>
                    <Col span={12}>
                        <Form.Item label="阈值">
                            <InputNumber
                                placeholder="标签置信度阈值 (0.35-1.0)"
                                min={0.35}
                                max={1.0}
                                step={0.05}
                                style={{ width: '100%' }}
                            />
                        </Form.Item>
                    </Col>
                    <Col span={12}>
                        <Form.Item label="处理数量限制">
                            <InputNumber
                                value={tagRequest.limit}
                                onChange={(value) => setTagRequest(prev => ({ ...prev, limit: value || 100 }))}
                                min={1}
                                max={10000}
                                style={{ width: '100%' }}
                            />
                        </Form.Item>
                    </Col>
                </Row>
            </Form>
        </Card>
    );

    const renderControlPanel = () => (
        <Card title="🎮 控制面板" size="small">
            <Space direction="vertical" style={{ width: '100%' }}>
                <Space>
                    <Button
                        type="primary"
                        icon={<PlayCircleOutlined />}
                        onClick={handleStartTagging}
                        loading={isTagging}
                        size="large"
                        disabled={webUIStatus !== 'running' && webUIStatus !== 'external'}
                    >
                        {isTagging ? '生成中...' : '开始生成标签'}
                    </Button>

                    <Button
                        danger
                        icon={<StopOutlined />}
                        onClick={handleStopTagging}
                        disabled={!isTagging}
                        size="large"
                    >
                        停止生成
                    </Button>
                </Space>

                <Space>
                    <Button
                        icon={<ClearOutlined />}
                        onClick={async () => {
                            try {
                                const result = await apiService.deleteAllTaggedImages();
                                setTaggedImages([]);
                                message.success(result.message);
                            } catch (error) {
                                console.error('删除标签文件失败:', error);
                                message.error('删除标签文件失败');
                            }
                        }}
                        size="small"
                    >
                        清空结果
                    </Button>
                    <Button
                        icon={<FileTextOutlined />}
                        onClick={handleOpenCharacterExtract}
                        size="small"
                        type="dashed"
                    >
                        提取角色特征
                    </Button>
                </Space>

                {isTagging && (
                    <div>
                        <Text strong>生成进度:</Text>
                        <Progress percent={progress} status="active" />
                        <Text type="secondary">{status}</Text>
                    </div>
                )}
            </Space>
        </Card>
    );

    return (
        <div style={{ padding: '24px' }}>
            <Typography.Title level={2}>🏷️ 图像标签生成器</Typography.Title>

            <Row gutter={[16, 16]}>
                {/* 左侧配置区域 */}
                <Col span={12}>
                    <Row gutter={[0, 16]}>
                        <Col span={24}>
                            {renderConfigPanel()}
                        </Col>
                        <Col span={24}>
                            {renderControlPanel()}
                        </Col>
                    </Row>
                </Col>

                {/* 右侧结果展示区域 */}
                <Col span={12}>
                    <Card title="📊 生成结果" size="small" style={{ height: '600px', overflow: 'auto' }}>
                        {/* 提示信息 */}
                        <div style={{ marginBottom: 16, padding: 16, background: '#f6ffed', border: '1px solid #b7eb8f', borderRadius: 4 }}>
                            <Text>
                                <span style={{ fontWeight: 'bold' }}>✅ 标签已保存到:</span> <Text code>backend/data/tags/</Text>
                            </Text>
                            <br />
                            <Text type="secondary" style={{ fontSize: '12px' }}>
                                每个图片生成对应的 .json 文件，包含标签、置信度和 WebUI 风格格式
                            </Text>
                        </div>

                        {/* 显示最新生成的标签文件 */}
                        {taggedImages.length > 0 ? (
                            <div>
                                {taggedImages.map((image, index) => (
                                    <Card key={image.id} size="small" style={{ marginBottom: 8 }}>
                                        <Row gutter={8}>
                                            <Col span={8}>
                                                <img
                                                    src={`http://localhost:50052/api/images/${image.image_path}`}
                                                    alt={`Tagged ${index + 1}`}
                                                    style={{ height: 80, width: '100%', objectFit: 'cover', borderRadius: 4 }}
                                                />
                                            </Col>
                                            <Col span={16}>
                                                <div>
                                                    <Text strong style={{ fontSize: '12px' }}>
                                                        {image.image_path.split('/').pop()}
                                                    </Text>
                                                </div>
                                                <div style={{ marginTop: 4 }}>
                                                    {image.tags.slice(0, 5).map(tag => (
                                                        <Tag key={tag.name} color="blue">
                                                            {tag.name} ({tag.score.toFixed(2)})
                                                        </Tag>
                                                    ))}
                                                    {image.tags.length > 5 && (
                                                        <Tag color="default">
                                                            +{image.tags.length - 5} more
                                                        </Tag>
                                                    )}
                                                </div>
                                            </Col>
                                        </Row>
                                    </Card>
                                ))}
                            </div>
                        ) : (
                            <div style={{
                                textAlign: 'center',
                                padding: '40px 20px',
                                color: '#999'
                            }}>
                                <div style={{ fontSize: '48px', marginBottom: '16px', opacity: 0.3 }}>
                                    🏷️
                                </div>
                                <Text type="secondary" style={{ fontSize: '16px' }}>
                                    暂无标签生成结果
                                </Text>
                                <Text type="secondary" style={{ fontSize: '12px', marginTop: '8px' }}>
                                    点击"开始生成标签"开始处理图片
                                </Text>
                            </div>
                        )}
                    </Card>
                </Col>
            </Row>

            {/* 任务管理 */}
            <Card
                title="📋 标签生成任务"
                style={{ marginTop: 24 }}
                extra={
                    <Space>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={loadTasks}
                            loading={taskLoading}
                        >
                            刷新任务
                        </Button>
                        <Button
                            icon={<DeleteOutlined />}
                            onClick={handleBatchDelete}
                            danger
                        >
                            批量删除
                        </Button>
                        <Button
                            type={showTaskManagement ? 'primary' : 'default'}
                            onClick={() => setShowTaskManagement(!showTaskManagement)}
                        >
                            {showTaskManagement ? '隐藏任务管理' : '显示任务管理'}
                        </Button>
                    </Space>
                }
            >
                {showTaskManagement && (
                    <Spin spinning={taskLoading}>
                        <Table
                            columns={[
                                {
                                    title: '任务名称',
                                    dataIndex: 'name',
                                    key: 'name',
                                    render: (text: string, record: Task) => (
                                        <Space>
                                            <Text strong>{text || `标签任务-${record.id.slice(-8)}`}</Text>
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
                                    render: (status: TaskStatus) => {
                                        const statusMap = {
                                            pending: { color: 'default', text: '等待中' },
                                            running: { color: 'processing', text: '运行中' },
                                            completed: { color: 'success', text: '已完成' },
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
                                        <Progress
                                            percent={progress}
                                            size="small"
                                            status={record.status === 'failed' ? 'exception' : 'active'}
                                        />
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
                                            <Tooltip title="查看详情">
                                                <Button
                                                    type="text"
                                                    icon={<EyeOutlined />}
                                                    onClick={() => {
                                                        setSelectedTask(record);
                                                        setIsDetailModalVisible(true);
                                                    }}
                                                />
                                            </Tooltip>
                                            {record.status === 'running' && (
                                                <Tooltip title="取消">
                                                    <Button
                                                        type="text"
                                                        icon={<StopOutlined />}
                                                        onClick={() => handleCancelTask(record.id)}
                                                    />
                                                </Tooltip>
                                            )}
                                            <Tooltip title="删除">
                                                <Button
                                                    type="text"
                                                    icon={<DeleteOutlined />}
                                                    danger
                                                    onClick={() => handleDeleteTask(record.id)}
                                                />
                                            </Tooltip>
                                        </Space>
                                    )
                                }
                            ]}
                            dataSource={tasks || []}
                            rowKey="id"
                            pagination={{
                                pageSize: 10,
                                showSizeChanger: true,
                                showQuickJumper: true,
                                showTotal: (total, range) =>
                                    `第 ${range[0]}-${range[1]} 条/共 ${total} 条`
                            }}
                            size="small"
                            scroll={{ y: 400 }}
                        />
                    </Spin>
                )}
            </Card>

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
                                    <Text strong>任务状态:</Text>
                                    <br />
                                    <Tag color={
                                        selectedTask.status === 'completed' ? 'success' :
                                            selectedTask.status === 'running' ? 'processing' :
                                                selectedTask.status === 'failed' ? 'error' : 'default'
                                    }>
                                        {selectedTask.status}
                                    </Tag>
                                </div>
                            </Col>
                        </Row>

                        {/* 错误信息显示 */}
                        {selectedTask.status === 'failed' && (selectedTask.error || selectedTask.error_message) && (
                            <div style={{ marginTop: 16, marginBottom: 16 }}>
                                <Text strong>失败原因:</Text>
                                <br />
                                <div style={{
                                    background: '#fff2f0',
                                    border: '1px solid #ffccc7',
                                    padding: '12px',
                                    borderRadius: '4px',
                                    marginTop: '8px',
                                    color: '#ff4d4f',
                                    wordBreak: 'break-word'
                                }}>
                                    {selectedTask.error || selectedTask.error_message}
                                </div>
                            </div>
                        )}

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
                                {(selectedTask as any).config ? JSON.stringify(JSON.parse((selectedTask as any).config), null, 2) : '{}'}
                            </pre>
                        </div>

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
                                {(() => {
                                    const logs = selectedTask?.id ? taskLogs[selectedTask.id] : null;
                                    return logs && logs.length > 0 ? (
                                        logs.map((log, index) => (
                                            <div key={index} style={{
                                                marginBottom: '4px',
                                                color: log?.level === 'error' ? '#ff4d4f' :
                                                    (log?.level === 'warning' || log?.level === 'warn') ? '#faad14' :
                                                        log?.level === 'info' ? '#1890ff' : '#666'
                                            }}>
                                                <span style={{ color: '#999' }}>[{log?.time}]</span>
                                                <span style={{
                                                    marginLeft: '8px',
                                                    fontWeight: 'bold',
                                                    textTransform: 'uppercase'
                                                }}>[{log?.level}]</span>
                                                <span style={{ marginLeft: '8px' }}>{log?.message}</span>
                                            </div>
                                        ))
                                    ) : (
                                        <div style={{ color: '#999', fontStyle: 'italic' }}>
                                            暂无日志信息
                                        </div>
                                    );
                                })()}
                            </div>
                        </div>
                    </div>
                )}
            </Modal>

            {/* 目录选择模态框 */}
            <Modal
                title="选择输入目录"
                open={isDirSelectModalVisible}
                onCancel={handleCloseDirSelect}
                footer={[
                    <Button key="cancel" onClick={handleCloseDirSelect}>
                        取消
                    </Button>
                ]}
                width={600}
            >
                <DirectoryTreeModal
                    fileTreeData={fileTreeData}
                    onSelect={(path) => handleSelectDirectory(path)}
                />
            </Modal>

            {/* 隐藏的HTML5目录选择器 */}
            <input
                ref={directoryInputRef as any}
                type="file"
                {...({ webkitdirectory: '', directory: '', multiple: true } as any)}
                style={{ display: 'none' }}
                onChange={handleDirectorySelect}
            />

            {/* 角色特征提取模态框 */}
            <Modal
                title="🎭 提取角色特征"
                open={isCharacterExtractModalVisible}
                onCancel={handleCloseCharacterExtract}
                footer={[
                    <Button key="cancel" onClick={handleCloseCharacterExtract}>
                        取消
                    </Button>,
                    <Button
                        key="save"
                        type="primary"
                        onClick={handleCreateCharacterProfile}
                        loading={isCreatingProfile}
                        disabled={!profileName.trim() || isExtracting}
                    >
                        生成配置文件
                    </Button>
                ]}
                width={800}
            >
                {isExtracting ? (
                    <div style={{ textAlign: 'center', padding: '40px 0' }}>
                        <Spin size="large" />
                        <div style={{ marginTop: 16 }}>正在提取标签...</div>
                    </div>
                ) : (
                    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                        <Text type="secondary">
                            从已生成的标签文件中提取角色特征，生成角色配置文件。
                            这些配置文件可以用于后续的图片生成。
                        </Text>

                        {/* 标签文件目录 */}
                        <div>
                            <Text strong>标签文件目录：</Text>
                            <Input
                                value={characterExtractPath}
                                onChange={(e) => setCharacterExtractPath(e.target.value)}
                                placeholder="例如: D:\GolangProject\PixivTailor2\backend\data\tags"
                                style={{ marginTop: 8 }}
                            />
                        </div>

                        {/* 提取结果 */}
                        {extractedTags && (
                            <Card title="📊 提取结果" size="small">
                                <div>
                                    <Text strong>总文件数：</Text> {extractedTags.total_files || 0}
                                </div>
                                <div style={{ marginTop: 16, maxHeight: '200px', overflow: 'auto' }}>
                                    {Object.entries(extractedTags.tag_frequency || {})
                                        .sort(([, a]: any, [, b]: any) => (b as number) - (a as number))
                                        .slice(0, 30)
                                        .map(([tag, freq]: [string, any]) => (
                                            <Tag key={tag} style={{ marginBottom: 8 }}>
                                                {tag} (频率: {freq}, 权重: {((extractedTags.tag_weights || {})[tag] || 0).toFixed(2)})
                                            </Tag>
                                        ))}
                                </div>
                            </Card>
                        )}

                        {/* 创建角色配置表单 */}
                        <Card title="创建角色配置" size="small">
                            <Space direction="vertical" style={{ width: '100%' }}>
                                <div>
                                    <Text strong>角色名称：</Text>
                                    <Input
                                        placeholder="必填，例如：艾尔莎"
                                        value={profileName}
                                        onChange={(e) => setProfileName(e.target.value)}
                                        style={{ marginTop: 8 }}
                                    />
                                </div>
                                <div>
                                    <Text strong>核心标签：</Text>
                                    <Input
                                        placeholder="可选，逗号分隔，例如：elsa_(frozen), character_name"
                                        value={coreTags}
                                        onChange={(e) => setCoreTags(e.target.value)}
                                        style={{ marginTop: 8 }}
                                    />
                                    <Text type="secondary" style={{ fontSize: '12px', display: 'block', marginTop: 4 }}>
                                        用于指定角色的核心特征标签（如角色名称）
                                    </Text>
                                </div>
                                <div>
                                    <Text strong>角色描述：</Text>
                                    <Input.TextArea
                                        placeholder="可选，对角色进行简要描述"
                                        value={profileDescription}
                                        onChange={(e) => setProfileDescription(e.target.value)}
                                        rows={3}
                                        style={{ marginTop: 8 }}
                                    />
                                </div>
                                <Text type="secondary" style={{ fontSize: '12px' }}>
                                    配置文件将保存在 backend/data/characters/ 目录
                                </Text>
                            </Space>
                        </Card>
                    </Space>
                )}
            </Modal>

        </div>
    );
};

// 目录树组件
const DirectoryTreeModal: React.FC<{
    fileTreeData: FileTreeNode[];
    onSelect: (path: string) => void;
}> = ({ fileTreeData, onSelect }) => {
    const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set());

    const toggleExpand = (key: string) => {
        setExpandedKeys(prev => {
            const newSet = new Set(prev);
            if (newSet.has(key)) {
                newSet.delete(key);
            } else {
                newSet.add(key);
            }
            return newSet;
        });
    };

    const renderDirectoryTree = (nodes: FileTreeNode[], parentPath: string): React.ReactNode => {
        return nodes.map(node => {
            // 计算当前节点的路径
            const nodePath = node.filePath || node.title;
            const currentPath = parentPath ? `${parentPath}/${nodePath}` : nodePath;

            // 显示文件和目录
            const isDirectory = !node.isLeaf;
            const hasChildren = node.children && node.children.length > 0;
            const isExpanded = expandedKeys.has(node.key);

            return (
                <div key={node.key}>
                    {/* 节点行 */}
                    <div
                        style={{
                            padding: '8px',
                            cursor: isDirectory ? 'pointer' : 'default',
                            borderRadius: '4px',
                            marginBottom: '4px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            marginLeft: `${(node.level || 0) * 20}px`,
                            opacity: isDirectory ? 1 : 0.9
                        }}
                        onClick={() => {
                            if (isDirectory) {
                                // 切换展开/折叠
                                if (hasChildren) {
                                    toggleExpand(node.key);
                                } else {
                                    // 没有子节点时，选择该目录
                                    onSelect(currentPath);
                                }
                            }
                        }}
                        onMouseEnter={(e) => {
                            if (isDirectory) {
                                e.currentTarget.style.backgroundColor = '#f5f5f5';
                            }
                        }}
                        onMouseLeave={(e) => {
                            if (isDirectory) {
                                e.currentTarget.style.backgroundColor = 'transparent';
                            }
                        }}
                    >
                        {/* 展开/折叠图标 */}
                        {isDirectory && hasChildren && (
                            <span style={{ width: '16px', display: 'inline-block', textAlign: 'center' }}>
                                {isExpanded ? '▼' : '▶'}
                            </span>
                        )}
                        {isDirectory && !hasChildren && (
                            <span style={{ width: '16px', display: 'inline-block' }}></span>
                        )}

                        {/* 文件/目录图标 */}
                        {isDirectory ? (
                            <FolderOutlined style={{ fontSize: '16px', color: '#1890ff' }} />
                        ) : (
                            <FileImageOutlined style={{ fontSize: '16px', color: '#999' }} />
                        )}

                        <span style={{ flex: 1 }}>{node.title}</span>

                        {/* 目录选择按钮 */}
                        {isDirectory && (
                            <Button
                                type="primary"
                                size="small"
                                style={{ marginLeft: 'auto' }}
                                onClick={(e) => {
                                    e.stopPropagation(); // 阻止事件冒泡
                                    onSelect(currentPath);
                                }}
                            >
                                选择
                            </Button>
                        )}

                        {/* 文件大小 */}
                        {!isDirectory && (
                            <span style={{ fontSize: '12px', color: '#999', marginLeft: 'auto' }}>
                                {node.fileSize ? `(${(node.fileSize / 1024).toFixed(1)} KB)` : ''}
                            </span>
                        )}
                    </div>

                    {/* 图片预览（如果是图片文件） */}
                    {!isDirectory && node.filePath && (
                        <div style={{ marginLeft: `${((node.level || 0) + 1) * 20}px`, marginBottom: '8px' }}>
                            <img
                                src={`http://localhost:50052/api/images/${node.filePath}`}
                                alt={node.title}
                                style={{
                                    width: '100px',
                                    height: '100px',
                                    objectFit: 'cover',
                                    borderRadius: '4px',
                                    cursor: 'pointer',
                                    border: '1px solid #d9d9d9'
                                }}
                                onError={(e) => {
                                    // 图片加载失败时隐藏
                                    e.currentTarget.style.display = 'none';
                                }}
                            />
                        </div>
                    )}

                    {/* 递归显示子节点（如果展开） */}
                    {hasChildren && isExpanded && (
                        <div>
                            {renderDirectoryTree(node.children!, currentPath)}
                        </div>
                    )}
                </div>
            );
        });
    };

    return (
        <div style={{ height: '600px', overflow: 'auto' }}>
            {fileTreeData.length > 0 ? (
                <div>
                    {renderDirectoryTree(fileTreeData, '')}
                </div>
            ) : (
                <div style={{
                    textAlign: 'center',
                    padding: '40px 20px',
                    color: '#999'
                }}>
                    <Spin />
                    <div style={{ marginTop: '16px' }}>正在加载文件树...</div>
                </div>
            )}
        </div>
    );
};

export default TaggerPage;
