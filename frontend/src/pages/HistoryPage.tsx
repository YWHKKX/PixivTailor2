import React, { useState, useEffect } from 'react';
import {
    Row,
    Col,
    Card,
    Table,
    Button,
    Input,
    Select,
    DatePicker,
    Space,
    Typography,
    Tag,
    Modal,
    Statistic,
    Tabs,
    message,
    Tooltip,
    Popconfirm,
    Spin
} from 'antd';
import {
    SearchOutlined,
    DownloadOutlined,
    DeleteOutlined,
    EyeOutlined,
    ReloadOutlined,
    FileImageOutlined,
    DatabaseOutlined,
    RobotOutlined
} from '@ant-design/icons';
import { apiService } from '@/services/api';

const { Title, Text } = Typography;
const { Option } = Select;
const { RangePicker } = DatePicker;
const { Search } = Input;
const { TabPane } = Tabs;

interface HistoryItem {
    id: string;
    type: 'generation' | 'crawl' | 'train' | 'tag' | 'classify';
    title: string;
    description: string;
    status: 'completed' | 'failed' | 'running' | 'pending' | 'cancelled';
    createdAt: string;
    completedAt?: string;
    duration?: number;
    resultCount?: number;
    tags: string[];
    params?: any;
    results?: string[];
    progress?: number;
    errorMessage?: string;
}

const HistoryPage: React.FC = () => {
    const [historyData, setHistoryData] = useState<HistoryItem[]>([]);
    const [filteredData, setFilteredData] = useState<HistoryItem[]>([]);
    const [selectedItem, setSelectedItem] = useState<HistoryItem | null>(null);
    const [isDetailModalVisible, setIsDetailModalVisible] = useState(false);
    const [activeTab, setActiveTab] = useState('all');
    const [loading, setLoading] = useState(false);

    // 从API获取任务历史数据
    useEffect(() => {
        loadHistoryData();
    }, []);

    const loadHistoryData = async () => {
        setLoading(true);
        try {
            const response = await apiService.getTasks(1, 100); // 获取前100个任务
            const tasks = response.tasks;

            // 转换任务数据为历史记录格式
            const historyItems: HistoryItem[] = tasks.map(task => ({
                id: task.id.toString(),
                type: task.type as any,
                title: task.name || `${task.type}任务`,
                description: `任务ID: ${task.id}`,
                status: task.status as any,
                createdAt: new Date(task.created_at).toLocaleString(),
                completedAt: task.completed_at ? new Date(task.completed_at).toLocaleString() : undefined,
                duration: task.completed_at && task.started_at ?
                    Math.floor((new Date(task.completed_at).getTime() - new Date(task.started_at).getTime()) / 1000) : undefined,
                resultCount: 0, // 需要从其他API获取
                tags: [task.type],
                params: task.config ? JSON.parse(task.config) : {},
                progress: task.progress,
                errorMessage: task.error_message
            }));

            setHistoryData(historyItems);
            setFilteredData(historyItems);
        } catch (error) {
            console.error('获取历史数据失败:', error);
            message.error('获取历史数据失败，请检查后端服务连接');
            setHistoryData([]);
            setFilteredData([]);
        } finally {
            setLoading(false);
        }
    };

    const columns = [
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            width: 80,
            render: (type: string) => (
                <Tag color={type === 'generation' ? 'blue' : 'green'}>
                    {type === 'generation' ? '生成' : '爬取'}
                </Tag>
            )
        },
        {
            title: '标题',
            dataIndex: 'title',
            key: 'title',
            render: (text: string, record: HistoryItem) => (
                <div>
                    <Text strong>{text}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                        {record.description}
                    </Text>
                </div>
            )
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status: string) => {
                const statusMap = {
                    completed: { color: 'success', text: '成功' },
                    failed: { color: 'error', text: '失败' },
                    running: { color: 'processing', text: '运行中' },
                    pending: { color: 'default', text: '等待中' },
                    cancelled: { color: 'warning', text: '已取消' }
                };
                const config = statusMap[status as keyof typeof statusMap] || { color: 'default', text: status };
                return <Tag color={config.color}>{config.text}</Tag>;
            }
        },
        {
            title: '结果数量',
            dataIndex: 'resultCount',
            key: 'resultCount',
            render: (count: number) => count ? `${count} 项` : '-'
        },
        {
            title: '创建时间',
            dataIndex: 'createdAt',
            key: 'createdAt',
            sorter: (a: HistoryItem, b: HistoryItem) =>
                new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
        },
        {
            title: '耗时',
            dataIndex: 'duration',
            key: 'duration',
            render: (duration: number) => duration ? `${Math.floor(duration / 60)}分${duration % 60}秒` : '-'
        },
        {
            title: '标签',
            dataIndex: 'tags',
            key: 'tags',
            render: (tags: string[]) => (
                <Space wrap>
                    {tags.slice(0, 2).map(tag => (
                        <Tag key={tag} color="blue">{tag}</Tag>
                    ))}
                    {tags.length > 2 && <Tag>+{tags.length - 2}</Tag>}
                </Space>
            )
        },
        {
            title: '操作',
            key: 'actions',
            render: (_: any, record: HistoryItem) => (
                <Space>
                    <Tooltip title="查看详情">
                        <Button
                            type="text"
                            icon={<EyeOutlined />}
                            onClick={() => handleViewDetail(record)}
                        />
                    </Tooltip>
                    {record.status === 'completed' && (
                        <Tooltip title="下载结果">
                            <Button
                                type="text"
                                icon={<DownloadOutlined />}
                                onClick={() => handleDownloadResult(record)}
                            />
                        </Tooltip>
                    )}
                    <Popconfirm
                        title="确定要删除这条记录吗？"
                        onConfirm={() => handleDelete(record.id)}
                        okText="确定"
                        cancelText="取消"
                    >
                        <Tooltip title="删除">
                            <Button type="text" danger icon={<DeleteOutlined />} />
                        </Tooltip>
                    </Popconfirm>
                </Space>
            )
        }
    ];

    const handleViewDetail = (item: HistoryItem) => {
        setSelectedItem(item);
        setIsDetailModalVisible(true);
    };

    const handleDelete = (id: string) => {
        setHistoryData(prev => prev.filter(item => item.id !== id));
        setFilteredData(prev => prev.filter(item => item.id !== id));
        message.success('记录已删除');
    };

    const handleDownloadResult = (item: HistoryItem) => {
        try {
            // 根据类型创建下载链接
            let downloadUrl = '';
            let filename = '';

            if (item.type === 'generation') {
                // 生成结果下载
                downloadUrl = `#/ai-generator?result=${item.id}`;
                filename = `generation_${item.id}.json`;
            } else if (item.type === 'crawl') {
                // 爬取结果下载
                downloadUrl = `#/crawler?result=${item.id}`;
                filename = `crawl_${item.id}.json`;
            } else {
                // 其他类型
                downloadUrl = `#/history?item=${item.id}`;
                filename = `${item.type}_${item.id}.json`;
            }

            // 创建下载链接
            const link = document.createElement('a');
            link.href = downloadUrl;
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);

            message.success('结果下载已开始');
        } catch (error) {
            console.error('下载结果失败:', error);
            message.error('下载结果失败');
        }
    };

    const handleExportData = () => {
        try {
            if (!historyData || historyData.length === 0) {
                message.warning('没有数据可导出');
                return;
            }

            // 创建CSV数据
            const csvData = historyData.map(item => ({
                id: item.id,
                type: item.type,
                title: item.title,
                description: item.description,
                status: item.status,
                createdAt: item.createdAt,
                completedAt: item.completedAt || '',
                duration: item.duration || '',
                resultCount: item.resultCount || '',
                tags: item.tags?.join(',') || ''
            }));

            // 转换为CSV格式
            const headers = ['ID', '类型', '标题', '描述', '状态', '创建时间', '完成时间', '耗时', '结果数量', '标签'];
            const csvContent = [
                headers.join(','),
                ...csvData.map(row => [
                    row.id,
                    row.type,
                    `"${row.title}"`,
                    `"${row.description}"`,
                    row.status,
                    `"${row.createdAt}"`,
                    `"${row.completedAt}"`,
                    `"${row.duration}"`,
                    `"${row.resultCount}"`,
                    `"${row.tags}"`
                ].join(','))
            ].join('\n');

            // 创建下载链接
            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = `history_data_${new Date().toISOString().split('T')[0]}.csv`;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);

            // 清理URL对象
            URL.revokeObjectURL(link.href);

            message.success(`数据导出成功！共导出 ${historyData.length} 条记录`);
        } catch (error) {
            console.error('导出数据失败:', error);
            message.error('导出数据失败');
        }
    };

    const handleBatchDelete = () => {
        Modal.confirm({
            title: '批量删除确认',
            content: `确定要删除所有 ${historyData?.length || 0} 条记录吗？此操作不可撤销。`,
            okText: '确定删除',
            cancelText: '取消',
            okType: 'danger',
            onOk: () => {
                try {
                    setHistoryData([]);
                    setFilteredData([]);
                    message.success('所有记录已删除');
                } catch (error) {
                    console.error('批量删除失败:', error);
                    message.error('批量删除失败');
                }
            }
        });
    };

    const handleSearch = (value: string) => {
        const filtered = historyData.filter(item =>
            item.title.toLowerCase().includes(value.toLowerCase()) ||
            item.description.toLowerCase().includes(value.toLowerCase()) ||
            item.tags.some(tag => tag.toLowerCase().includes(value.toLowerCase()))
        );
        setFilteredData(filtered);
    };

    const handleFilter = (type: string) => {
        if (type === 'all') {
            setFilteredData(historyData);
        } else {
            const filtered = historyData.filter(item => item.type === type);
            setFilteredData(filtered);
        }
        setActiveTab(type);
    };

    const handleDateFilter = (dates: any) => {
        if (!dates || dates.length === 0) {
            setFilteredData(historyData);
            return;
        }

        const [start, end] = dates;
        const filtered = historyData.filter(item => {
            const itemDate = new Date(item.createdAt);
            return itemDate >= start && itemDate <= end;
        });
        setFilteredData(filtered);
    };

    const generationCount = historyData.filter(item => item.type === 'generation').length;
    const crawlCount = historyData.filter(item => item.type === 'crawl').length;
    const successCount = historyData.filter(item => item.status === 'completed').length;
    const totalResults = historyData.reduce((sum, item) => sum + (item.resultCount || 0), 0);

    return (
        <div>
            {/* 统计卡片 */}
            <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
                <Col xs={24} sm={6}>
                    <Card>
                        <Statistic
                            title="生成任务"
                            value={generationCount}
                            valueStyle={{ color: '#1890ff' }}
                            prefix={<RobotOutlined />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={6}>
                    <Card>
                        <Statistic
                            title="爬取任务"
                            value={crawlCount}
                            valueStyle={{ color: '#52c41a' }}
                            prefix={<DatabaseOutlined />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={6}>
                    <Card>
                        <Statistic
                            title="成功任务"
                            value={successCount}
                            valueStyle={{ color: '#faad14' }}
                            prefix={<FileImageOutlined />}
                        />
                    </Card>
                </Col>
                <Col xs={24} sm={6}>
                    <Card>
                        <Statistic
                            title="总结果数"
                            value={totalResults}
                            valueStyle={{ color: '#722ed1' }}
                            prefix={<DownloadOutlined />}
                        />
                    </Card>
                </Col>
            </Row>

            {/* 筛选和搜索 */}
            <Card style={{ marginBottom: 24 }}>
                <Row gutter={[16, 16]} align="middle">
                    <Col xs={24} sm={8}>
                        <Search
                            placeholder="搜索历史记录..."
                            onSearch={handleSearch}
                            prefix={<SearchOutlined />}
                            allowClear
                        />
                    </Col>
                    <Col xs={24} sm={6}>
                        <RangePicker
                            placeholder={['开始日期', '结束日期']}
                            onChange={handleDateFilter}
                            style={{ width: '100%' }}
                        />
                    </Col>
                    <Col xs={24} sm={6}>
                        <Select
                            placeholder="筛选状态"
                            style={{ width: '100%' }}
                            allowClear
                        >
                            <Option value="completed">成功</Option>
                            <Option value="failed">失败</Option>
                            <Option value="running">运行中</Option>
                        </Select>
                    </Col>
                    <Col xs={24} sm={4}>
                        <Button
                            icon={<ReloadOutlined />}
                            onClick={loadHistoryData}
                            loading={loading}
                        >
                            刷新
                        </Button>
                    </Col>
                </Row>
            </Card>

            {/* 标签页 */}
            <Card>
                <Tabs
                    activeKey={activeTab}
                    onChange={handleFilter}
                    tabBarExtraContent={
                        <Space>
                            <Button
                                icon={<DownloadOutlined />}
                                onClick={handleExportData}
                            >
                                导出数据
                            </Button>
                            <Button
                                icon={<DeleteOutlined />}
                                danger
                                onClick={handleBatchDelete}
                            >
                                批量删除
                            </Button>
                        </Space>
                    }
                >
                    <TabPane tab={`全部 (${historyData.length})`} key="all">
                        <Spin spinning={loading}>
                            <Table
                                columns={columns}
                                dataSource={filteredData}
                                rowKey="id"
                                pagination={{ pageSize: 10 }}
                                scroll={{ x: 1000 }}
                            />
                        </Spin>
                    </TabPane>
                    <TabPane tab={`生成任务 (${generationCount})`} key="generation">
                        <Table
                            columns={columns}
                            dataSource={filteredData.filter(item => item.type === 'generation')}
                            rowKey="id"
                            pagination={{ pageSize: 10 }}
                            scroll={{ x: 1000 }}
                        />
                    </TabPane>
                    <TabPane tab={`爬取任务 (${crawlCount})`} key="crawl">
                        <Table
                            columns={columns}
                            dataSource={filteredData.filter(item => item.type === 'crawl')}
                            rowKey="id"
                            pagination={{ pageSize: 10 }}
                            scroll={{ x: 1000 }}
                        />
                    </TabPane>
                </Tabs>
            </Card>

            {/* 详情模态框 */}
            <Modal
                title="任务详情"
                open={isDetailModalVisible}
                onCancel={() => setIsDetailModalVisible(false)}
                footer={null}
                width={800}
            >
                {selectedItem && (
                    <div>
                        <Row gutter={[16, 16]}>
                            <Col span={12}>
                                <Text strong>任务类型: </Text>
                                <Tag color={selectedItem.type === 'generation' ? 'blue' : 'green'}>
                                    {selectedItem.type === 'generation' ? 'AI生成' : '数据爬取'}
                                </Tag>
                            </Col>
                            <Col span={12}>
                                <Text strong>状态: </Text>
                                <Tag color={selectedItem.status === 'completed' ? 'success' : 'error'}>
                                    {selectedItem.status === 'completed' ? '成功' : '失败'}
                                </Tag>
                            </Col>
                            <Col span={24}>
                                <Text strong>标题: </Text>
                                <Text>{selectedItem.title}</Text>
                            </Col>
                            <Col span={24}>
                                <Text strong>描述: </Text>
                                <Text>{selectedItem.description}</Text>
                            </Col>
                            <Col span={12}>
                                <Text strong>创建时间: </Text>
                                <Text>{selectedItem.createdAt}</Text>
                            </Col>
                            <Col span={12}>
                                <Text strong>完成时间: </Text>
                                <Text>{selectedItem.completedAt || '-'}</Text>
                            </Col>
                            <Col span={24}>
                                <Text strong>标签: </Text>
                                <Space wrap>
                                    {selectedItem.tags.map(tag => (
                                        <Tag key={tag} color="blue">{tag}</Tag>
                                    ))}
                                </Space>
                            </Col>
                        </Row>

                        {selectedItem.params && (
                            <div style={{ marginTop: 16 }}>
                                <Title level={5}>参数配置</Title>
                                <pre style={{
                                    background: '#f5f5f5',
                                    padding: 12,
                                    borderRadius: 4,
                                    fontSize: 12,
                                    overflow: 'auto'
                                }}>
                                    {JSON.stringify(selectedItem.params, null, 2)}
                                </pre>
                            </div>
                        )}

                        {selectedItem.results && selectedItem.results.length > 0 && (
                            <div style={{ marginTop: 16 }}>
                                <Title level={5}>生成结果</Title>
                                <Row gutter={[8, 8]}>
                                    {selectedItem.results.map((_result, index) => (
                                        <Col span={6} key={index}>
                                            <div style={{
                                                aspectRatio: '1',
                                                background: '#f0f0f0',
                                                borderRadius: 4,
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                                fontSize: 24
                                            }}>
                                                🖼️
                                            </div>
                                        </Col>
                                    ))}
                                </Row>
                            </div>
                        )}
                    </div>
                )}
            </Modal>
        </div>
    );
};

export default HistoryPage;
