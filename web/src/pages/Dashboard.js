import React, { useState, useEffect, useContext } from 'react';
import { 
  Layout, Menu, Button, Table, Space, Card, 
  Typography, Modal, Form, Input, message, Empty 
} from 'antd';
import { 
  PlusOutlined, EditOutlined, DeleteOutlined, 
  FileTextOutlined, UserOutlined, LogoutOutlined 
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';
import { UserContext } from '../context/User';
import '../styles/Dashboard.css';

const { Header, Content, Sider } = Layout;
const { Title } = Typography;

const Dashboard = () => {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [modalTitle, setModalTitle] = useState('创建新项目');
  const [form] = Form.useForm();
  const [editingId, setEditingId] = useState(null);
  const [userData, setUserData] = useState({});
  const navigate = useNavigate();
  const [userState, userDispatch] = useContext(UserContext);

  // 获取用户信息
  const fetchUserData = async () => {
    // 使用原有系统中的用户信息
    const user = localStorage.getItem('user');
    if (user) {
      setUserData(JSON.parse(user));
    }
  };

  // 获取项目列表
  const fetchProjects = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/projects');
      const { success, message, data } = res.data;
      if (success) {
        setProjects(data || []);
      } else {
        showError(message || '获取项目列表失败');
      }
    } catch (error) {
      console.error('获取项目列表失败', error);
      showError(error.message || '获取项目列表失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUserData();
    fetchProjects();
  }, []);

  // 处理创建/编辑项目
  const handleCreateOrUpdateProject = async (values) => {
    try {
      let res;
      if (editingId) {
        // 更新项目
        res = await API.put(`/api/projects/${editingId}`, values);
      } else {
        // 创建项目
        res = await API.post('/api/projects', values);
      }

      const { success, message, data } = res.data;
      if (success) {
        showSuccess(editingId ? '项目更新成功' : '项目创建成功');
        setModalVisible(false);
        fetchProjects();
      } else {
        showError(message || '操作失败');
      }
    } catch (error) {
      console.error('操作失败', error);
      showError(error.message || '操作失败，请稍后重试');
    }
  };

  // 删除项目
  const handleDeleteProject = async (id) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个项目吗？此操作不可撤销。',
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          const res = await API.delete(`/api/projects/${id}`);
          const { success, message } = res.data;
          if (success) {
            showSuccess('项目删除成功');
            fetchProjects();
          } else {
            showError(message || '删除失败');
          }
        } catch (error) {
          console.error('删除失败', error);
          showError(error.message || '删除失败，请稍后重试');
        }
      }
    });
  };

  // 打开项目
  const handleOpenProject = (id) => {
    navigate(`/editor/${id}`);
  };

  // 显示创建项目模态框
  const showCreateModal = () => {
    setModalTitle('创建新项目');
    setEditingId(null);
    form.resetFields();
    setModalVisible(true);
  };

  // 显示编辑项目模态框
  const showEditModal = (project) => {
    setModalTitle('编辑项目');
    setEditingId(project.id);
    form.setFieldsValue({
      title: project.title,
      description: project.description,
      genre: project.genre,
    });
    setModalVisible(true);
  };

  // 退出登录
  const handleLogout = () => {
    localStorage.removeItem('user');
    userDispatch({ type: 'logout' });
    navigate('/login');
    showSuccess('退出登录成功');
  };

  // 表格列定义
  const columns = [
    {
      title: '项目名称',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'genre',
      key: 'genre',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (text) => new Date(text).toLocaleString(),
    },
    {
      title: '上次编辑',
      dataIndex: 'last_edited_at',
      key: 'last_edited_at',
      render: (text) => text ? new Date(text).toLocaleString() : '未编辑',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button
            type="primary"
            icon={<FileTextOutlined />}
            onClick={() => handleOpenProject(record.id)}
          >
            打开
          </Button>
          <Button
            icon={<EditOutlined />}
            onClick={() => showEditModal(record)}
          >
            编辑
          </Button>
          <Button
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDeleteProject(record.id)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header className="dashboard-header">
        <div className="logo">网文大纲续写助手</div>
        <div className="user-info">
          <Space>
            <UserOutlined />
            <span>{userData.username || '用户'}</span>
            <Button type="link" icon={<LogoutOutlined />} onClick={handleLogout}>
              退出登录
            </Button>
          </Space>
        </div>
      </Header>
      <Layout>
        <Sider width={200} className="dashboard-sider">
          <Menu
            mode="inline"
            defaultSelectedKeys={['projects']}
            style={{ height: '100%', borderRight: 0 }}
          >
            <Menu.Item key="projects" icon={<FileTextOutlined />}>
              我的项目
            </Menu.Item>
            <Menu.Item key="account" icon={<UserOutlined />}>
              账户信息
            </Menu.Item>
          </Menu>
        </Sider>
        <Layout className="dashboard-content-layout">
          <Content className="dashboard-content">
            <div className="dashboard-header-actions">
              <Title level={4}>我的项目</Title>
              <Button 
                type="primary" 
                icon={<PlusOutlined />} 
                onClick={showCreateModal}
              >
                创建新项目
              </Button>
            </div>
            
            <Card className="project-list-card">
              {projects.length > 0 ? (
                <Table
                  dataSource={projects}
                  columns={columns}
                  rowKey="id"
                  loading={loading}
                  pagination={{ pageSize: 10 }}
                />
              ) : (
                <Empty
                  description="暂无项目"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                >
                  <Button 
                    type="primary" 
                    icon={<PlusOutlined />} 
                    onClick={showCreateModal}
                  >
                    创建新项目
                  </Button>
                </Empty>
              )}
            </Card>
          </Content>
        </Layout>
      </Layout>

      {/* 创建/编辑项目模态框 */}
      <Modal
        title={modalTitle}
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateOrUpdateProject}
        >
          <Form.Item
            name="title"
            label="项目名称"
            rules={[{ required: true, message: '请输入项目名称' }]}
          >
            <Input placeholder="请输入项目名称" />
          </Form.Item>
          
          <Form.Item
            name="description"
            label="项目描述"
          >
            <Input.TextArea placeholder="请输入项目描述" rows={4} />
          </Form.Item>
          
          <Form.Item
            name="genre"
            label="作品类型"
          >
            <Input placeholder="如：玄幻、科幻、都市等" />
          </Form.Item>
          
          <Form.Item>
            <Button type="primary" htmlType="submit" style={{ marginRight: 8 }}>
              {editingId ? '保存修改' : '创建项目'}
            </Button>
            <Button onClick={() => setModalVisible(false)}>
              取消
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </Layout>
  );
};

export default Dashboard; 