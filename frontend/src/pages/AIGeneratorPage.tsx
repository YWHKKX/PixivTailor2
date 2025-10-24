import React, { useState } from 'react';
import {
    Row,
    Col,
    Card,
    Input,
    Button,
    Slider,
    Select,
    Switch,
    Progress,
    Space,
    Typography,
    message,
    Divider,
    Tooltip
} from 'antd';
import {
    PlayCircleOutlined,
    StopOutlined,
    DownloadOutlined,
    ClearOutlined,
    HistoryOutlined,
    SaveOutlined,
    CopyOutlined
} from '@ant-design/icons';

const { TextArea } = Input;
const { Text } = Typography;
const { Option } = Select;

interface GenerationParams {
    prompt: string;
    negativePrompt: string;
    steps: number;
    cfgScale: number;
    width: number;
    height: number;
    seed: number;
    model: string;
    sampler: string;
    batchSize: number;
    enableHR: boolean;
}

const AIGeneratorPage: React.FC = () => {
    const [params, setParams] = useState<GenerationParams>({
        prompt: '',
        negativePrompt: '',
        steps: 20,
        cfgScale: 7.0,
        width: 512,
        height: 512,
        seed: -1,
        model: 'stable-diffusion-v1.5',
        sampler: 'Euler',
        batchSize: 1,
        enableHR: false
    });

    const [isGenerating, setIsGenerating] = useState(false);
    const [progress, setProgress] = useState(0);
    const [generatedImages, setGeneratedImages] = useState<string[]>([]);
    const [status, setStatus] = useState('就绪');

    const models = [
        { value: 'stable-diffusion-v1.5', label: 'Stable Diffusion v1.5' },
        { value: 'stable-diffusion-v2.1', label: 'Stable Diffusion v2.1' },
        { value: 'dreamshaper', label: 'DreamShaper' },
        { value: 'realistic-vision', label: 'Realistic Vision' }
    ];

    const samplers = [
        { value: 'Euler', label: 'Euler' },
        { value: 'Euler a', label: 'Euler a' },
        { value: 'DPM++ 2M', label: 'DPM++ 2M' },
        { value: 'DDIM', label: 'DDIM' },
        { value: 'LMS', label: 'LMS' }
    ];

    const handleGenerate = async () => {
        if (!params.prompt.trim()) {
            message.error('请输入提示词');
            return;
        }

        setIsGenerating(true);
        setProgress(0);
        setStatus('正在生成...');

        // 模拟生成过程
        const interval = setInterval(() => {
            setProgress(prev => {
                if (prev >= 100) {
                    clearInterval(interval);
                    setIsGenerating(false);
                    setStatus('生成完成');
                    setGeneratedImages(prev => [...prev, `generated_${Date.now()}.jpg`]);
                    message.success('图像生成完成！');
                    return 100;
                }
                return prev + 10;
            });
        }, 200);
    };

    const handleStop = () => {
        setIsGenerating(false);
        setProgress(0);
        setStatus('已停止');
        message.info('生成已停止');
    };

    const handleClear = () => {
        setGeneratedImages([]);
        setProgress(0);
        setStatus('已清空');
        message.info('已清空所有结果');
    };

    const handleDownload = (imageUrl: string) => {
        // 模拟下载
        message.success(`正在下载 ${imageUrl}`);
    };

    const handleDownloadAll = () => {
        if (generatedImages.length === 0) {
            message.warning('没有可下载的图像');
            return;
        }
        message.success(`正在下载 ${generatedImages.length} 张图像`);
    };

    const handleSaveConfig = () => {
        // 保存配置到本地存储
        localStorage.setItem('ai-generator-config', JSON.stringify(params));
        message.success('配置已保存');
    };

    const handleLoadConfig = () => {
        const saved = localStorage.getItem('ai-generator-config');
        if (saved) {
            setParams(JSON.parse(saved));
            message.success('配置已加载');
        } else {
            message.warning('没有找到保存的配置');
        }
    };

    const handleRandomSeed = () => {
        setParams(prev => ({ ...prev, seed: Math.floor(Math.random() * 1000000) }));
    };

    return (
        <div>
            <Row gutter={[24, 24]}>
                {/* 左侧控制面板 */}
                <Col xs={24} lg={8}>
                    <Space direction="vertical" size="large" style={{ width: '100%' }}>
                        {/* 提示词输入 */}
                        <Card title="🎨 提示词设置" size="small">
                            <Space direction="vertical" style={{ width: '100%' }} size="middle">
                                <div>
                                    <Text strong>正面提示词</Text>
                                    <TextArea
                                        value={params.prompt}
                                        onChange={(e) => setParams(prev => ({ ...prev, prompt: e.target.value }))}
                                        placeholder="描述你想要生成的图像..."
                                        rows={4}
                                        maxLength={1000}
                                        showCount
                                        style={{ marginTop: 8 }}
                                    />
                                </div>
                                <div>
                                    <Text strong>负面提示词</Text>
                                    <TextArea
                                        value={params.negativePrompt}
                                        onChange={(e) => setParams(prev => ({ ...prev, negativePrompt: e.target.value }))}
                                        placeholder="描述你不想要的内容..."
                                        rows={2}
                                        maxLength={500}
                                        showCount
                                        style={{ marginTop: 8 }}
                                    />
                                </div>
                            </Space>
                        </Card>

                        {/* 生成参数 */}
                        <Card title="⚙️ 生成参数" size="small">
                            <Space direction="vertical" style={{ width: '100%' }} size="middle">
                                <div>
                                    <Text strong>采样步数: {params.steps}</Text>
                                    <Slider
                                        min={1}
                                        max={100}
                                        value={params.steps}
                                        onChange={(value) => setParams(prev => ({ ...prev, steps: value }))}
                                        marks={{ 1: '1', 20: '20', 50: '50', 100: '100' }}
                                    />
                                </div>
                                <div>
                                    <Text strong>CFG Scale: {params.cfgScale}</Text>
                                    <Slider
                                        min={1}
                                        max={30}
                                        step={0.5}
                                        value={params.cfgScale}
                                        onChange={(value) => setParams(prev => ({ ...prev, cfgScale: value }))}
                                        marks={{ 1: '1', 7: '7', 15: '15', 30: '30' }}
                                    />
                                </div>
                                <Row gutter={16}>
                                    <Col span={12}>
                                        <Text strong>宽度: {params.width}</Text>
                                        <Slider
                                            min={64}
                                            max={2048}
                                            step={64}
                                            value={params.width}
                                            onChange={(value) => setParams(prev => ({ ...prev, width: value }))}
                                            marks={{ 64: '64', 512: '512', 1024: '1024', 2048: '2048' }}
                                        />
                                    </Col>
                                    <Col span={12}>
                                        <Text strong>高度: {params.height}</Text>
                                        <Slider
                                            min={64}
                                            max={2048}
                                            step={64}
                                            value={params.height}
                                            onChange={(value) => setParams(prev => ({ ...prev, height: value }))}
                                            marks={{ 64: '64', 512: '512', 1024: '1024', 2048: '2048' }}
                                        />
                                    </Col>
                                </Row>
                                <div>
                                    <Text strong>随机种子: {params.seed === -1 ? '随机' : params.seed}</Text>
                                    <Space style={{ marginTop: 8 }}>
                                        <Input
                                            type="number"
                                            value={params.seed === -1 ? '' : params.seed}
                                            onChange={(e) => setParams(prev => ({
                                                ...prev,
                                                seed: e.target.value ? parseInt(e.target.value) : -1
                                            }))}
                                            placeholder="-1 (随机)"
                                            style={{ width: 120 }}
                                        />
                                        <Button onClick={handleRandomSeed}>随机</Button>
                                    </Space>
                                </div>
                            </Space>
                        </Card>

                        {/* 高级设置 */}
                        <Card title="🔧 高级设置" size="small">
                            <Space direction="vertical" style={{ width: '100%' }} size="middle">
                                <div>
                                    <Text strong>模型选择</Text>
                                    <Select
                                        value={params.model}
                                        onChange={(value) => setParams(prev => ({ ...prev, model: value }))}
                                        style={{ width: '100%', marginTop: 8 }}
                                    >
                                        {models.map(model => (
                                            <Option key={model.value} value={model.value}>
                                                {model.label}
                                            </Option>
                                        ))}
                                    </Select>
                                </div>
                                <div>
                                    <Text strong>采样器</Text>
                                    <Select
                                        value={params.sampler}
                                        onChange={(value) => setParams(prev => ({ ...prev, sampler: value }))}
                                        style={{ width: '100%', marginTop: 8 }}
                                    >
                                        {samplers.map(sampler => (
                                            <Option key={sampler.value} value={sampler.value}>
                                                {sampler.label}
                                            </Option>
                                        ))}
                                    </Select>
                                </div>
                                <div>
                                    <Text strong>批次大小: {params.batchSize}</Text>
                                    <Slider
                                        min={1}
                                        max={8}
                                        value={params.batchSize}
                                        onChange={(value) => setParams(prev => ({ ...prev, batchSize: value }))}
                                        marks={{ 1: '1', 4: '4', 8: '8' }}
                                    />
                                </div>
                                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                                    <Text strong>高分辨率修复</Text>
                                    <Switch
                                        checked={params.enableHR}
                                        onChange={(checked) => setParams(prev => ({ ...prev, enableHR: checked }))}
                                    />
                                </div>
                            </Space>
                        </Card>

                        {/* 控制按钮 */}
                        <Card size="small">
                            <Space direction="vertical" style={{ width: '100%' }} size="middle">
                                <Row gutter={8}>
                                    <Col span={12}>
                                        <Button
                                            type="primary"
                                            size="large"
                                            icon={<PlayCircleOutlined />}
                                            onClick={handleGenerate}
                                            loading={isGenerating}
                                            block
                                            style={{ height: 48 }}
                                        >
                                            {isGenerating ? '生成中...' : '开始生成'}
                                        </Button>
                                    </Col>
                                    <Col span={12}>
                                        <Button
                                            danger
                                            size="large"
                                            icon={<StopOutlined />}
                                            onClick={handleStop}
                                            disabled={!isGenerating}
                                            block
                                            style={{ height: 48 }}
                                        >
                                            停止
                                        </Button>
                                    </Col>
                                </Row>

                                <div>
                                    <Text strong>生成进度</Text>
                                    <Progress
                                        percent={progress}
                                        status={isGenerating ? 'active' : 'normal'}
                                        strokeColor={{
                                            '0%': '#108ee9',
                                            '100%': '#87d068',
                                        }}
                                    />
                                </div>

                                <div>
                                    <Text strong>状态: </Text>
                                    <Text type={isGenerating ? 'warning' : 'success'}>{status}</Text>
                                </div>

                                <Divider />

                                <Space wrap>
                                    <Button icon={<SaveOutlined />} onClick={handleSaveConfig}>
                                        保存配置
                                    </Button>
                                    <Button icon={<HistoryOutlined />} onClick={handleLoadConfig}>
                                        加载配置
                                    </Button>
                                    <Button icon={<ClearOutlined />} onClick={handleClear}>
                                        清空结果
                                    </Button>
                                </Space>
                            </Space>
                        </Card>
                    </Space>
                </Col>

                {/* 右侧输出区域 */}
                <Col xs={24} lg={16}>
                    <Card
                        title="🖼️ 生成结果"
                        extra={
                            <Space>
                                <Button
                                    icon={<DownloadOutlined />}
                                    onClick={handleDownloadAll}
                                    disabled={generatedImages.length === 0}
                                >
                                    下载全部
                                </Button>
                                <Button
                                    icon={<ClearOutlined />}
                                    onClick={handleClear}
                                    disabled={generatedImages.length === 0}
                                >
                                    清空
                                </Button>
                            </Space>
                        }
                    >
                        {generatedImages.length === 0 ? (
                            <div style={{
                                textAlign: 'center',
                                padding: '60px 20px',
                                background: '#fafafa',
                                borderRadius: 8,
                                border: '2px dashed #d9d9d9'
                            }}>
                                <div style={{ fontSize: 48, marginBottom: 16 }}>🎨</div>
                                <Text type="secondary" style={{ fontSize: 16 }}>
                                    生成结果将显示在这里
                                </Text>
                            </div>
                        ) : (
                            <Row gutter={[16, 16]}>
                                {generatedImages.map((image, index) => (
                                    <Col xs={24} sm={12} md={8} key={index}>
                                        <Card
                                            hoverable
                                            cover={
                                                <div style={{
                                                    height: 200,
                                                    background: '#f0f0f0',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'center',
                                                    fontSize: 48
                                                }}>
                                                    🖼️
                                                </div>
                                            }
                                            actions={[
                                                <Tooltip title="下载">
                                                    <DownloadOutlined onClick={() => handleDownload(image)} />
                                                </Tooltip>,
                                                <Tooltip title="复制">
                                                    <CopyOutlined />
                                                </Tooltip>
                                            ]}
                                        >
                                            <Card.Meta
                                                title={`生成图像 ${index + 1}`}
                                                description={`${params.width} × ${params.height}`}
                                            />
                                        </Card>
                                    </Col>
                                ))}
                            </Row>
                        )}
                    </Card>
                </Col>
            </Row>
        </div>
    );
};

export default AIGeneratorPage;
