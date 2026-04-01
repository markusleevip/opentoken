import React, { useState } from 'react';
import {
  Card,
  Typography,
  Tabs,
  Alert,
  Steps,
  Divider,
  Tag,
  Button,
  message,
} from 'antd';
import {
  BookOutlined,
  CopyOutlined,
  CheckOutlined,
  LinkOutlined,
  KeyOutlined,
  CodeOutlined,
} from '@ant-design/icons';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;
const { Step } = Steps;

// 代码示例
const CURL_EXAMPLE = `curl -X POST http://localhost:8084/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "qwen3.5-plus",
    "stream": true,
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'`;

const PYTHON_EXAMPLE = `import openai

# 设置 Base URL 和 API Key
client = openai.OpenAI(
    base_url="http://localhost:8084/v1",
    api_key="YOUR_API_KEY"
)

# 发送请求
response = client.chat.completions.create(
    model="qwen3.5-plus",
    stream=True,
    messages=[
        {"role": "user", "content": "Hello, how are you?"}
    ]
)

# 处理流式响应
for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")`;

const JAVASCRIPT_EXAMPLE = `// 使用 fetch API
const response = await fetch('http://localhost:8084/v1/chat/completions', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_API_KEY'
  },
  body: JSON.stringify({
    model: 'qwen3.5-plus',
    stream: true,
    messages: [
      { role: 'user', content: 'Hello, how are you?' }
    ]
  })
});

// 处理流式响应
const reader = response.body?.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader?.read() || {};
  if (done) break;

  const chunk = decoder.decode(value);
  // 解析 SSE 格式的响应
  const lines = chunk.split('\\n');
  for (const line of lines) {
    if (line.startsWith('data: ')) {
      const data = line.slice(6);
      if (data === '[DONE]') continue;
      const json = JSON.parse(data);
      console.log(json.choices[0]?.delta?.content || '');
    }
  }
}`;

const API_CONFIG_EXAMPLE = `# OpenAI 兼容配置
# 适用于任何支持 OpenAI API 格式的客户端

Base URL: http://localhost:8084/v1
API Key: 从管理后台获取的 API Key

# 支持的模型（取决于在线节点）
# 可通过 /v1/models 查看可用模型列表`;

const APIDocs: React.FC = () => {
  const [copiedIndex, setCopiedIndex] = useState<string | null>(null);

  const copyToClipboard = async (text: string, index: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedIndex(index);
      message.success('已复制到剪贴板');
      setTimeout(() => setCopiedIndex(null), 2000);
    } catch {
      message.error('复制失败');
    }
  };

  const CodeBlock: React.FC<{ code: string; language: string; copyKey: string }> = ({
    code,
    language,
    copyKey,
  }) => (
    <div style={{ position: 'relative', marginTop: 16 }}>
      <div
        style={{
          position: 'absolute',
          top: 8,
          right: 8,
          zIndex: 1,
        }}
      >
        <Button
          size="small"
          icon={copiedIndex === copyKey ? <CheckOutlined /> : <CopyOutlined />}
          onClick={() => copyToClipboard(code, copyKey)}
        >
          {copiedIndex === copyKey ? '已复制' : '复制'}
        </Button>
      </div>
      <pre
        style={{
          background: '#f6f8fa',
          padding: 16,
          borderRadius: 8,
          overflow: 'auto',
          fontSize: 13,
          lineHeight: 1.6,
          fontFamily: 'Consolas, Monaco, "Courier New", monospace',
        }}
      >
        <code>{code}</code>
      </pre>
    </div>
  );

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <div style={{ marginBottom: 24 }}>
          <Title level={4} style={{ margin: 0 }}>
            <BookOutlined style={{ marginRight: 8 }} />
            API 使用文档
          </Title>
          <Text type="secondary">
            了解如何使用 OpenToken API 访问大语言模型
          </Text>
        </div>

        <Alert
          message="OpenAI 兼容 API"
          description="OpenToken 提供与 OpenAI 兼容的 API 接口，您可以直接使用任何支持 OpenAI API 格式的客户端或 SDK。"
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Steps direction="vertical" style={{ marginBottom: 24 }}>
          <Step
            title="获取 API Key"
            description={
              <div>
                <Text>
                  在 <a href="#/api-keys">API Key 管理</a> 页面创建一个新的 API Key。
                </Text>
                <br />
                <Text type="secondary">
                  创建后会显示完整的 Token，请立即复制保存，之后无法再次查看。
                </Text>
              </div>
            }
            icon={<KeyOutlined />}
          />
          <Step
            title="配置 Base URL"
            description={
              <div>
                <Text>将 Base URL 设置为您的 OpenToken 服务器地址：</Text>
                <Tag color="blue" style={{ marginLeft: 8 }}>
                  <LinkOutlined /> http://localhost:8084/v1
                </Tag>
                <br />
                <Text type="secondary">
                  如果前端和后端不在同一域名，请使用后端实际 IP 和端口
                </Text>
              </div>
            }
            icon={<LinkOutlined />}
          />
          <Step
            title="发送请求"
            description="使用您熟悉的编程语言或工具发送请求"
            icon={<CodeOutlined />}
          />
        </Steps>

        <Divider />

        <Tabs defaultActiveKey="curl">
          <TabPane tab="cURL" key="curl">
            <Paragraph>
              <Text strong>使用命令行快速测试：</Text>
            </Paragraph>
            <CodeBlock code={CURL_EXAMPLE} language="bash" copyKey="curl" />
          </TabPane>

          <TabPane tab="Python" key="python">
            <Paragraph>
              <Text strong>使用 OpenAI Python SDK：</Text>
              <br />
              <Text type="secondary">
                先安装 SDK：<code>pip install openai</code>
              </Text>
            </Paragraph>
            <CodeBlock code={PYTHON_EXAMPLE} language="python" copyKey="python" />
          </TabPane>

          <TabPane tab="JavaScript" key="javascript">
            <Paragraph>
              <Text strong>使用原生 fetch API：</Text>
            </Paragraph>
            <CodeBlock code={JAVASCRIPT_EXAMPLE} language="javascript" copyKey="js" />
          </TabPane>

          <TabPane tab="配置说明" key="config">
            <Paragraph>
              <Text strong>通用配置参数：</Text>
            </Paragraph>
            <CodeBlock code={API_CONFIG_EXAMPLE} language="yaml" copyKey="config" />
          </TabPane>
        </Tabs>

        <Divider />

        <div style={{ background: '#f6ffed', padding: 16, borderRadius: 8 }}>
          <Title level={5} style={{ marginTop: 0 }}>
            支持的端点
          </Title>
          <ul style={{ marginBottom: 0 }}>
            <li>
              <code>GET /v1/models</code> - 获取可用模型列表
            </li>
            <li>
              <code>POST /v1/chat/completions</code> - 聊天补全（支持流式输出）
            </li>
          </ul>
        </div>

        <div style={{ marginTop: 16, background: '#fff7e6', padding: 16, borderRadius: 8 }}>
          <Title level={5} style={{ marginTop: 0 }}>
            注意事项
          </Title>
          <ul style={{ marginBottom: 0 }}>
            <li>所有请求必须在 Header 中包含 <code>Authorization: Bearer YOUR_API_KEY</code></li>
            <li>支持的模型取决于当前在线的节点，可通过 <a href="#/models">模型列表</a> 查看</li>
            <li>流式响应使用 Server-Sent Events (SSE) 格式</li>
          </ul>
        </div>
      </Card>
    </div>
  );
};

export default APIDocs;
