import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
    Row,
    Col,
    Card,
    Input,
    InputNumber,
    Button,
    Slider,
    Select,
    Switch,
    Progress,
    Space,
    Typography,
    message,
    Image,
    Tag,
    Table,
    Tooltip,
    Badge,
    Spin,
    Modal
} from 'antd';
import {
    PlayCircleOutlined,
    StopOutlined,
    DownloadOutlined,
    ClearOutlined,
    SaveOutlined,
    CopyOutlined,
    PoweroffOutlined,
    ReloadOutlined,
    EyeOutlined,
    DeleteOutlined
} from '@ant-design/icons';
import { apiService } from '../services/api';
import { wsManager } from '../services/websocket';
import { GenerationParams, LoraConfig, Task, TaskStatus } from '../services/appState';
import { aiService, AIGenerationResult } from '../services/aiService';

const { TextArea } = Input;
const { Text } = Typography;
const { Option } = Select;

// 默认生成参数
const DEFAULT_PARAMS: GenerationParams = {
    prompt: '',
    negative_prompt: '',
    steps: 20,
    cfg_scale: 7.0,
    width: 512,
    height: 512,
    seed: -1,
    model: '',
    sampler: 'DPM++ 2M Karras',
    batch_size: 1,
    batch_count: 1,
    loop_count: 1,
    enable_hr: false,
    hr_scale: 2.0,
    hr_upscaler: 'Latent',
    hr_steps: 0,
    hr_denoising_strength: 0.7
};

const AIGeneratorPage: React.FC = () => {
    // ==================== 缓存管理 ====================
    const CACHE_KEY = 'ai_generator_config';

    const loadCache = (): any => {
        try {
            const cached = localStorage.getItem(CACHE_KEY);
            if (cached) {
                return JSON.parse(cached);
            }
        } catch (error) {
            console.error('加载缓存失败:', error);
        }
        return null;
    };

    const saveCache = (data: any) => {
        try {
            localStorage.setItem(CACHE_KEY, JSON.stringify(data));
        } catch (error) {
            console.error('保存缓存失败:', error);
        }
    };

    // ==================== 状态管理 ====================
    const cachedData = loadCache();
    const [params, setParams] = useState<GenerationParams>(
        cachedData?.params || DEFAULT_PARAMS
    );
    const [isGenerating, setIsGenerating] = useState(false);
    const [progress, setProgress] = useState(0);
    const [generatedImages, setGeneratedImages] = useState<string[]>([]);
    const [status, setStatus] = useState('就绪');

    // 图片查看器状态
    const [imageViewerVisible, setImageViewerVisible] = useState(false);
    const [currentImageIndex, setCurrentImageIndex] = useState(0);
    const [viewerImages, setViewerImages] = useState<string[]>([]);

    // 配置管理
    const [configs, setConfigs] = useState<any[]>([]);
    const [selectedConfig, setSelectedConfig] = useState<string>(cachedData?.selectedConfig || '');
    const [configCategories, setConfigCategories] = useState<string[]>([]);
    const [selectedCategory, setSelectedCategory] = useState<string>(cachedData?.selectedCategory || '');

    // 角色特征管理
    const [characters, setCharacters] = useState<any[]>([]);
    const [selectedCharacters, setSelectedCharacters] = useState<string[]>(cachedData?.selectedCharacters || []);
    // 用户输入的原始提示词（不包含角色特征）
    const [userPrompt, setUserPrompt] = useState<string>(cachedData?.userPrompt || '');

    // WebUI 管理
    const [webUIStatus, setWebUIStatus] = useState<string>('stopped');
    const [logStream, setLogStream] = useState<EventSource | null>(null);

    // LoRA 管理
    const [loras, setLoras] = useState<LoraConfig[]>([]);
    const [selectedLoras, setSelectedLoras] = useState<string[]>(cachedData?.selectedLoras || []);

    // 任务管理
    const [tasks, setTasks] = useState<Task[]>([]);
    const [taskLoading, setTaskLoading] = useState(false);
    const [showTaskManagement, setShowTaskManagement] = useState(false);

    // 任务详情
    const [taskDetailVisible, setTaskDetailVisible] = useState(false);
    const [selectedTask, setSelectedTask] = useState<Task | null>(null);
    const [taskLogs, setTaskLogs] = useState<string[]>([]);
    const [taskDetailLoading, setTaskDetailLoading] = useState(false);
    const logContainerRef = useRef<HTMLDivElement>(null);

    // ==================== 自动保存配置 ====================
    useEffect(() => {
        const dataToCache = {
            params,
            selectedConfig,
            selectedCategory,
            selectedCharacters,
            userPrompt,
            selectedLoras
        };
        saveCache(dataToCache);
    }, [params, selectedConfig, selectedCategory, selectedCharacters, userPrompt, selectedLoras]);

    // ==================== 配置管理 ====================
    const loadConfigs = useCallback(async () => {
        try {
            const configs = await apiService.getConfigs();
            setConfigs(Array.isArray(configs) ? configs : []);
        } catch (error) {
            console.error('加载配置失败:', error);
            setConfigs([]);
        }
    }, []);

    const loadCategories = useCallback(async () => {
        try {
            const categories = await apiService.getConfigCategories();
            setConfigCategories(Array.isArray(categories) ? categories : []);
        } catch (error) {
            console.error('加载分类失败:', error);
            setConfigCategories([]);
        }
    }, []);

    const loadCharacters = useCallback(async () => {
        try {
            const charactersList = await apiService.listCharacterProfiles();
            setCharacters(Array.isArray(charactersList) ? charactersList : []);
        } catch (error) {
            console.error('加载角色配置失败:', error);
            setCharacters([]);
        }
    }, []);

    const applyConfig = useCallback(async (configId: string) => {
        try {
            const config = await apiService.getConfig(configId);
            if (config) {
                // 更新 userPrompt（这是用户输入的纯文本，不包含角色特征）
                setUserPrompt(config.prompt || '');

                setParams(prev => ({
                    ...prev,
                    prompt: config.prompt || '',
                    negative_prompt: config.negative_prompt || '',
                    steps: config.steps || 20,
                    cfg_scale: config.cfg_scale || 7.0,
                    width: config.width || 512,
                    height: config.height || 512,
                    sampler: config.sampler || 'DPM++ 2M Karras',
                    batch_size: config.batch_size || 1,
                    enable_hr: config.enable_hr || false,
                    hr_scale: config.hr_scale || 2.0,
                    hr_upscaler: config.hr_upscaler || 'Latent',
                    hr_steps: config.hr_steps || 0,
                    hr_denoising_strength: config.hr_denoising_strength || 0.7,
                    loras: config.loras || [],
                    vae: config.vae || '',
                    restore_faces: config.restore_faces || false,
                    tiling: config.tiling || false,
                    clip_skip: config.clip_skip || 2
                }));

                // 设置LoRA数据
                if (config.loras && Array.isArray(config.loras)) {
                    setLoras(config.loras);
                    setSelectedLoras(config.loras.map((lora: LoraConfig) => lora.lora_key));
                }

                message.success('配置已应用');
            }
        } catch (error) {
            message.error('应用配置失败');
        }
    }, []);

    // 更新提示词，合并用户输入的提示词和角色特征
    const updatePromptWithCharacters = useCallback((basePrompt: string | null = null) => {
        // 使用传入的参数或当前的 userPrompt
        const currentUserPrompt = basePrompt !== null ? basePrompt : userPrompt;
        console.log('updatePromptWithCharacters - currentUserPrompt:', currentUserPrompt);
        console.log('updatePromptWithCharacters - selectedCharacters:', selectedCharacters);

        // 收集所有选中角色特征的标签
        const selectedTags: string[] = [];

        selectedCharacters.forEach(characterName => {
            const character = characters.find(c => c.name === characterName);
            console.log('processing character:', characterName, character);
            if (character && character.tags && character.tags.length > 0) {
                const tagsString = character.tags.map((tag: string) => {
                    const weight = character.tag_weights?.[tag] || 0;
                    return weight > 0.7 ? `((${tag}:${weight.toFixed(2)}))` : tag;
                }).join(', ');
                selectedTags.push(tagsString);
            }
        });

        console.log('selectedTags:', selectedTags);

        // 构建最终提示词：用户输入 + BREAK + 每个角色特征之间用BREAK分隔
        let finalPrompt = currentUserPrompt;
        if (selectedTags.length > 0) {
            // 每个角色特征之间用BREAK分隔
            const allTagsString = selectedTags.join('\nBREAK\n');
            if (currentUserPrompt) {
                finalPrompt = `${currentUserPrompt}\nBREAK\n${allTagsString}`;
            } else {
                finalPrompt = allTagsString;
            }
        }

        console.log('finalPrompt:', finalPrompt);

        // 更新 params.prompt
        setParams(prev => ({ ...prev, prompt: finalPrompt }));
    }, [selectedCharacters, characters, userPrompt]);

    const resetToDefaultConfig = useCallback(() => {
        // 重置为默认参数
        setUserPrompt('');
        setSelectedCharacters([]);
        setParams(prev => ({
            ...prev,
            prompt: '',
            negative_prompt: '',
            steps: 20,
            cfg_scale: 7.0,
            width: 512,
            height: 512,
            seed: -1,
            model: '',
            sampler: 'DPM++ 2M Karras',
            batch_size: 1,
            batch_count: 1,
            loop_count: 1,
            enable_hr: false,
            hr_scale: 2.0,
            hr_upscaler: 'Latent',
            hr_steps: 0,
            hr_denoising_strength: 0.7,
            loras: prev.loras || [],
            vae: prev.vae || '',
            restore_faces: prev.restore_faces || false,
            tiling: prev.tiling || false,
            clip_skip: prev.clip_skip || 2
        }));

        // 清空LoRA配置
        setLoras([]);
        setSelectedLoras([]);

        message.success('配置已重置为默认值');
    }, []);

    // ==================== 任务管理 ====================
    const loadTasks = useCallback(async () => {
        try {
            setTaskLoading(true);
            const tasks = await apiService.getTasks();
            setTasks(tasks || []);

            // 自动加载最近完成任务的图片
            if (tasks && tasks.length > 0) {
                const completedTasks = tasks.filter(task => task.status === 'completed' && task.result);
                if (completedTasks.length > 0) {
                    // 获取最新的完成任务
                    const latestTask = completedTasks[0];
                    if (latestTask) {
                        try {
                            const result = typeof latestTask.result === 'string'
                                ? JSON.parse(latestTask.result)
                                : latestTask.result;
                            if (result && result.images && Array.isArray(result.images)) {
                                setGeneratedImages(result.images);
                            }
                        } catch (error) {
                            console.error('解析任务结果失败:', error);
                        }
                    }
                }
            }
        } catch (error) {
            console.error('加载任务失败:', error);
            message.error('加载任务失败');
            setTasks([]);
        } finally {
            setTaskLoading(false);
        }
    }, []);

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
        const totalTasks = tasks?.length || 0;

        // 使用Modal.info来显示选择选项
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
                            style={{ marginRight: 8, marginBottom: 8 }}
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
                        <Button
                            type="primary"
                            danger
                            style={{ marginRight: 8, marginBottom: 8 }}
                            onClick={async () => {
                                try {
                                    const result = await apiService.cleanupTasks('all');
                                    await loadTasks();
                                    message.success(`已删除 ${result.cleaned_count} 个任务`);
                                    Modal.destroyAll();
                                } catch (error) {
                                    console.error('批量删除失败:', error);
                                    message.error('批量删除失败');
                                }
                            }}
                            disabled={totalTasks === 0}
                        >
                            删除所有任务 ({totalTasks} 个)
                        </Button>
                    </div>
                </div>
            ),
            okText: '关闭',
            onOk: () => {
                Modal.destroyAll();
            }
        });
    }, [tasks]);

    const handleViewTaskDetail = useCallback(async (taskId: string) => {
        try {
            setTaskDetailLoading(true);
            setTaskDetailVisible(true);

            // 获取任务详情
            const task = await apiService.getTask(taskId);
            setSelectedTask(task);

            // 获取任务日志（这里模拟日志数据，实际应该从后端获取）
            const logs = [
                `任务 ${taskId} 开始执行`,
                `配置加载完成: ${task.config_id || 'N/A'}`,
                `开始调用 WebUI API`,
                `WebUI 响应状态: ${task.status}`,
                `生成进度: ${task.progress || 0}%`,
                task.status === 'completed' ? '任务执行完成' :
                    task.status === 'failed' ? '任务执行失败' : '任务执行中...'
            ];
            setTaskLogs(logs);

        } catch (error) {
            console.error('获取任务详情失败:', error);
            message.error('获取任务详情失败');
        } finally {
            setTaskDetailLoading(false);
        }
    }, []);

    const handleCloseTaskDetail = useCallback(() => {
        setTaskDetailVisible(false);
        setSelectedTask(null);
        setTaskLogs([]);
    }, []);

    // ==================== 图像生成 ====================
    const handleGenerate = useCallback(async () => {
        if (!params.prompt.trim()) {
            message.warning('请输入提示词');
            return;
        }

        if (webUIStatus !== 'running' && webUIStatus !== 'external') {
            message.warning('请先启动 WebUI');
            return;
        }

        try {
            setIsGenerating(true);
            setProgress(0);
            setStatus('正在生成...');
            setGeneratedImages([]);

            let result: AIGenerationResult;

            if (selectedConfig) {
                // 使用配置生成
                result = await aiService.generateWithConfig(selectedConfig, params);
            } else {
                // 直接生成
                result = await aiService.generateImages(params);
            }

            // 检查任务状态
            if (result.status === 'pending' || result.status === 'running') {
                setStatus('任务已创建，正在生成中...');
                message.info('AI生成任务已创建，请等待生成完成');
                // 加载任务列表以显示新任务
                loadTasks();
            } else if (result.images && result.images.length > 0) {
                setGeneratedImages(result.images);
                setStatus('生成完成');
                message.success(`成功生成 ${result.images.length} 张图片`);
                // 生成完成后加载任务列表
                loadTasks();
            } else {
                setStatus('生成失败');
                message.error('生成失败，请检查参数');
            }
        } catch (error) {
            console.error('生成失败:', error);

            // 检查是否是并发限制错误
            if (error instanceof Error && error.message.includes('429')) {
                setStatus('等待中...');
                message.warning('请等待当前生成任务完成后再试');
                // 不设置isGenerating为false，保持等待状态
                return;
            }

            setStatus('生成失败');
            message.error(`生成失败: ${error instanceof Error ? error.message : '未知错误'}`);
        } finally {
            // 只有在非等待状态时才设置为false
            if (status !== '等待中...') {
                setIsGenerating(false);
                setProgress(0);
            }
        }
    }, [params, selectedConfig, webUIStatus]);

    const handleStop = useCallback(async () => {
        try {
            setIsGenerating(false);
            setProgress(0);
            setStatus('已停止');
            message.info('生成已停止');
        } catch (error) {
            console.error('停止失败:', error);
        }
    }, []);

    // ==================== WebUI 管理 ====================
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
                startWebUIStatusMonitoring();
            } else {
                const errorData = await response.json();
                message.error(`启动WebUI失败: ${errorData.message || '未知错误'}`);
            }
        } catch (error) {
            message.error(`启动WebUI失败: ${error instanceof Error ? error.message : '未知错误'}`);
        }
    }, []);

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
            message.error(`停止WebUI失败: ${error instanceof Error ? error.message : '未知错误'}`);
        }
    }, [logStream]);

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

                setTimeout(() => {
                    if (webUIStatus === 'running' && !logStream) {
                        startLogStream();
                    }
                }, 5000);
            }
        };
    }, [logStream, webUIStatus]);

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
    }, [logStream, startLogStream]);

    // ==================== 工具函数 ====================
    const handleRandomSeed = useCallback(() => {
        setParams(prev => ({ ...prev, seed: Math.floor(Math.random() * 1000000) }));
    }, []);

    const handleClearImages = useCallback(() => {
        setGeneratedImages([]);
        message.info('已清空图片');
    }, []);

    // 图片查看器相关函数
    const handleImageClick = useCallback((_imageUrl: string, index: number) => {
        console.log('handleImageClick called:', index, 'Current viewer visible:', imageViewerVisible);

        // 防止重复触发 - 如果查看器已经显示，则不重复打开
        if (imageViewerVisible) {
            console.log('Viewer already visible, skipping');
            return;
        }

        setViewerImages(generatedImages);
        setCurrentImageIndex(index);
        setImageViewerVisible(true);
    }, [generatedImages, imageViewerVisible]);

    const handlePrevImage = useCallback(() => {
        setCurrentImageIndex(prev => prev > 0 ? prev - 1 : viewerImages.length - 1);
    }, [viewerImages.length]);

    const handleNextImage = useCallback(() => {
        setCurrentImageIndex(prev => prev < viewerImages.length - 1 ? prev + 1 : 0);
    }, [viewerImages.length]);

    const handleCloseViewer = useCallback(() => {
        setImageViewerVisible(false);
    }, []);

    // 键盘导航支持
    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            if (!imageViewerVisible) return;

            switch (event.key) {
                case 'ArrowLeft':
                    event.preventDefault();
                    handlePrevImage();
                    break;
                case 'ArrowRight':
                    event.preventDefault();
                    handleNextImage();
                    break;
                case 'Escape':
                    event.preventDefault();
                    handleCloseViewer();
                    break;
            }
        };

        document.addEventListener('keydown', handleKeyDown);
        return () => document.removeEventListener('keydown', handleKeyDown);
    }, [imageViewerVisible, handlePrevImage, handleNextImage, handleCloseViewer]);

    const handleDownloadImage = useCallback((imageUrl: string, index: number) => {
        const link = document.createElement('a');
        link.href = imageUrl;
        link.download = `generated_image_${index + 1}.png`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        message.success('图片下载已开始');
    }, []);

    const handleCopyPrompt = useCallback(() => {
        navigator.clipboard.writeText(params.prompt);
        message.success('提示词已复制到剪贴板');
    }, [params.prompt]);

    // ==================== 任务表格列定义 ====================
    const taskColumns = [
        {
            title: '任务名称',
            dataIndex: 'name',
            key: 'name',
            render: (text: string, record: Task) => (
                <Space>
                    <Text strong>{text || `AI生成任务-${record.id.slice(-8)}`}</Text>
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
            title: '生成数量',
            key: 'images',
            render: (record: Task) => (
                <Space direction="vertical" size="small">
                    <div>
                        <Text type="secondary">生成: </Text>
                        <Text strong style={{ color: '#1890ff' }}>
                            {record.images_generated || 0}
                        </Text>
                    </div>
                    <div>
                        <Text type="secondary">成功: </Text>
                        <Text strong style={{ color: '#52c41a' }}>
                            {record.images_success || 0}
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
                        <Tooltip title="取消">
                            <Button
                                type="text"
                                icon={<StopOutlined />}
                                onClick={() => handleCancelTask(record.id)}
                            />
                        </Tooltip>
                    )}
                    <Tooltip title="查看详情">
                        <Button
                            type="text"
                            icon={<EyeOutlined />}
                            onClick={() => handleViewTaskDetail(record.id)}
                        />
                    </Tooltip>
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
    ];

    // ==================== 生命周期 ====================
    useEffect(() => {
        loadConfigs();
        loadCategories();
        loadCharacters();
        loadTasks(); // 添加任务加载
        startWebUIStatusMonitoring();

        // WebSocket任务更新监听
        const handleTaskUpdate = (data: any) => {
            console.log('收到任务更新:', data);

            // wsManager 已经解包了 message.data，所以 data 直接就是 TaskUpdate 对象
            if (!data.task_id) {
                console.warn('任务更新数据格式不正确:', data);
                return;
            }

            const taskId = data.task_id;

            setTasks(prevTasks => {
                const updatedTasks = [...prevTasks];
                const taskIndex = updatedTasks.findIndex(task => task.id === taskId);

                if (taskIndex !== -1) {
                    // 更新现有任务
                    const updatedTask = {
                        ...updatedTasks[taskIndex],
                        ...data,
                        id: taskId // 确保ID正确
                    };
                    updatedTasks[taskIndex] = updatedTask;
                    console.log('已更新任务:', updatedTask);
                } else {
                    // 新任务，添加到列表
                    const newTask = {
                        id: taskId,
                        ...data
                    };
                    updatedTasks.unshift(newTask);
                    console.log('已添加新任务:', newTask);
                }
                return updatedTasks;
            });

            // 如果任务完成且有图片，更新生成结果
            if (data.status === 'completed' && data.result) {
                try {
                    const result = typeof data.result === 'string'
                        ? JSON.parse(data.result)
                        : data.result;

                    if (result && result.images && Array.isArray(result.images) && result.images.length > 0) {
                        setGeneratedImages(result.images);
                        setStatus('生成完成');
                        message.success(`成功生成 ${result.images.length} 张图片`);
                    }
                } catch (error) {
                    console.error('解析任务结果失败:', error);
                }
            }

            // 如果任务完成且当前处于等待状态，重置状态
            if (data.status === 'completed') {
                setIsGenerating(false);
                setStatus('就绪');
            }
        };

        const handleLogMessage = (data: any) => {
            console.log('收到日志消息:', data);
            // 如果当前有打开的任务详情，且日志属于该任务，则更新日志
            if (selectedTask && data.task_id === selectedTask.id) {
                setTaskLogs(prevLogs => {
                    const newLogs = [...prevLogs, data.message || data.data?.message || '未知日志'];
                    // 自动滚动到底部
                    setTimeout(() => {
                        if (logContainerRef.current) {
                            logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
                        }
                    }, 100);
                    return newLogs;
                });
            }
        };

        // 注册WebSocket事件监听器
        wsManager.on('task_update', handleTaskUpdate);
        wsManager.on('log_message', handleLogMessage);
        wsManager.on('global_log', handleLogMessage);

        return () => {
            if (logStream) {
                logStream.close();
            }
            // 清理WebSocket监听器
            wsManager.off('task_update', handleTaskUpdate);
            wsManager.off('log_message', handleLogMessage);
            wsManager.off('global_log', handleLogMessage);
        };
    }, []); // 空依赖数组，只在组件挂载时执行一次

    // ==================== 渲染函数 ====================
    const renderWebUIStatus = () => {
        const getStatusColor = (status: string) => {
            switch (status) {
                case 'running':
                case 'external':
                    return '#52c41a';
                case 'starting':
                    return '#faad14';
                case 'stopped':
                    return '#ff4d4f';
                default:
                    return '#d9d9d9';
            }
        };

        const getStatusText = (status: string) => {
            switch (status) {
                case 'running':
                    return '运行中 (内部管理)';
                case 'external':
                    return '运行中 (外部启动)';
                case 'starting':
                    return '启动中...';
                case 'stopped':
                    return '已停止';
                default:
                    return '未知状态';
            }
        };

        return (
            <Space>
                <Text strong>WebUI状态:</Text>
                <Text style={{ color: getStatusColor(webUIStatus) }}>
                    {getStatusText(webUIStatus)}
                </Text>
                {webUIStatus === 'stopped' ? (
                    <Button
                        type="primary"
                        icon={<PoweroffOutlined />}
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
    };

    const renderParameterControls = () => (
        <Card title="生成参数" size="small">
            <Row gutter={[16, 16]}>
                <Col span={24}>
                    <Space direction="vertical" style={{ width: '100%' }}>
                        <div>
                            <Text strong>提示词:</Text>
                            <TextArea
                                value={userPrompt}
                                onChange={(e) => {
                                    const newUserPrompt = e.target.value;
                                    setUserPrompt(newUserPrompt);
                                    // 同步更新到 params.prompt，包含角色特征
                                    updatePromptWithCharacters(newUserPrompt);
                                }}
                                placeholder="输入提示词..."
                                maxLength={1000}
                                showCount
                                autoSize={false}
                                style={{
                                    width: '100%',
                                    height: '100px',
                                    resize: 'none',
                                    fontSize: '14px',
                                    lineHeight: '1.5',
                                    overflow: 'hidden'
                                }}
                            />
                            {/* 显示最终提示词（包含角色特征） */}
                            {params.prompt !== userPrompt && (
                                <div style={{ marginTop: 8, padding: 8, backgroundColor: '#f0f0f0', borderRadius: 4, fontSize: '12px' }}>
                                    <Text strong style={{ color: '#666' }}>最终提示词:</Text>
                                    <div style={{
                                        marginTop: 4,
                                        marginBottom: 0,
                                        wordBreak: 'break-word',
                                        maxHeight: '100px',
                                        overflow: 'auto',
                                        fontSize: '12px',
                                        lineHeight: '1.5',
                                        fontFamily: 'monospace'
                                    }}>
                                        {(() => {
                                            // 解析提示词，为不同部分添加颜色
                                            const parts = params.prompt.split('\n');
                                            // 定义多个角色特征的颜色（与用户输入的蓝色区分开）
                                            const characterColors = [
                                                { color: '#52c41a', bgColor: '#f6ffed' }, // 绿色
                                                { color: '#13c2c2', bgColor: '#e6fffb' }, // 青色
                                                { color: '#722ed1', bgColor: '#f9f0ff' }, // 紫色
                                                { color: '#eb2f96', bgColor: '#fff0f6' }, // 粉色
                                                { color: '#fa541c', bgColor: '#fff2e8' }, // 橙红色
                                                { color: '#faad14', bgColor: '#fffbe6' }, // 金色
                                                { color: '#2f54eb', bgColor: '#f0f5ff' }, // 深蓝
                                                { color: '#cf1322', bgColor: '#fff1f0' }, // 红色
                                            ];

                                            let characterIndex = 0;
                                            return parts.map((part, index) => {
                                                let color = '#333';
                                                let bgColor = 'transparent';

                                                // 用户输入的提示词（第一行且不是BREAK）
                                                if (index === 0 && part && part.trim() !== 'BREAK') {
                                                    color = '#0958d9';
                                                    bgColor = '#e6f7ff';
                                                }
                                                // BREAK 分隔符
                                                else if (part.trim() === 'BREAK') {
                                                    color = '#fff';
                                                    bgColor = '#fa8c16';
                                                }
                                                // 角色特征标签（包含权重标记的标签）
                                                else if (part.includes(':0.') || part.includes('(:')) {
                                                    const charColor = characterColors[characterIndex % characterColors.length];
                                                    if (charColor) {
                                                        color = charColor.color;
                                                        bgColor = charColor.bgColor;
                                                    }
                                                    characterIndex++;
                                                }

                                                return (
                                                    <div
                                                        key={index}
                                                        style={{
                                                            color,
                                                            backgroundColor: bgColor,
                                                            padding: bgColor !== 'transparent' ? '2px 4px' : '0',
                                                            marginBottom: '2px',
                                                            borderRadius: bgColor !== 'transparent' ? '2px' : '0'
                                                        }}
                                                    >
                                                        {part}
                                                    </div>
                                                );
                                            });
                                        })()}
                                    </div>
                                </div>
                            )}
                        </div>
                        <div>
                            <Text strong>负面提示词:</Text>
                            <TextArea
                                value={params.negative_prompt}
                                onChange={(e) => setParams(prev => ({ ...prev, negative_prompt: e.target.value }))}
                                placeholder="输入负面提示词..."
                                maxLength={500}
                                showCount
                                autoSize={false}
                                style={{
                                    width: '100%',
                                    height: '80px',
                                    resize: 'none',
                                    fontSize: '14px',
                                    lineHeight: '1.5',
                                    overflow: 'hidden'
                                }}
                            />
                        </div>
                    </Space>
                </Col>

                <Col span={12}>
                    <Text strong>步数: {params.steps}</Text>
                    <Slider
                        min={1}
                        max={150}
                        value={params.steps}
                        onChange={(value) => setParams(prev => ({ ...prev, steps: value }))}
                    />
                </Col>

                <Col span={12}>
                    <Text strong>CFG Scale: {params.cfg_scale}</Text>
                    <Slider
                        min={1}
                        max={30}
                        step={0.5}
                        value={params.cfg_scale}
                        onChange={(value) => setParams(prev => ({ ...prev, cfg_scale: value }))}
                    />
                </Col>

                <Col span={8}>
                    <Text strong>宽度:</Text>
                    <InputNumber
                        value={params.width}
                        onChange={(value) => setParams(prev => ({ ...prev, width: value || 512 }))}
                        min={64}
                        max={2048}
                        step={64}
                        style={{ width: '100%' }}
                    />
                </Col>

                <Col span={8}>
                    <Text strong>高度:</Text>
                    <InputNumber
                        value={params.height}
                        onChange={(value) => setParams(prev => ({ ...prev, height: value || 512 }))}
                        min={64}
                        max={2048}
                        step={64}
                        style={{ width: '100%' }}
                    />
                </Col>

                <Col span={8}>
                    <Text strong>种子:</Text>
                    <InputNumber
                        value={params.seed}
                        onChange={(value) => setParams(prev => ({ ...prev, seed: value || -1 }))}
                        min={-1}
                        max={2147483647}
                        style={{ width: '100%' }}
                    />
                    <Button
                        size="small"
                        onClick={handleRandomSeed}
                        style={{ marginTop: 4 }}
                    >
                        随机
                    </Button>
                </Col>

                <Col span={12}>
                    <Text strong>采样器:</Text>
                    <Select
                        value={params.sampler}
                        onChange={(value) => setParams(prev => ({ ...prev, sampler: value }))}
                        style={{ width: '100%' }}
                    >
                        <Option value="DPM++ 2M Karras">DPM++ 2M Karras</Option>
                        <Option value="DPM++ SDE Karras">DPM++ SDE Karras</Option>
                        <Option value="Euler a">Euler a</Option>
                        <Option value="Euler">Euler</Option>
                        <Option value="LMS">LMS</Option>
                        <Option value="Heun">Heun</Option>
                        <Option value="DPM2">DPM2</Option>
                        <Option value="DPM2 a">DPM2 a</Option>
                        <Option value="DPM++ 2S a">DPM++ 2S a</Option>
                        <Option value="DPM++ 2M">DPM++ 2M</Option>
                        <Option value="DPM++ SDE">DPM++ SDE</Option>
                        <Option value="DPM fast">DPM fast</Option>
                        <Option value="DPM adaptive">DPM adaptive</Option>
                        <Option value="LMS Karras">LMS Karras</Option>
                        <Option value="DPM2 Karras">DPM2 Karras</Option>
                        <Option value="DPM2 a Karras">DPM2 a Karras</Option>
                        <Option value="DPM++ 2S a Karras">DPM++ 2S a Karras</Option>
                    </Select>
                </Col>

                <Col span={12}>
                    <Text strong>批次大小: {params.batch_size}</Text>
                    <Slider
                        min={1}
                        max={8}
                        value={params.batch_size}
                        onChange={(value) => setParams(prev => ({ ...prev, batch_size: value }))}
                    />
                </Col>

                <Col span={12}>
                    <Text strong>批次数量: {params.batch_count}</Text>
                    <Slider
                        min={1}
                        max={10}
                        value={params.batch_count}
                        onChange={(value) => setParams(prev => ({ ...prev, batch_count: value }))}
                    />
                </Col>

                <Col span={12}>
                    <Text strong>循环数量: {params.loop_count}</Text>
                    <Slider
                        min={1}
                        max={20}
                        value={params.loop_count}
                        onChange={(value) => setParams(prev => ({ ...prev, loop_count: value }))}
                    />
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                        发包次数，每次发包间隔2秒
                    </Text>
                </Col>

                <Col span={24}>
                    <Space>
                        <Text strong>高分辨率修复:</Text>
                        <Switch
                            checked={params.enable_hr}
                            onChange={(checked) => setParams(prev => ({ ...prev, enable_hr: checked }))}
                        />
                        {params.enable_hr && (
                            <>
                                <Text strong>放大倍数: {params.hr_scale}</Text>
                                <Slider
                                    min={1}
                                    max={4}
                                    step={0.1}
                                    value={params.hr_scale}
                                    onChange={(value) => setParams(prev => ({ ...prev, hr_scale: value }))}
                                    style={{ width: 100 }}
                                />
                            </>
                        )}
                    </Space>
                </Col>
            </Row>

            {/* LoRA 配置 - 固定高度区域 */}
            <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
                <Col span={24}>
                    <Text strong>LoRA 模型:</Text>
                    <div style={{
                        marginTop: 8,
                        height: '200px',
                        overflow: 'auto',
                        border: '1px solid #f0f0f0',
                        borderRadius: '6px',
                        padding: '8px',
                        backgroundColor: '#fafafa'
                    }}>
                        {loras.length > 0 ? (
                            loras.map((lora) => (
                                <div key={lora.lora_key} style={{
                                    marginBottom: 8,
                                    padding: 8,
                                    border: '1px solid #d9d9d9',
                                    borderRadius: 4,
                                    backgroundColor: selectedLoras.includes(lora.lora_key) ? '#f0f8ff' : '#ffffff'
                                }}>
                                    <Space direction="vertical" size={4} style={{ width: '100%' }}>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                            <Text strong style={{ fontSize: 12 }}>{lora.name}</Text>
                                            <Text style={{ fontSize: 11, color: '#666' }}>权重: {lora.weight}</Text>
                                        </div>
                                        {lora.description && (
                                            <Text style={{ fontSize: 11, color: '#666' }}>{lora.description}</Text>
                                        )}
                                        <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                                            {lora.tags && lora.tags.map(tag => (
                                                <Tag key={tag} color="blue">{tag}</Tag>
                                            ))}
                                        </div>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                            <Switch
                                                size="small"
                                                checked={selectedLoras.includes(lora.lora_key)}
                                                onChange={(checked) => {
                                                    if (checked) {
                                                        setSelectedLoras(prev => [...prev, lora.lora_key]);
                                                    } else {
                                                        setSelectedLoras(prev => prev.filter(key => key !== lora.lora_key));
                                                    }
                                                }}
                                            />
                                            <Text style={{ fontSize: 10, color: '#999' }}>{lora.path}</Text>
                                        </div>
                                    </Space>
                                </div>
                            ))
                        ) : (
                            <div style={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                height: '100%',
                                color: '#999',
                                fontSize: '14px'
                            }}>
                                暂无 LoRA 模型
                            </div>
                        )}
                    </div>
                </Col>
            </Row>
        </Card>
    );

    const renderConfigSelector = () => (
        <Card title="配置管理" size="small">
            <Row gutter={[16, 16]}>
                <Col span={12}>
                    <Text strong>分类:</Text>
                    <Select
                        value={selectedCategory}
                        onChange={(value) => {
                            setSelectedCategory(value);
                            setSelectedConfig('');
                        }}
                        style={{ width: '100%' }}
                        placeholder="选择分类"
                        allowClear
                    >
                        {(Array.isArray(configCategories) ? configCategories : []).map(category => (
                            <Option key={category} value={category}>{category}</Option>
                        ))}
                    </Select>
                </Col>

                <Col span={12}>
                    <Text strong>配置:</Text>
                    <Select
                        value={selectedConfig}
                        onChange={(value) => {
                            setSelectedConfig(value);
                            if (value === 'none') {
                                // 重置为默认配置
                                resetToDefaultConfig();
                            } else if (value) {
                                applyConfig(value);
                            }
                        }}
                        style={{ width: '100%' }}
                        placeholder="选择配置"
                        allowClear
                    >
                        <Option key="none" value="none">无 (重置配置)</Option>
                        {(Array.isArray(configs) ? configs : [])
                            .filter(config => !selectedCategory || config.category === selectedCategory)
                            .map(config => (
                                <Option key={config.id} value={config.id}>{config.name}</Option>
                            ))}
                    </Select>
                </Col>

                <Col span={24}>
                    <Text strong>角色特征:</Text>
                    <div style={{
                        marginTop: 8,
                        maxHeight: '400px',
                        overflow: 'auto',
                        border: '1px solid #f0f0f0',
                        borderRadius: '6px',
                        padding: '8px',
                        backgroundColor: '#fafafa'
                    }}>
                        {characters.length > 0 ? (
                            characters.map((character) => (
                                <div key={character.name} style={{
                                    marginBottom: 8,
                                    padding: 8,
                                    border: '1px solid #d9d9d9',
                                    borderRadius: 4,
                                    backgroundColor: selectedCharacters.includes(character.name) ? '#f0f8ff' : '#ffffff'
                                }}>
                                    <Space direction="vertical" size={4} style={{ width: '100%' }}>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                            <Text strong style={{ fontSize: 12 }}>{character.name}</Text>
                                            <Text style={{ fontSize: 11, color: '#666' }}>{character.tags?.length || 0} 标签</Text>
                                        </div>
                                        {character.description && (
                                            <Text style={{ fontSize: 11, color: '#666' }}>{character.description}</Text>
                                        )}
                                        <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                                            {character.tags && character.tags.slice(0, 10).map((tag: string) => {
                                                const weight = character.tag_weights?.[tag] || 0;
                                                return (
                                                    <Tag key={tag} color="purple">
                                                        {tag} {weight > 0 ? `(${(weight * 100).toFixed(0)}%)` : ''}
                                                    </Tag>
                                                );
                                            })}
                                        </div>
                                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                            <Switch
                                                size="small"
                                                checked={selectedCharacters.includes(character.name)}
                                                onChange={(checked) => {
                                                    // 更新选中列表
                                                    const newSelected = checked
                                                        ? [...selectedCharacters, character.name]
                                                        : selectedCharacters.filter(name => name !== character.name);

                                                    setSelectedCharacters(newSelected);

                                                    // 重新构建提示词（包含所有选中的角色特征）
                                                    const selectedTags: string[] = [];
                                                    newSelected.forEach(characterName => {
                                                        const characterData = characters.find(c => c.name === characterName);
                                                        if (characterData && characterData.tags && characterData.tags.length > 0) {
                                                            const tagsString = characterData.tags.map((tag: string) => {
                                                                const weight = characterData.tag_weights?.[tag] || 0;
                                                                return weight > 0.7 ? `((${tag}:${weight.toFixed(2)}))` : tag;
                                                            }).join(', ');
                                                            selectedTags.push(tagsString);
                                                        }
                                                    });

                                                    // 构建最终提示词
                                                    let finalPrompt = userPrompt;
                                                    if (selectedTags.length > 0) {
                                                        // 每个角色特征之间用BREAK分隔
                                                        const allTagsString = selectedTags.join('\nBREAK\n');
                                                        if (userPrompt) {
                                                            finalPrompt = `${userPrompt}\nBREAK\n${allTagsString}`;
                                                        } else {
                                                            finalPrompt = allTagsString;
                                                        }
                                                    }

                                                    // 更新 params.prompt
                                                    console.log('更新最终提示词:', finalPrompt);
                                                    setParams(prev => ({ ...prev, prompt: finalPrompt }));
                                                }}
                                            />
                                            <Text style={{ fontSize: 10, color: '#999' }}>{character.name}</Text>
                                        </div>
                                    </Space>
                                </div>
                            ))
                        ) : (
                            <div style={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                height: '60px',
                                color: '#999',
                                fontSize: '14px'
                            }}>
                                暂无角色特征配置
                            </div>
                        )}
                    </div>
                </Col>

                <Col span={24}>
                    <Space>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={() => {
                                loadConfigs();
                                loadCharacters();
                            }}
                            size="small"
                        >
                            刷新配置
                        </Button>
                        <Button
                            icon={<SaveOutlined />}
                            size="small"
                        >
                            应用配置
                        </Button>
                    </Space>
                </Col>
            </Row>
        </Card>
    );

    const renderControlButtons = () => (
        <Card title="控制面板" size="small">
            <Space direction="vertical" style={{ width: '100%' }}>
                <Space>
                    <Button
                        type="primary"
                        icon={<PlayCircleOutlined />}
                        onClick={handleGenerate}
                        loading={isGenerating}
                        disabled={webUIStatus !== 'running' && webUIStatus !== 'external'}
                        size="large"
                    >
                        {isGenerating ? (status === '等待中...' ? '等待中...' : '生成中...') : '开始生成'}
                    </Button>

                    <Button
                        danger
                        icon={<StopOutlined />}
                        onClick={handleStop}
                        disabled={!isGenerating}
                        size="large"
                    >
                        停止生成
                    </Button>
                </Space>

                <Space>
                    <Button
                        icon={<CopyOutlined />}
                        onClick={handleCopyPrompt}
                        size="small"
                    >
                        复制提示词
                    </Button>

                    <Button
                        icon={<ClearOutlined />}
                        onClick={handleClearImages}
                        size="small"
                    >
                        清空图片
                    </Button>
                </Space>

                {isGenerating && (
                    <div>
                        <Text strong>生成进度:</Text>
                        <Progress percent={progress} status="active" />
                        <Text type="secondary">{status}</Text>
                    </div>
                )}
            </Space>
        </Card>
    );

    const renderGeneratedImages = () => (
        <Card title="生成结果" size="small" style={{ height: '600px', overflow: 'auto' }}>
            {generatedImages.length > 0 ? (
                <Row gutter={[12, 12]}>
                    {generatedImages.map((imageUrl, index) => (
                        <Col span={24} key={index}>
                            <Card
                                size="small"
                                cover={
                                    <div
                                        style={{
                                            cursor: 'pointer',
                                            position: 'relative'
                                        }}
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            handleImageClick(imageUrl, index);
                                        }}
                                    >
                                        <Image
                                            src={imageUrl}
                                            alt={`Generated ${index + 1}`}
                                            style={{
                                                height: 150,
                                                objectFit: 'contain',
                                                backgroundColor: '#f8f8f8'
                                            }}
                                            placeholder={<div style={{ height: 150, background: '#f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>加载中...</div>}
                                            preview={{
                                                visible: false
                                            }}
                                        />
                                        <div style={{
                                            position: 'absolute',
                                            top: 8,
                                            right: 8,
                                            background: 'rgba(0,0,0,0.6)',
                                            color: 'white',
                                            padding: '2px 6px',
                                            borderRadius: '4px',
                                            fontSize: '12px'
                                        }}>
                                            {index + 1}/{generatedImages.length}
                                        </div>
                                    </div>
                                }
                                actions={[
                                    <Button
                                        key="download"
                                        icon={<DownloadOutlined />}
                                        onClick={() => handleDownloadImage(imageUrl, index)}
                                        size="small"
                                        type="primary"
                                    >
                                        下载
                                    </Button>
                                ]}
                                style={{ marginBottom: 8 }}
                            />
                        </Col>
                    ))}
                </Row>
            ) : (
                <div style={{
                    textAlign: 'center',
                    padding: '30px 20px',
                    color: '#999',
                    height: '150px',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center'
                }}>
                    <div style={{ fontSize: '48px', marginBottom: '16px', opacity: 0.3 }}>🖼️</div>
                    <Text type="secondary" style={{ fontSize: '16px' }}>暂无生成的图片</Text>
                    <Text type="secondary" style={{ fontSize: '12px', marginTop: '8px' }}>点击"生成图片"开始创作</Text>
                </div>
            )}
        </Card>
    );


    // ==================== 主渲染 ====================
    return (
        <div style={{ padding: '24px' }}>
            <Typography.Title level={2}>AI 图像生成器</Typography.Title>

            {/* WebUI 状态 */}
            <Card size="small" style={{ marginBottom: 16 }}>
                {renderWebUIStatus()}
            </Card>

            <Row gutter={[16, 16]}>
                {/* 左侧区域 */}
                <Col span={12}>
                    <Row gutter={[0, 16]}>
                        {/* 左上：生成参数 */}
                        <Col span={24}>
                            {renderParameterControls()}
                        </Col>
                        {/* 左下：控制面板（移动到生成参数下方） */}
                        <Col span={24}>
                            {renderControlButtons()}
                        </Col>
                    </Row>
                </Col>

                {/* 右侧区域 */}
                <Col span={12}>
                    <Row gutter={[0, 16]}>
                        {/* 右上：生成结果 */}
                        <Col span={24}>
                            {renderGeneratedImages()}
                        </Col>
                        {/* 右下：配置管理 */}
                        <Col span={24}>
                            {renderConfigSelector()}
                        </Col>
                    </Row>
                </Col>
            </Row>

            {/* 任务管理 */}
            <Card
                title="📋 AI生成任务管理"
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
                            columns={taskColumns}
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
                open={taskDetailVisible}
                onCancel={handleCloseTaskDetail}
                footer={[
                    <Button key="close" onClick={handleCloseTaskDetail}>
                        关闭
                    </Button>
                ]}
                width={800}
                style={{ top: 20 }}
            >
                {selectedTask && (
                    <div>
                        {/* 任务基本信息 */}
                        <Card title="基本信息" size="small" style={{ marginBottom: 16 }}>
                            <Row gutter={[16, 8]}>
                                <Col span={12}>
                                    <Text strong>任务ID: </Text>
                                    <Text copyable>{selectedTask.id}</Text>
                                </Col>
                                <Col span={12}>
                                    <Text strong>状态: </Text>
                                    <Tag color={
                                        selectedTask.status === 'completed' ? 'success' :
                                            selectedTask.status === 'failed' ? 'error' :
                                                selectedTask.status === 'running' ? 'processing' : 'default'
                                    }>
                                        {selectedTask.status === 'completed' ? '已完成' :
                                            selectedTask.status === 'failed' ? '失败' :
                                                selectedTask.status === 'running' ? '运行中' :
                                                    selectedTask.status === 'pending' ? '等待中' : selectedTask.status}
                                    </Tag>
                                </Col>
                                <Col span={12}>
                                    <Text strong>创建时间: </Text>
                                    <Text>{new Date(selectedTask.created_at).toLocaleString()}</Text>
                                </Col>
                                <Col span={12}>
                                    <Text strong>完成时间: </Text>
                                    <Text>{selectedTask.completed_at ? new Date(selectedTask.completed_at).toLocaleString() : '未完成'}</Text>
                                </Col>
                                <Col span={24}>
                                    <Text strong>配置ID: </Text>
                                    <Text>{selectedTask.config_id || 'N/A'}</Text>
                                </Col>
                                {selectedTask.status === 'failed' && (selectedTask.error || selectedTask.error_message) && (
                                    <Col span={24}>
                                        <Text strong>失败原因: </Text>
                                        <Text style={{ color: '#ff4d4f' }}>{selectedTask.error || selectedTask.error_message}</Text>
                                    </Col>
                                )}
                            </Row>
                        </Card>

                        {/* 进度信息 */}
                        <Card title="进度信息" size="small" style={{ marginBottom: 16 }}>
                            <div style={{ marginBottom: 16 }}>
                                <Text strong>总体进度: </Text>
                                <Progress
                                    percent={selectedTask.progress || 0}
                                    status={
                                        selectedTask.status === 'failed' ? 'exception' :
                                            selectedTask.status === 'completed' ? 'success' : 'active'
                                    }
                                />
                            </div>
                            <Row gutter={[16, 8]}>
                                <Col span={8}>
                                    <Text strong>生成图片: </Text>
                                    <Text style={{ color: '#1890ff' }}>{selectedTask.images_generated || 0}</Text>
                                </Col>
                                <Col span={8}>
                                    <Text strong>成功图片: </Text>
                                    <Text style={{ color: '#52c41a' }}>{selectedTask.images_success || 0}</Text>
                                </Col>
                                <Col span={8}>
                                    <Text strong>失败图片: </Text>
                                    <Text style={{ color: '#ff4d4f' }}>{(selectedTask.images_generated || 0) - (selectedTask.images_success || 0)}</Text>
                                </Col>
                            </Row>
                        </Card>

                        {/* 任务日志 */}
                        <Card
                            title={
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                    <span>执行日志</span>
                                    <Space>
                                        <Button
                                            size="small"
                                            onClick={() => setTaskLogs([])}
                                            disabled={taskLogs.length === 0}
                                        >
                                            清空日志
                                        </Button>
                                        {selectedTask?.status === 'running' && (
                                            <Badge status="processing" text="实时更新中" />
                                        )}
                                    </Space>
                                </div>
                            }
                            size="small"
                        >
                            <div
                                ref={logContainerRef}
                                style={{
                                    height: '300px',
                                    overflow: 'auto',
                                    backgroundColor: '#f5f5f5',
                                    padding: '12px',
                                    borderRadius: '4px',
                                    fontFamily: 'monospace',
                                    fontSize: '12px',
                                    lineHeight: '1.5'
                                }}
                            >
                                {taskDetailLoading ? (
                                    <div style={{ textAlign: 'center', padding: '20px' }}>
                                        <Spin size="small" />
                                        <div style={{ marginTop: '8px' }}>加载日志中...</div>
                                    </div>
                                ) : (
                                    taskLogs.map((log, index) => (
                                        <div key={index} style={{
                                            marginBottom: '4px',
                                            color: log.includes('失败') || log.includes('错误') ? '#ff4d4f' :
                                                log.includes('完成') || log.includes('成功') ? '#52c41a' :
                                                    log.includes('开始') ? '#1890ff' : '#666'
                                        }}>
                                            [{new Date().toLocaleTimeString()}] {log}
                                        </div>
                                    ))
                                )}
                            </div>
                        </Card>
                    </div>
                )}
            </Modal>

            {/* 图片查看器Modal */}
            <Modal
                title="图片查看器"
                open={imageViewerVisible}
                onCancel={handleCloseViewer}
                footer={null}
                width="90%"
                style={{ top: 20 }}
                bodyStyle={{ padding: 0, height: '80vh' }}
            >
                <div style={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    position: 'relative'
                }}>
                    {/* 主图片显示区域 */}
                    <div style={{
                        width: '100%',
                        height: 'calc(100% - 60px)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        position: 'relative',
                        padding: '20px'
                    }}>
                        {viewerImages.length > 0 && (
                            <img
                                src={viewerImages[currentImageIndex]}
                                alt={`Image ${currentImageIndex + 1}`}
                                style={{
                                    maxWidth: '100%',
                                    maxHeight: '100%',
                                    width: 'auto',
                                    height: 'auto',
                                    objectFit: 'contain'
                                }}
                            />
                        )}
                    </div>

                    {/* 导航控制区域 */}
                    <div style={{
                        height: '60px',
                        width: '100%',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        background: '#f5f5f5',
                        borderTop: '1px solid #d9d9d9'
                    }}>
                        <Space size="large">
                            {/* 上一张按钮 */}
                            <Button
                                icon={<ReloadOutlined style={{ transform: 'scaleX(-1)' }} />}
                                onClick={handlePrevImage}
                                disabled={viewerImages.length <= 1}
                                size="large"
                            >
                                上一张
                            </Button>

                            {/* 图片计数器 */}
                            <div style={{
                                fontSize: '16px',
                                fontWeight: 'bold',
                                color: '#1890ff',
                                minWidth: '100px',
                                textAlign: 'center'
                            }}>
                                {currentImageIndex + 1} / {viewerImages.length}
                            </div>

                            {/* 下一张按钮 */}
                            <Button
                                icon={<ReloadOutlined />}
                                onClick={handleNextImage}
                                disabled={viewerImages.length <= 1}
                                size="large"
                            >
                                下一张
                            </Button>
                        </Space>
                    </div>

                    {/* 键盘导航提示 */}
                    <div style={{
                        position: 'absolute',
                        top: 10,
                        right: 10,
                        background: 'rgba(0,0,0,0.6)',
                        color: 'white',
                        padding: '4px 8px',
                        borderRadius: '4px',
                        fontSize: '12px'
                    }}>
                        使用 ← → 键切换图片
                    </div>
                </div>
            </Modal>

        </div>
    );
};

export default AIGeneratorPage;
