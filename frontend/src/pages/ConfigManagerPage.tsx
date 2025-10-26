import React, { useState, useEffect } from 'react';
import {
    Row,
    Col,
    Card,
    Button,
    Table,
    Input,
    InputNumber,
    Select,
    Space,
    Typography,
    message,
    Modal,
    Form,
    Tag,
    Upload,
    Drawer,
    Divider,
    Tooltip
} from 'antd';
import {
    PlusOutlined,
    EditOutlined,
    EyeOutlined,
    SearchOutlined,
    ImportOutlined
} from '@ant-design/icons';
import { apiService } from '../services/api';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;
const { TextArea } = Input;

interface GenerationConfig {
    id: string;
    name: string;
    description: string;
    category: string;
    prompt: string;
    negative_prompt: string;
    steps: number;
    cfg_scale: number;
    width: number;
    height: number;
    seed: number;
    model: string;
    sampler: string;
    batch_size: number;
    enable_hr: boolean;
    is_default: boolean;
    created_at: string;
    updated_at: string;
}

const ConfigManagerPage: React.FC = () => {
    const [configs, setConfigs] = useState<GenerationConfig[]>([]);
    const [loading, setLoading] = useState(false);
    const [selectedConfig, setSelectedConfig] = useState<GenerationConfig | null>(null);
    const [modalVisible, setModalVisible] = useState(false);
    const [drawerVisible, setDrawerVisible] = useState(false);
    const [editingConfig, setEditingConfig] = useState<GenerationConfig | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedCategory, setSelectedCategory] = useState('');
    const [categories, setCategories] = useState<string[]>([]);
    const [pagination, setPagination] = useState({
        current: 1,
        pageSize: 10,
        total: 0
    });

    const [form] = Form.useForm();

    // 加载配置列表
    const loadConfigs = async (page = 1, pageSize = 10, query = '', category = '') => {
        setLoading(true);
        try {
            console.log('加载配置参数:', { page, pageSize, category, query });
            const response = await apiService.listConfigs({
                page,
                pageSize,
                category,
                search: query
            });
            console.log('配置响应:', response);
            setConfigs(response.configs || []);
            setPagination({
                current: page,
                pageSize: pageSize,
                total: response.total || 0
            });
        } catch (error) {
            console.error('加载配置失败:', error);
            message.error('加载配置失败');
        } finally {
            setLoading(false);
        }
    };

    // 加载分类
    const loadCategories = async () => {
        try {
            const categories = await apiService.getCategories();
            setCategories(categories);
        } catch (error) {
            console.error('加载分类失败:', error);
        }
    };

    useEffect(() => {
        loadConfigs();
        loadCategories();
    }, []);

    // 搜索
    const handleSearch = () => {
        loadConfigs(1, pagination.pageSize, searchQuery, selectedCategory);
    };

    // 重置搜索
    const handleReset = () => {
        setSearchQuery('');
        setSelectedCategory('');
        loadConfigs();
    };

    // 创建配置
    const handleCreate = () => {
        setEditingConfig(null);
        form.resetFields();
        setModalVisible(true);
    };

    // 编辑配置
    const handleEdit = (config: GenerationConfig) => {
        setEditingConfig(config);
        form.setFieldsValue({
            name: config.name,
            description: config.description,
            category: config.category,
            prompt: config.prompt,
            negative_prompt: config.negative_prompt,
            steps: config.steps,
            cfg_scale: config.cfg_scale,
            width: config.width,
            height: config.height,
            seed: config.seed,
            model: config.model,
            sampler: config.sampler,
            batch_size: config.batch_size,
            enable_hr: config.enable_hr,
            is_default: config.is_default
        });
        setModalVisible(true);
    };

    // 查看配置
    const handleView = (config: GenerationConfig) => {
        setSelectedConfig(config);
        setDrawerVisible(true);
    };


    // 应用配置
    const handleSave = async (values: any) => {
        try {
            const configPayload = {
                name: values.name,
                description: values.description,
                category: values.category,
                prompt: values.prompt,
                negative_prompt: values.negative_prompt,
                steps: values.steps,
                cfg_scale: values.cfg_scale,
                width: values.width,
                height: values.height,
                seed: values.seed,
                model: values.model,
                sampler: values.sampler,
                batch_size: values.batch_size,
                enable_hr: values.enable_hr,
                is_default: values.is_default
            };

            if (editingConfig) {
                // 使用文件系统API更新配置
                await apiService.updateConfigFile(editingConfig.id, configPayload);
                message.success('配置文件更新成功');
            } else {
                // 使用文件系统API创建配置
                await apiService.createConfigFile(configPayload);
                message.success('配置文件创建成功');
            }

            setModalVisible(false);
            loadConfigs(pagination.current, pagination.pageSize, searchQuery, selectedCategory);
        } catch (error) {
            message.error('保存失败');
        }
    };

    // 导入配置
    const handleImport = (file: any) => {
        const reader = new FileReader();
        reader.onload = async (e) => {
            try {
                const content = e.target?.result as string;
                const configData = JSON.parse(content);

                await apiService.importConfig(configData);
                message.success('导入成功');
                loadConfigs();
            } catch (error) {
                message.error('导入失败');
            }
        };
        reader.readAsText(file);
        return false; // 阻止默认上传行为
    };


    // 表格列定义
    const columns = [
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
            width: 150,
            render: (text: string, record: GenerationConfig) => (
                <Space direction="vertical" size={0}>
                    <Paragraph
                        ellipsis={{ rows: 2, expandable: false, symbol: '...' }}
                        style={{ margin: 0, fontSize: 14 }}
                    >
                        <Text strong>{text}</Text>
                    </Paragraph>
                    {record.is_default && <Tag color="gold" style={{ fontSize: 10 }}>默认</Tag>}
                </Space>
            )
        },
        {
            title: '描述',
            dataIndex: 'description',
            key: 'description',
            width: 200,
            render: (text: string) => (
                <Paragraph
                    ellipsis={{ rows: 3, expandable: false, symbol: '...' }}
                    style={{ margin: 0, fontSize: 12 }}
                >
                    {text || '无描述'}
                </Paragraph>
            )
        },
        {
            title: '分类',
            dataIndex: 'category',
            key: 'category',
            width: 100,
            render: (category: string) => (
                <Paragraph
                    ellipsis={{ rows: 1, expandable: false, symbol: '...' }}
                    style={{ margin: 0 }}
                >
                    <Tag>{category}</Tag>
                </Paragraph>
            )
        },
        {
            title: '提示词',
            dataIndex: 'prompt',
            key: 'prompt',
            width: 250,
            render: (text: string) => (
                <Tooltip title={text} placement="topLeft">
                    <Paragraph
                        ellipsis={{ rows: 3, expandable: false, symbol: '...' }}
                        style={{ margin: 0, fontSize: 11 }}
                    >
                        <Text code>{text || '无提示词'}</Text>
                    </Paragraph>
                </Tooltip>
            )
        },
        {
            title: '参数',
            key: 'params',
            width: 150,
            render: (_: any, record: GenerationConfig) => (
                <Paragraph
                    ellipsis={{ rows: 3, expandable: false, symbol: '...' }}
                    style={{ margin: 0 }}
                >
                    <Space direction="vertical" size={0}>
                        <Text style={{ fontSize: 11 }}>
                            {record.width}×{record.height}
                        </Text>
                        <Text style={{ fontSize: 11 }}>
                            {record.steps}步, CFG:{record.cfg_scale}
                        </Text>
                        <Text style={{ fontSize: 11 }}>
                            {record.model}
                        </Text>
                    </Space>
                </Paragraph>
            )
        },
        {
            title: '操作',
            key: 'actions',
            render: (_: any, record: GenerationConfig) => (
                <Space>
                    <Tooltip title="查看">
                        <Button icon={<EyeOutlined />} onClick={() => handleView(record)} />
                    </Tooltip>
                    <Tooltip title="编辑">
                        <Button icon={<EditOutlined />} onClick={() => handleEdit(record)} />
                    </Tooltip>
                </Space>
            )
        }
    ];

    return (
        <div>
            <Card>
                <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
                    <Col>
                        <Title level={3}>配置文件管理</Title>
                    </Col>
                    <Col>
                        <Space>
                            <Upload
                                accept=".json"
                                beforeUpload={handleImport}
                                showUploadList={false}
                            >
                                <Button icon={<ImportOutlined />}>导入配置</Button>
                            </Upload>
                            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                                创建配置
                            </Button>
                        </Space>
                    </Col>
                </Row>

                <Row gutter={16} style={{ marginBottom: 16 }}>
                    <Col span={8}>
                        <Input
                            placeholder="搜索配置名称或描述"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            prefix={<SearchOutlined />}
                            onPressEnter={handleSearch}
                        />
                    </Col>
                    <Col span={6}>
                        <Select
                            placeholder="选择分类"
                            value={selectedCategory}
                            onChange={setSelectedCategory}
                            style={{ width: '100%' }}
                            allowClear
                        >
                            {categories.map(cat => (
                                <Option key={cat} value={cat}>{cat}</Option>
                            ))}
                        </Select>
                    </Col>
                    <Col span={6}>
                        <Space>
                            <Button type="primary" onClick={handleSearch}>搜索</Button>
                            <Button onClick={handleReset}>重置</Button>
                        </Space>
                    </Col>
                </Row>

                <Table
                    columns={columns}
                    dataSource={configs}
                    loading={loading}
                    rowKey="id"
                    pagination={{
                        ...pagination,
                        onChange: (page, pageSize) => {
                            loadConfigs(page, pageSize || 10, searchQuery, selectedCategory);
                        }
                    }}
                />
            </Card>

            {/* 创建/编辑配置模态框 */}
            <Modal
                title={editingConfig ? '编辑配置' : '创建配置'}
                open={modalVisible}
                onCancel={() => setModalVisible(false)}
                onOk={() => form.submit()}
                okText="保存配置"
                cancelText="取消"
                width={800}
            >
                <Form
                    form={form}
                    layout="vertical"
                    onFinish={handleSave}
                >
                    <Row gutter={16}>
                        <Col span={12}>
                            <Form.Item
                                name="name"
                                label="配置名称"
                                rules={[{ required: true, message: '请输入配置名称' }]}
                            >
                                <Input />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item
                                name="category"
                                label="分类"
                            >
                                <Select placeholder="选择分类">
                                    {categories.map(cat => (
                                        <Option key={cat} value={cat}>{cat}</Option>
                                    ))}
                                </Select>
                            </Form.Item>
                        </Col>
                    </Row>

                    <Form.Item
                        name="description"
                        label="描述"
                    >
                        <TextArea rows={2} />
                    </Form.Item>

                    <Form.Item
                        name="prompt"
                        label="正面提示词"
                        rules={[{ required: true, message: '请输入正面提示词' }]}
                    >
                        <TextArea rows={4} placeholder="请输入正面提示词" />
                    </Form.Item>

                    <Form.Item
                        name="negative_prompt"
                        label="负面提示词"
                    >
                        <TextArea rows={2} placeholder="请输入负面提示词" />
                    </Form.Item>

                    <Row gutter={16}>
                        <Col span={8}>
                            <Form.Item
                                name="steps"
                                label="采样步数"
                                rules={[{ required: true, message: '请输入采样步数' }]}
                            >
                                <InputNumber min={1} max={100} style={{ width: '100%' }} />
                            </Form.Item>
                        </Col>
                        <Col span={8}>
                            <Form.Item
                                name="cfg_scale"
                                label="CFG Scale"
                                rules={[{ required: true, message: '请输入CFG Scale' }]}
                            >
                                <InputNumber min={1} max={30} step={0.5} style={{ width: '100%' }} />
                            </Form.Item>
                        </Col>
                        <Col span={8}>
                            <Form.Item
                                name="batch_size"
                                label="批次大小"
                                rules={[{ required: true, message: '请输入批次大小' }]}
                            >
                                <InputNumber min={1} max={8} style={{ width: '100%' }} />
                            </Form.Item>
                        </Col>
                    </Row>

                    <Row gutter={16}>
                        <Col span={8}>
                            <Form.Item
                                name="width"
                                label="宽度"
                                rules={[{ required: true, message: '请输入宽度' }]}
                            >
                                <InputNumber min={64} max={2048} step={64} style={{ width: '100%' }} />
                            </Form.Item>
                        </Col>
                        <Col span={8}>
                            <Form.Item
                                name="height"
                                label="高度"
                                rules={[{ required: true, message: '请输入高度' }]}
                            >
                                <InputNumber min={64} max={2048} step={64} style={{ width: '100%' }} />
                            </Form.Item>
                        </Col>
                        <Col span={8}>
                            <Form.Item
                                name="seed"
                                label="随机种子"
                            >
                                <InputNumber min={-1} style={{ width: '100%' }} />
                            </Form.Item>
                        </Col>
                    </Row>

                    <Row gutter={16}>
                        <Col span={12}>
                            <Form.Item
                                name="model"
                                label="模型"
                                rules={[{ required: true, message: '请输入模型名称' }]}
                            >
                                <Input placeholder="请输入模型名称" />
                            </Form.Item>
                        </Col>
                        <Col span={12}>
                            <Form.Item
                                name="sampler"
                                label="采样器"
                                rules={[{ required: true, message: '请输入采样器' }]}
                            >
                                <Select placeholder="请选择采样器">
                                    <Option value="Euler">Euler</Option>
                                    <Option value="Euler a">Euler a</Option>
                                    <Option value="DPM++ 2M">DPM++ 2M</Option>
                                    <Option value="DDIM">DDIM</Option>
                                    <Option value="LMS">LMS</Option>
                                </Select>
                            </Form.Item>
                        </Col>
                    </Row>

                    <Form.Item
                        name="enable_hr"
                        label="高分辨率修复"
                        valuePropName="checked"
                    >
                        <input type="checkbox" />
                    </Form.Item>

                    <Form.Item
                        name="is_default"
                        label="设为默认配置"
                        valuePropName="checked"
                    >
                        <input type="checkbox" />
                    </Form.Item>
                </Form>
            </Modal>

            {/* 查看配置抽屉 */}
            <Drawer
                title="配置详情"
                placement="right"
                width={600}
                open={drawerVisible}
                onClose={() => setDrawerVisible(false)}
            >
                {selectedConfig && (
                    <div>
                        <Title level={4}>{selectedConfig.name}</Title>
                        <Text type="secondary">{selectedConfig.description}</Text>

                        <Divider />

                        <Space direction="vertical" style={{ width: '100%' }}>
                            <div>
                                <Text strong>分类: </Text>
                                <Tag>{selectedConfig.category}</Tag>
                            </div>

                            <div>
                                <Text strong>提示词: </Text>
                                <Text code>{selectedConfig.prompt}</Text>
                            </div>

                            <div>
                                <Text strong>负面提示词: </Text>
                                <Text code>{selectedConfig.negative_prompt}</Text>
                            </div>

                            <div>
                                <Text strong>参数: </Text>
                                <Space direction="vertical" size={0}>
                                    <Text>尺寸: {selectedConfig.width} × {selectedConfig.height}</Text>
                                    <Text>步数: {selectedConfig.steps}, CFG: {selectedConfig.cfg_scale}</Text>
                                    <Text>模型: {selectedConfig.model}</Text>
                                    <Text>采样器: {selectedConfig.sampler}</Text>
                                    <Text>批次大小: {selectedConfig.batch_size}</Text>
                                    <Text>高分辨率修复: {selectedConfig.enable_hr ? '是' : '否'}</Text>
                                    <Text>随机种子: {selectedConfig.seed === -1 ? '随机' : selectedConfig.seed}</Text>
                                </Space>
                            </div>

                            <div>
                                <Text strong>创建时间: </Text>
                                <Text>{new Date(selectedConfig.created_at).toLocaleString()}</Text>
                            </div>

                            <div>
                                <Text strong>更新时间: </Text>
                                <Text>{new Date(selectedConfig.updated_at).toLocaleString()}</Text>
                            </div>
                        </Space>
                    </div>
                )}
            </Drawer>
        </div>
    );
};

export default ConfigManagerPage;
