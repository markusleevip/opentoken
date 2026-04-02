import React, { useState, useEffect, useCallback } from 'react';
import {
  Table,
  Button,
  Input,
  Tag,
  Card,
  Typography,
  Tooltip,
  Badge,
  Empty,
} from 'antd';
import { ReloadOutlined, CopyOutlined, RobotOutlined } from '@ant-design/icons';
import { modelsAPI, Model } from '../api/models';

const { Title, Text } = Typography;

const ModelList: React.FC = () => {
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');

  // 获取模型列表
  const fetchModels = useCallback(async () => {
    setLoading(true);
    try {
      const response = await modelsAPI.list();
      console.log('API 返回:', response);
      console.log('模型数据:', response.data);
      setModels(response.data);
    } catch (error) {
      console.error('获取模型列表失败:', error);
      // 添加错误提示
      // eslint-disable-next-line no-alert
      alert('获取模型列表失败: ' + (error instanceof Error ? error.message : String(error)));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchModels();
  }, [fetchModels]);

  // 复制到剪贴板
  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      // 可以在这里添加 toast 提示
    }).catch(() => {
      console.error('复制失败');
    });
  };

  // 格式化时间戳
  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN');
  };

  // 过滤后的数据
  const filteredData = models.filter(
    (model) =>
      model.id.toLowerCase().includes(searchText.toLowerCase())
  );

  const columns = [
    {
      title: '模型名称',
      dataIndex: 'id',
      key: 'id',
      render: (text: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <RobotOutlined style={{ color: '#1890ff' }} />
          <Text strong>{text}</Text>
        </div>
      ),
    },
    {
      title: '状态',
      key: 'status',
      width: 100,
      render: () => (
        <Badge status="success" text="可用" />
      ),
    },
    {
      title: '提供者',
      dataIndex: 'owned_by',
      key: 'owned_by',
      width: 120,
      render: (text: string) => (
        <Tag color="blue">{text}</Tag>
      ),
    },
    {
      title: '添加时间',
      dataIndex: 'created',
      key: 'created',
      width: 180,
      render: (timestamp: number) => formatDate(timestamp),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: Model) => (
        <Tooltip title="复制模型名称">
          <Button
            icon={<CopyOutlined />}
            size="small"
            onClick={() => copyToClipboard(record.id)}
          />
        </Tooltip>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 24,
          }}
        >
          <div>
            <Title level={4} style={{ margin: 0 }}>
              可用模型列表
            </Title>
            <Text type="secondary" style={{ fontSize: 12 }}>
              当前在线节点共提供 {models.length} 个可用模型
            </Text>
          </div>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchModels}
            loading={loading}
          >
            刷新
          </Button>
        </div>

        <Input.Search
          placeholder="搜索模型名称"
          allowClear
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          style={{ marginBottom: 16, maxWidth: 400 }}
        />

        {models.length === 0 && !loading ? (
          <Empty
            description={
              <div>
                <Text type="secondary">暂无可用模型</Text>
                <br />
                <Text type="secondary" style={{ fontSize: 12 }}>
                  请确保有节点已连接并上报了支持的模型
                </Text>
              </div>
            }
          />
        ) : (
          <Table
            columns={columns}
            dataSource={filteredData}
            rowKey="id"
            loading={loading}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 个模型`,
            }}
          />
        )}
      </Card>
    </div>
  );
};

export default ModelList;
