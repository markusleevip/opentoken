import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom';
import { Menu } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  KeyOutlined,
  RobotOutlined,
  DesktopOutlined,
  BookOutlined,
} from '@ant-design/icons';
import APIKeyList from '../pages/APIKeyList';
import ModelList from '../pages/ModelList';
import NodeManagement from '../pages/NodeManagement';
import APIDocs from '../pages/APIDocs';

// 带侧边栏的布局组件
const Layout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/api-keys',
      icon: <KeyOutlined />,
      label: 'API Key 管理',
    },
    {
      key: '/nodes',
      icon: <DesktopOutlined />,
      label: '节点管理',
    },
    {
      key: '/models',
      icon: <RobotOutlined />,
      label: '模型列表',
    },
    {
      key: '/docs',
      icon: <BookOutlined />,
      label: 'API 文档',
    },
  ];

  return (
    <div style={{ minHeight: '100vh', background: '#f0f2f5' }}>
      {/* 顶部标题栏 */}
      <div style={{ padding: '16px 24px', background: '#fff', borderBottom: '1px solid #e8e8e8' }}>
        <h1 style={{ margin: 0, fontSize: 20 }}>OpenToken 管理后台</h1>
      </div>

      <div style={{ display: 'flex' }}>
        {/* 左侧导航菜单 */}
        <div style={{ width: 200, background: '#fff', minHeight: 'calc(100vh - 64px)' }}>
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
            style={{ height: '100%', borderRight: 0 }}
          />
        </div>

        {/* 主内容区 */}
        <div style={{ flex: 1, padding: 0 }}>
          <Outlet />
        </div>
      </div>
    </div>
  );
};

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Navigate to="/api-keys" replace />,
      },
      {
        path: 'api-keys',
        element: <APIKeyList />,
      },
      {
        path: 'nodes',
        element: <NodeManagement />,
      },
      {
        path: 'models',
        element: <ModelList />,
      },
      {
        path: 'docs',
        element: <APIDocs />,
      },
    ],
  },
]);
