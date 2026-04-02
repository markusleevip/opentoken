import React, { useState, useEffect, useCallback } from 'react';
import {
  Table,
  Button,
  Input,
  Switch,
  Popconfirm,
  Space,
  Tag,
  message,
  Card,
  Typography,
  Tooltip,
  Modal,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  CopyOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { apiKeyAPI, APIKey } from '../api/apikey';
import CreateAPIKeyModal from '../components/CreateAPIKeyModal';

const { Title } = Typography;

const APIKeyList: React.FC = () => {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [copyingId, setCopyingId] = useState<number | null>(null);
  const [viewingKey, setViewingKey] = useState<{ id: number; name: string; token: string } | null>(null);

  // 获取 API Key 列表
  const fetchAPIKeys = useCallback(async () => {
    setLoading(true);
    try {
      const data = await apiKeyAPI.list();
      setApiKeys(data);
    } catch (error) {
      message.error(error instanceof Error ? error.message : '获取 API Key 失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAPIKeys();
  }, [fetchAPIKeys]);

  // 切换启用状态
  const handleToggleEnabled = async (id: number, enabled: boolean) => {
    try {
      await apiKeyAPI.update(id, { enabled });
      message.success(enabled ? '已启用' : '已禁用');
      // 更新本地状态
      setApiKeys((prev) =>
        prev.map((key) =>
          key.id === id ? { ...key, enabled } : key
        )
      );
    } catch (error) {
      message.error(error instanceof Error ? error.message : '操作失败');
    }
  };

  // 删除 API Key
  const handleDelete = async (id: number) => {
    try {
      await apiKeyAPI.delete(id);
      message.success('删除成功');
      setApiKeys((prev) => prev.filter((key) => key.id !== id));
    } catch (error) {
      message.error(error instanceof Error ? error.message : '删除失败');
    }
  };

  // 改进的复制函数，支持降级方案
  const copyText = async (text: string, successMsg = '已复制到剪贴板') => {
    try {
      // 优先使用现代 Clipboard API
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        message.success(successMsg);
        return true;
      }

      // 降级方案：使用传统的 execCommand
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();

      const successful = document.execCommand('copy');
      document.body.removeChild(textArea);

      if (successful) {
        message.success(successMsg);
        return true;
      }
      throw new Error('execCommand failed');
    } catch (err) {
      message.error('复制失败，请手动复制');
      return false;
    }
  };

  // 复制完整 Token
  const handleCopyFullToken = async (id: number, name: string) => {
    setCopyingId(id);
    try {
      const response = await apiKeyAPI.getFullToken(id);
      await copyText(response.token, '完整 Token 已复制到剪贴板');
    } catch (error) {
      message.error(error instanceof Error ? error.message : '复制失败');
    } finally {
      setCopyingId(null);
    }
  };

  // 查看完整 Token
  const handleViewToken = async (record: APIKey) => {
    setCopyingId(record.id);
    try {
      const response = await apiKeyAPI.getFullToken(record.id);
      setViewingKey({
        id: record.id,
        name: record.name,
        token: response.token,
      });
    } catch (error) {
      message.error(error instanceof Error ? error.message : '获取 Token 失败');
    } finally {
      setCopyingId(null);
    }
  };

  // 创建成功回调
  const handleCreateSuccess = (newKey: { id: number; name: string; token: string }) => {
    fetchAPIKeys();
  };

  // 过滤后的数据
  const filteredData = apiKeys.filter(
    (key) =>
      key.name.toLowerCase().includes(searchText.toLowerCase()) ||
      key.description?.toLowerCase().includes(searchText.toLowerCase())
  );

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: APIKey) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          {record.description && (
            <div style={{ fontSize: 12, color: '#888', marginTop: 4 }}>
              {record.description}
            </div>
          )}
        </div>
      ),
    },
    {
      title: 'Token 前缀',
      dataIndex: 'token_prefix',
      key: 'token_prefix',
      render: (text: string) => (
        <code style={{ background: '#f5f5f5', padding: '2px 8px', borderRadius: 4 }}>
          {text}
        </code>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 100,
      render: (enabled: boolean, record: APIKey) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggleEnabled(record.id, checked)}
          checkedChildren="启用"
          unCheckedChildren="禁用"
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text: string) => new Date(text).toLocaleString('zh-CN'),
    },
    {
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      width: 180,
      render: (text: string | null) =>
        text ? new Date(text).toLocaleString('zh-CN') : <Tag color="default">从未使用</Tag>,
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_: unknown, record: APIKey) => (
        <Space size="small">
          <Tooltip title="查看完整 Token">
            <Button
              icon={<EyeOutlined />}
              size="small"
              loading={copyingId === record.id && !viewingKey}
              onClick={() => handleViewToken(record)}
            />
          </Tooltip>
          <Tooltip title="复制完整 Token">
            <Button
              icon={<CopyOutlined />}
              size="small"
              loading={copyingId === record.id && !!viewingKey}
              onClick={() => handleCopyFullToken(record.id, record.name)}
            />
          </Tooltip>
          <Popconfirm
            title="确认删除"
            description={`确定要删除 "${record.name}" 吗？此操作不可恢复。`}
            onConfirm={() => handleDelete(record.id)}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Button icon={<DeleteOutlined />} size="small" danger />
          </Popconfirm>
        </Space>
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
          <Title level={4} style={{ margin: 0 }}>
            API Key 管理
          </Title>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            创建 API Key
          </Button>
        </div>

        <Input.Search
          placeholder="搜索名称或描述"
          allowClear
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          style={{ marginBottom: 16, maxWidth: 400 }}
        />

        <Table
          columns={columns}
          dataSource={filteredData}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      <CreateAPIKeyModal
        visible={createModalVisible}
        onClose={() => setCreateModalVisible(false)}
        onSuccess={handleCreateSuccess}
      />

      {/* 查看 Token Modal */}
      <Modal
        open={!!viewingKey}
        title={`API Key: ${viewingKey?.name || ''}`}
        onCancel={() => setViewingKey(null)}
        footer={[
          <Button key="copy" type="primary" icon={<CopyOutlined />} onClick={() => {
            if (viewingKey) {
              copyText(viewingKey.token);
            }
          }}>
            复制
          </Button>,
          <Button key="close" onClick={() => setViewingKey(null)}>
            关闭
          </Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          <div style={{ marginBottom: 8, fontWeight: 500 }}>完整 Token</div>
          <Input.TextArea
            value={viewingKey?.token || ''}
            readOnly
            rows={3}
            style={{ fontFamily: 'monospace', fontSize: 14 }}
          />
          <div style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
            提示：此 Token 已加密存储在数据库中，可随时查看。
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default APIKeyList;
