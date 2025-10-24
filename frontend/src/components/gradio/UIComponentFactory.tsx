import React from 'react';
import { Button, Slider, Input, Select, Switch, Progress, Card, Row, Col, Space } from 'antd';
import { PlayCircleOutlined, StopOutlined, DownloadOutlined, ClearOutlined } from '@ant-design/icons';

export interface UIComponentFactory {
    createInputPanel(): React.ReactElement;
    createOutputPanel(): React.ReactElement;
    createControlPanel(): React.ReactElement;
    createSettingsPanel(): React.ReactElement;
}

export class GradioUIComponentFactory implements UIComponentFactory {

    createInputPanel(): React.ReactElement {
        return (
            <Card title="🎨 输入面板" className="gradio-input-panel">
                <Space direction="vertical" style={{ width: '100%' }} size="large">
                    <div>
                        <label className="gradio-label">提示词</label>
                        <Input.TextArea
                            id="prompt_input"
                            placeholder="输入你的提示词..."
                            rows={4}
                            maxLength={1000}
                            showCount
                        />
                    </div>

                    <div>
                        <label className="gradio-label">负面提示词</label>
                        <Input.TextArea
                            id="negative_prompt_input"
                            placeholder="输入负面提示词..."
                            rows={2}
                            maxLength={500}
                            showCount
                        />
                    </div>

                    <Row gutter={16}>
                        <Col span={12}>
                            <div>
                                <label className="gradio-label">步数</label>
                                <Slider
                                    id="steps_slider"
                                    min={1}
                                    max={100}
                                    defaultValue={20}
                                    marks={{ 1: '1', 20: '20', 50: '50', 100: '100' }}
                                />
                            </div>
                        </Col>
                        <Col span={12}>
                            <div>
                                <label className="gradio-label">CFG Scale</label>
                                <Slider
                                    id="cfg_scale_slider"
                                    min={1.0}
                                    max={30.0}
                                    defaultValue={7.0}
                                    step={0.5}
                                    marks={{ 1: '1', 7: '7', 15: '15', 30: '30' }}
                                />
                            </div>
                        </Col>
                    </Row>

                    <Row gutter={16}>
                        <Col span={12}>
                            <div>
                                <label className="gradio-label">宽度</label>
                                <Slider
                                    id="width_slider"
                                    min={64}
                                    max={2048}
                                    defaultValue={512}
                                    step={64}
                                    marks={{ 64: '64', 512: '512', 1024: '1024', 2048: '2048' }}
                                />
                            </div>
                        </Col>
                        <Col span={12}>
                            <div>
                                <label className="gradio-label">高度</label>
                                <Slider
                                    id="height_slider"
                                    min={64}
                                    max={2048}
                                    defaultValue={512}
                                    step={64}
                                    marks={{ 64: '64', 512: '512', 1024: '1024', 2048: '2048' }}
                                />
                            </div>
                        </Col>
                    </Row>

                    <div>
                        <label className="gradio-label">随机种子</label>
                        <Input
                            id="seed_input"
                            type="number"
                            defaultValue={-1}
                            placeholder="-1 (随机)"
                        />
                    </div>
                </Space>
            </Card>
        );
    }

    createOutputPanel(): React.ReactElement {
        return (
            <Card title="🖼️ 输出面板" className="gradio-output-panel">
                <Space direction="vertical" style={{ width: '100%' }} size="large">
                    <div className="gradio-gallery" id="result_gallery">
                        <div className="gallery-placeholder">
                            <div className="placeholder-content">
                                <div className="placeholder-icon">🎨</div>
                                <div className="placeholder-text">生成结果将显示在这里</div>
                            </div>
                        </div>
                    </div>

                    <Row gutter={16}>
                        <Col span={8}>
                            <Button
                                id="download_all_btn"
                                type="primary"
                                icon={<DownloadOutlined />}
                                block
                            >
                                下载全部
                            </Button>
                        </Col>
                        <Col span={8}>
                            <Button
                                id="clear_btn"
                                icon={<ClearOutlined />}
                                block
                            >
                                清空
                            </Button>
                        </Col>
                        <Col span={8}>
                            <Button
                                id="save_config_btn"
                                block
                            >
                                保存配置
                            </Button>
                        </Col>
                    </Row>
                </Space>
            </Card>
        );
    }

    createControlPanel(): React.ReactElement {
        return (
            <Card title="🎮 控制面板" className="gradio-control-panel">
                <Space direction="vertical" style={{ width: '100%' }} size="large">
                    <Row gutter={16}>
                        <Col span={12}>
                            <Button
                                id="generate_btn"
                                type="primary"
                                size="large"
                                icon={<PlayCircleOutlined />}
                                block
                            >
                                生成
                            </Button>
                        </Col>
                        <Col span={12}>
                            <Button
                                id="stop_btn"
                                danger
                                size="large"
                                icon={<StopOutlined />}
                                block
                                disabled
                            >
                                停止
                            </Button>
                        </Col>
                    </Row>

                    <div>
                        <label className="gradio-label">进度</label>
                        <Progress
                            percent={0}
                            status="active"
                            strokeColor={{
                                '0%': '#108ee9',
                                '100%': '#87d068',
                            }}
                        />
                    </div>

                    <div>
                        <label className="gradio-label">状态</label>
                        <Input
                            id="status_display"
                            value="就绪"
                            readOnly
                            className="status-display"
                        />
                    </div>
                </Space>
            </Card>
        );
    }

    createSettingsPanel(): React.ReactElement {
        return (
            <Card title="⚙️ 高级设置" className="gradio-settings-panel">
                <Space direction="vertical" style={{ width: '100%' }} size="large">
                    <div>
                        <label className="gradio-label">模型选择</label>
                        <Select
                            id="model_dropdown"
                            defaultValue="Model A"
                            style={{ width: '100%' }}
                            options={[
                                { value: 'Model A', label: 'Model A' },
                                { value: 'Model B', label: 'Model B' },
                                { value: 'Model C', label: 'Model C' },
                            ]}
                        />
                    </div>

                    <div>
                        <label className="gradio-label">采样器</label>
                        <Select
                            id="sampler_dropdown"
                            defaultValue="Euler"
                            style={{ width: '100%' }}
                            options={[
                                { value: 'Euler', label: 'Euler' },
                                { value: 'DPM++', label: 'DPM++' },
                                { value: 'DDIM', label: 'DDIM' },
                            ]}
                        />
                    </div>

                    <div>
                        <label className="gradio-label">批次大小</label>
                        <Slider
                            id="batch_size_slider"
                            min={1}
                            max={8}
                            defaultValue={1}
                            marks={{ 1: '1', 4: '4', 8: '8' }}
                        />
                    </div>

                    <div>
                        <Switch
                            id="enable_hr_checkbox"
                            checkedChildren="启用"
                            unCheckedChildren="禁用"
                        />
                        <span style={{ marginLeft: 8 }}>高分辨率修复</span>
                    </div>
                </Space>
            </Card>
        );
    }
}

// 创建单例实例
export const uiComponentFactory = new GradioUIComponentFactory();
