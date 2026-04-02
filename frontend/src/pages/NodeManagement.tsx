import React, { useState, useEffect, useCallback } from 'react';
import {
  Table,
  Button,
  Input,
  Card,
  Typography,
  Tooltip,
  Tag,
  Modal,
  Form,
  message,
  Empty,
  Tabs,
  Space,
  Popconfirm,
} from 'antd';
import {
  ReloadOutlined,
  CopyOutlined,
  PlusOutlined,
  DesktopOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  EyeOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import { nodeAPI, NodeCredential, OnlineNode } from '../api/node';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

const NodeManagement: React.FC = () => {
  // Token 管理状态
  const [credentials, setCredentials] = useState<NodeCredential[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [createForm] = Form.useForm();
  const [createdToken, setCreatedToken] = useState<{ id: number; name: string; token: string } | null>(null);
  const [viewingToken, setViewingToken] = useState<{ id: number; name: string; token: string } | null>(null);
  const [copyingId, setCopyingId] = useState<number | null>(null);
  const [copied, setCopied] = useState(false);

  // 在线节点状态
  const [onlineNodes, setOnlineNodes] = useState<OnlineNode[]>([]);
  const [loadingNodes, setLoadingNodes] = useState(false);

  // 获取节点凭证列表
  const fetchCredentials = useCallback(async () => {
    setLoading(true);
    try {
      const data = await nodeAPI.listCredentials();
      setCredentials(data);
    } catch (error) {
      message.error(error instanceof Error ? error.message : '获取节点凭证失败');
    } finally {
      setLoading(false);
    }
  }, []);

  // 获取在线节点列表
  const fetchOnlineNodes = useCallback(async () => {
    setLoadingNodes(true);
    try {
      const data = await nodeAPI.listOnlineNodes();
      setOnlineNodes(data);
    } catch (error) {
      message.error(error instanceof Error ? error.message : '获取在线节点失败');
    } finally {
      setLoadingNodes(false);
    }
  }, []);

  useEffect(() => {
    fetchCredentials();
    fetchOnlineNodes();
  }, [fetchCredentials, fetchOnlineNodes]);

  // 创建节点凭证
  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields();
      const result = await nodeAPI.createCredential({ name: values.name });
      setCreatedToken({ id: result.id, name: result.name, token: result.token });
      fetchCredentials();
      createForm.resetFields();
    } catch (error) {
      if (error instanceof Error) {
        message.error(error.message);
      }
    }
  };

  // 改进的复制函数，支持降级方案
  const copyText = async (text: string, successMsg = '已复制到剪贴板') => {
    try {
      // 优先使用现代 Clipboard API
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        message.success(successMsg);
        setTimeout(() => setCopied(false), 2000);
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
        setCopied(true);
        message.success(successMsg);
        setTimeout(() => setCopied(false), 2000);
        return true;
      }
      throw new Error('execCommand failed');
    } catch (err) {
      message.error('复制失败，请手动复制');
      return false;
    }
  };

  // 查看完整 Token
  const handleViewToken = async (record: NodeCredential) => {
    const id = record.ID || record.id || 0;
    setCopyingId(id);
    try {
      const response = await nodeAPI.getFullToken(id);
      setViewingToken({
        id,
        name: record.Name || record.name || '',
        token: response.token,
      });
    } catch (error) {
      message.error(error instanceof Error ? error.message : '获取 Token 失败');
    } finally {
      setCopyingId(null);
    }
  };

  // 复制完整 Token
  const handleCopyToken = async (record: NodeCredential) => {
    const id = record.ID || record.id || 0;
    setCopyingId(id);
    try {
      const response = await nodeAPI.getFullToken(id);
      await copyText(response.token, 'Token 已复制到剪贴板');
    } catch (error) {
      message.error(error instanceof Error ? error.message : '复制失败');
    } finally {
      setCopyingId(null);
    }
  };

  // 删除节点凭证
  const handleDelete = async (record: NodeCredential) => {
    const id = record.ID || record.id || 0;
    try {
      await nodeAPI.deleteCredential(id);
      message.success('删除成功');
      setCredentials((prev) => prev.filter((cred) => (cred.ID || cred.id) !== id));
    } catch (error) {
      message.error(error instanceof Error ? error.message : '删除失败');
    }
  };

  // 关闭创建弹窗
  const handleCloseModal = () => {
    setCreateModalVisible(false);
    setCreatedToken(null);
    setCopied(false);
    createForm.resetFields();
  };

  // 过滤后的凭证数据
  const filteredCredentials = credentials.filter(
    (cred) =>
      (cred.Name || cred.name || '').toLowerCase().includes(searchText.toLowerCase())
  );

  // 凭证列表列定义
  const credentialColumns = [
    {
      title: '节点名称',
      dataIndex: 'Name',
      key: 'name',
      render: (_text: string, record: NodeCredential) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <DesktopOutlined style={{ color: '#52c41a' }} />
          <Text strong>{record.Name || record.name}</Text>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'Enabled',
      key: 'enabled',
      width: 100,
      render: (_enabled: boolean, record: NodeCredential) => {
        const enabled = record.Enabled ?? record.enabled;
        return enabled ? (
          <Tag color="success">启用</Tag>
        ) : (
          <Tag color="default">禁用</Tag>
        );
      },
    },
    {
      title: 'Token 前缀',
      dataIndex: 'TokenPrefix',
      key: 'tokenPrefix',
      render: (_text: string, record: NodeCredential) => {
        const prefix = record.TokenPrefix || record.token_prefix || '';
        return (
          <code style={{ background: '#f5f5f5', padding: '2px 8px', borderRadius: 4, fontFamily: 'monospace' }}>
            {prefix}
          </code>
        );
      },
    },
    {
      title: '创建时间',
      dataIndex: 'CreatedAt',
      key: 'createdAt',
      width: 180,
      render: (_text: string, record: NodeCredential) => {
        const date = record.CreatedAt || record.created_at;
        return date ? new Date(date).toLocaleString('zh-CN') : '-';
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_: unknown, record: NodeCredential) => {
        const id = record.ID || record.id || 0;
        return (
          <Space size="small">
            <Tooltip title="查看完整 Token">
              <Button
                icon={<EyeOutlined />}
                size="small"
                loading={copyingId === id && !viewingToken}
                onClick={() => handleViewToken(record)}
              />
            </Tooltip>
            <Tooltip title="复制完整 Token">
              <Button
                icon={<CopyOutlined />}
                size="small"
                loading={copyingId === id && !!viewingToken}
                onClick={() => handleCopyToken(record)}
              />
            </Tooltip>
            <Popconfirm
              title="确认删除"
              description={`确定要删除 "${record.Name || record.name}" 吗？此操作不可恢复。`}
              onConfirm={() => handleDelete(record)}
              okText="删除"
              cancelText="取消"
              okButtonProps={{ danger: true }}
            >
              <Button icon={<DeleteOutlined />} size="small" danger />
            </Popconfirm>
          </Space>
        );
      },
    },
  ];

  // 在线节点列定义
  const onlineNodeColumns = [
    {
      title: '节点名称',
      dataIndex: 'node_name',
      key: 'node_name',
      render: (text: string) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <CheckCircleOutlined style={{ color: '#52c41a' }} />
          <Text strong>{text}</Text>
        </div>
      ),
    },
    {
      title: '支持模型',
      dataIndex: 'models',
      key: 'models',
      render: (models: string[]) => (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {models.map((model) => (
            <Tag key={model} size="small">{model}</Tag>
          ))}
        </div>
      ),
    },
    {
      title: '连接时间',
      dataIndex: 'connected_at',
      key: 'connected_at',
      width: 180,
      render: (text: string) => new Date(text).toLocaleString('zh-CN'),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <Tabs defaultActiveKey="tokens">
          <TabPane
            tab={
              <span>
                <ApiOutlined />
                节点 Token 管理
              </span>
            }
            key="tokens"
          >
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
                  节点 Token 管理
                </Title>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  管理节点连接服务器的凭证 Token
                </Text>
              </div>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setCreateModalVisible(true)}
              >
                生成新 Token
              </Button>
            </div>

            <Input.Search
              placeholder="搜索节点名称"
              allowClear
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              style={{ marginBottom: 16, maxWidth: 400 }}
            />

            <Table
              columns={credentialColumns}
              dataSource={filteredCredentials}
              rowKey="ID"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showTotal: (total) => `共 ${total} 个凭证`,
              }}
            />
          </TabPane>

          <TabPane
            tab={
              <span>
                <DesktopOutlined />
                在线节点 ({onlineNodes.length})
              </span>
            }
            key="online"
          >
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
                  在线节点列表
                </Title>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  当前共有 {onlineNodes.length} 个节点在线
                </Text>
              </div>
              <Button
                icon={<ReloadOutlined />}
                onClick={fetchOnlineNodes}
                loading={loadingNodes}
              >
                刷新
              </Button>
            </div>

            {onlineNodes.length === 0 && !loadingNodes ? (
              <Empty
                description={
                  <div>
                    <Text type="secondary">暂无在线节点</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      请使用 Token 启动节点并连接到服务器
                    </Text>
                  </div>
                }
              />
            ) : (
              <Table
                columns={onlineNodeColumns}
                dataSource={onlineNodes}
                rowKey="node_id"
                loading={loadingNodes}
                pagination={{
                  pageSize: 10,
                  showSizeChanger: true,
                  showTotal: (total) => `共 ${total} 个节点`,
                }}
              />
            )}
          </TabPane>
        </Tabs>
      </Card>

      {/* 生成 Token 弹窗 */}
      <Modal
        title="生成节点 Token"
        open={createModalVisible}
        onOk={handleCreate}
        onCancel={handleCloseModal}
        okText="生成"
        cancelText="取消"
        width={600}
      >
        {!createdToken ? (
          <Form form={createForm} layout="vertical">
            <Form.Item
              name="name"
              label="节点名称"
              rules={[
                { required: true, message: '请输入节点名称' },
                { max: 50, message: '名称最多 50 个字符' },
              ]}
            >
              <Input placeholder="例如：node-local-1" />
            </Form.Item>
            <Paragraph type="secondary" style={{ fontSize: 12 }}>
              生成 Token 后请立即复制保存，Token 只会在此时显示一次，之后可在列表中查看。
            </Paragraph>
          </Form>
        ) : (
          <div style={{ padding: '20px 0' }}>
            <div style={{ textAlign: 'center', marginBottom: 24 }}>
              <CheckCircleOutlined style={{ fontSize: 48, color: '#52c41a' }} />
              <Title level={4} style={{ marginTop: 16 }}>Token 生成成功</Title>
            </div>

            <div style={{ background: '#f6ffed', padding: 16, borderRadius: 8, marginBottom: 16 }}>
              <Text strong>节点名称：</Text>
              <Text>{createdToken.name}</Text>
            </div>

            <div style={{ marginBottom: 16 }}>
              <Text strong>Token（请立即复制保存）：</Text>
              <Space.Compact style={{ width: '100%', marginTop: 8 }}>
                <Input
                  value={createdToken.token}
                  readOnly
                  style={{ flex: 1, fontFamily: 'monospace' }}
                />
                <Button
                  icon={<CopyOutlined />}
                  onClick={() => copyText(createdToken.token)}
                  type={copied ? 'default' : 'primary'}
                >
                  {copied ? '已复制' : '复制'}
                </Button>
              </Space.Compact>
              <div style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
                此 Token 只会显示一次，请立即复制保存。
                如遗失可在列表中点击"查看完整 Token"获取。
              </div>
            </div>

            <div style={{ background: '#fff7e6', padding: 12, borderRadius: 8 }}>
              <Text strong style={{ color: '#fa8c16' }}>使用说明：</Text>
              <Paragraph style={{ fontSize: 12, marginTop: 8, marginBottom: 0 }}>
                在节点配置文件中设置：<br />
                <code>node-agent.token = &quot;{createdToken.token}&quot;</code>
              </Paragraph>
            </div>
          </div>
        )}
      </Modal>

      {/* 查看 Token Modal */}
      <Modal
        open={!!viewingToken}
        title={`节点 Token: ${viewingToken?.name || ''}`}
        onCancel={() => setViewingToken(null)}
        footer={[
          <Button key="copy" type="primary" icon={<CopyOutlined />} onClick={() => {
            if (viewingToken) {
              copyText(viewingToken.token);
            }
          }}>
            复制
          </Button>,
          <Button key="close" onClick={() => setViewingToken(null)}>
            关闭
          </Button>,
        ]}
      >
        <div style={{ marginTop: 16 }}>
          <div style={{ marginBottom: 8, fontWeight: 500 }}>完整 Token</div>
          <Input.TextArea
            value={viewingToken?.token || ''}
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

export default NodeManagement;
