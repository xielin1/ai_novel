import React, { useState, useEffect, useContext } from 'react';
import {
  Button,
  Container,
  Divider,
  Grid,
  Header,
  Icon,
  Menu,
  Modal,
  Form,
  Table,
  Segment,
  Message,
  Dimmer,
  Loader
} from 'semantic-ui-react';
import { Link, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';
import { UserContext } from '../context/User';

const Dashboard = () => {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState('创建新项目');
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    genre: ''
  });
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

  // 处理输入变化
  const handleInputChange = (e, { name, value }) => {
    setFormData({ ...formData, [name]: value });
  };

  // 处理创建/编辑项目
  const handleSubmit = async () => {
    if (!formData.title) {
      showError('项目名称不能为空');
      return;
    }

    try {
      let res;
      if (editingId) {
        // 更新项目
        res = await API.put(`/api/projects/${editingId}`, formData);
      } else {
        // 创建项目
        res = await API.post('/api/projects', formData);
      }

      const { success, message, data } = res.data;
      if (success) {
        showSuccess(editingId ? '项目更新成功' : '项目创建成功');
        setModalOpen(false);
        fetchProjects();
        // 重置表单
        setFormData({
          title: '',
          description: '',
          genre: ''
        });
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
    if (window.confirm('确定要删除这个项目吗？此操作不可撤销。')) {
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
  };

  // 打开项目
  const handleOpenProject = (id) => {
    navigate(`/editor/${id}`);
  };

  // 显示创建项目模态框
  const showCreateModal = () => {
    setModalTitle('创建新项目');
    setEditingId(null);
    setFormData({
      title: '',
      description: '',
      genre: ''
    });
    setModalOpen(true);
  };

  // 显示编辑项目模态框
  const showEditModal = (project) => {
    setModalTitle('编辑项目');
    setEditingId(project.id);
    setFormData({
      title: project.title,
      description: project.description || '',
      genre: project.genre || ''
    });
    setModalOpen(true);
  };

  // 退出登录
  const handleLogout = () => {
    localStorage.removeItem('user');
    userDispatch({ type: 'logout' });
    navigate('/login');
    showSuccess('退出登录成功');
  };

  return (
    <Container fluid style={{ padding: '2em' }}>
      <Grid>
        <Grid.Row>
          <Grid.Column width={16}>
            <Segment clearing>
              <Header as='h2' floated='left'>
                网文大纲续写助手
              </Header>
              <Button 
                floated='right' 
                onClick={handleLogout}
                color='red'
              >
                <Icon name='sign-out' /> 退出登录
              </Button>
              <span style={{ marginRight: '1em', float: 'right', lineHeight: '36px' }}>
                <Icon name='user' /> {userData.username || '用户'}
              </span>
            </Segment>
          </Grid.Column>
        </Grid.Row>

        <Grid.Row>
          <Grid.Column width={3}>
            <Menu vertical fluid>
              <Menu.Item active>
                <Icon name='file alternate' /> 我的项目
              </Menu.Item>
              <Menu.Item>
                <Icon name='user' /> 账户信息
              </Menu.Item>
            </Menu>
          </Grid.Column>
          
          <Grid.Column width={13}>
            <Segment>
              <Header as='h3' floated='left'>我的项目</Header>
              <Button 
                primary 
                floated='right'
                icon 
                labelPosition='left'
                onClick={showCreateModal}
              >
                <Icon name='plus' />
                创建新项目
              </Button>
              <Divider clearing />
              
              {loading ? (
                <Dimmer active inverted>
                  <Loader>加载中...</Loader>
                </Dimmer>
              ) : projects.length > 0 ? (
                <Table celled>
                  <Table.Header>
                    <Table.Row>
                      <Table.HeaderCell>项目名称</Table.HeaderCell>
                      <Table.HeaderCell>描述</Table.HeaderCell>
                      <Table.HeaderCell>类型</Table.HeaderCell>
                      <Table.HeaderCell>创建时间</Table.HeaderCell>
                      <Table.HeaderCell>上次编辑</Table.HeaderCell>
                      <Table.HeaderCell>操作</Table.HeaderCell>
                    </Table.Row>
                  </Table.Header>
                  <Table.Body>
                    {projects.map(project => (
                      <Table.Row key={project.id}>
                        <Table.Cell>{project.title}</Table.Cell>
                        <Table.Cell>{project.description}</Table.Cell>
                        <Table.Cell>{project.genre}</Table.Cell>
                        <Table.Cell>
                          {new Date(project.created_at).toLocaleString()}
                        </Table.Cell>
                        <Table.Cell>
                          {project.last_edited_at ? new Date(project.last_edited_at).toLocaleString() : '未编辑'}
                        </Table.Cell>
                        <Table.Cell>
                          <Button.Group>
                            <Button primary onClick={() => handleOpenProject(project.id)}>
                              <Icon name='file text' /> 打开
                            </Button>
                            <Button onClick={() => showEditModal(project)}>
                              <Icon name='edit' /> 编辑
                            </Button>
                            <Button negative onClick={() => handleDeleteProject(project.id)}>
                              <Icon name='trash' /> 删除
                            </Button>
                          </Button.Group>
                        </Table.Cell>
                      </Table.Row>
                    ))}
                  </Table.Body>
                </Table>
              ) : (
                <Message info>
                  <Message.Header>暂无项目</Message.Header>
                  <p>点击"创建新项目"按钮开始吧！</p>
                </Message>
              )}
            </Segment>
          </Grid.Column>
        </Grid.Row>
      </Grid>

      {/* 创建/编辑项目模态框 */}
      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
      >
        <Modal.Header>{modalTitle}</Modal.Header>
        <Modal.Content>
          <Form>
            <Form.Input
              label='项目名称'
              name='title'
              value={formData.title}
              onChange={handleInputChange}
              placeholder='请输入项目名称'
              required
            />
            <Form.TextArea
              label='项目描述'
              name='description'
              value={formData.description}
              onChange={handleInputChange}
              placeholder='请输入项目描述'
              rows={4}
            />
            <Form.Input
              label='作品类型'
              name='genre'
              value={formData.genre}
              onChange={handleInputChange}
              placeholder='如：玄幻、科幻、都市等'
            />
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setModalOpen(false)}>
            取消
          </Button>
          <Button primary onClick={handleSubmit}>
            {editingId ? '保存修改' : '创建项目'}
          </Button>
        </Modal.Actions>
      </Modal>
    </Container>
  );
};

export default Dashboard; 