import React, { useState } from 'react';
import {
  Modal,
  Form,
  Input,
  Button,
  message,
  Alert,
  Space,
  Typography,
  Result,
} from 'antd';
import { CopyOutlined } from '@ant-design/icons';
import { apiKeyAPI, CreateAPIKeyResponse } from '../api/apikey';

const { TextArea } = Input;
const { Paragraph } = Typography;

interface CreateAPIKeyModalProps {
  visible: boolean;
  onClose: () => void;
  onSuccess: (newKey: { id: number; name: string; token: string }) => void;
}

const CreateAPIKeyModal: React.FC<CreateAPIKeyModalProps> = ({
  visible,
  onClose,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  const [createdKey, setCreatedKey] = useState<CreateAPIKeyResponse | null>(null);
  const [copied, setCopied] = useState(false);

  const handleSubmit = async (values: { name: string; description?: string }) => {
    setSubmitting(true);
    try {
      const data = await apiKeyAPI.create({
        name: values.name,
        description: values.description,
      });
      setCreatedKey(data);
      onSuccess({ id: data.id, name: data.name, token: data.token });
    } catch (error) {
      message.error(error instanceof Error ? error.message : '创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    form.resetFields();
    setCreatedKey(null);
    setCopied(false);
    onClose();
  };

  const copyToClipboard = async (text: string) => {
    try {
      // 优先使用现代 Clipboard API
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        message.success('已复制到剪贴板');
        setTimeout(() => setCopied(false), 2000);
        return;
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
        message.success('已复制到剪贴板');
        setTimeout(() => setCopied(false), 2000);
      } else {
        throw new Error('execCommand failed');
      }
    } catch (err) {
      message.error('复制失败，请手动复制');
    }
  };

  // 创建成功后的展示界面
  if (createdKey) {
    return (
      <Modal
        open={visible}
        title="API Key 创建成功"
        onCancel={handleClose}
        footer={[
          <Button key="close" type="primary" onClick={handleClose}>
            完成
          </Button>,
        ]}
        width={560}
      >
        <Result
          status="success"
          title="创建成功"
          subTitle="请立即复制保存您的 API Key，此令牌只会显示一次！"
          style={{ padding: '24px 0' }}
        />

        <Alert
          message="重要提示"
          description="API Key 是访问 LLM 服务的凭证，请妥善保管。如果泄露，请立即删除并重新创建。"
          type="warning"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <div style={{ marginBottom: 16 }}>
          <div style={{ marginBottom: 8, fontWeight: 500 }}>API Key</div>
          <Space.Compact style={{ width: '100%' }}>
            <Input
              value={createdKey.token}
              readOnly
              style={{ flex: 1, fontFamily: 'monospace' }}
            />
            <Button
              icon={<CopyOutlined />}
              onClick={() => copyToClipboard(createdKey.token)}
              type={copied ? 'default' : 'primary'}
            >
              {copied ? '已复制' : '复制'}
            </Button>
          </Space.Compact>
          <div style={{ marginTop: 8, fontSize: 12, color: '#666' }}>
            此 API Key 只会显示一次，请立即复制保存。
            如遗失可在列表中点击"查看完整 Token"获取。
          </div>
        </div>

        <div style={{ marginBottom: 16 }}>
          <div style={{ marginBottom: 8, fontWeight: 500 }}>名称</div>
          <Input value={createdKey.name} readOnly />
        </div>

        {createdKey.description && (
          <div style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 8, fontWeight: 500 }}>描述</div>
            <Input.TextArea value={createdKey.description} readOnly rows={2} />
          </div>
        )}

        <div style={{ marginBottom: 16 }}>
          <div style={{ marginBottom: 8, fontWeight: 500 }}>创建时间</div>
          <Input value={new Date(createdKey.created_at).toLocaleString('zh-CN')} readOnly />
        </div>
      </Modal>
    );
  }

  return (
    <Modal
      open={visible}
      title="创建 API Key"
      onCancel={handleClose}
      footer={null}
      width={480}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        style={{ marginTop: 16 }}
      >
        <Form.Item
          name="name"
          label="名称"
          rules={[
            { required: true, message: '请输入名称' },
            { max: 64, message: '名称最多 64 个字符' },
          ]}
        >
          <Input placeholder="例如：production-key" />
        </Form.Item>

        <Form.Item
          name="description"
          label="描述"
          rules={[{ max: 255, message: '描述最多 255 个字符' }]}
        >
          <TextArea
            placeholder="可选：描述这个 API Key 的用途"
            rows={3}
          />
        </Form.Item>

        <Form.Item style={{ marginTop: 24 }}>
          <Space style={{ width: '100%', justifyContent: 'flex-end' }}>
            <Button onClick={handleClose}>取消</Button>
            <Button type="primary" htmlType="submit" loading={submitting}>
              创建
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default CreateAPIKeyModal;
