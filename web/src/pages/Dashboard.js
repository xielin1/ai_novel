import React, { useState, useEffect, useContext } from 'react';
import {
  Button,
  Container,
  Card,
  Header,
  Icon,
  Modal,
  Form,
  Segment,
  Message,
  Dimmer,
  Loader,
  Grid,
  Transition,
  Label,
  Input
} from 'semantic-ui-react';
import { useNavigate } from 'react-router-dom';
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
  const [searchTerm, setSearchTerm] = useState('');
  const navigate = useNavigate();
  const [userState] = useContext(UserContext);

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

  // 过滤项目列表
  const filteredProjects = projects.filter(project => {
    if (!searchTerm) return true;
    
    const term = searchTerm.toLowerCase();
    return (
      project.title.toLowerCase().includes(term) ||
      (project.description && project.description.toLowerCase().includes(term)) ||
      (project.genre && project.genre.toLowerCase().includes(term))
    );
  });

  // 格式化日期显示
  const formatDate = dateString => {
    const date = new Date(dateString);
    const now = new Date();
    const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) {
      return '今天';
    } else if (diffDays === 1) {
      return '昨天';
    } else if (diffDays < 7) {
      return `${diffDays}天前`;
    } else {
      return date.toLocaleDateString();
    }
  };

  return (
    <Container style={{ padding: '2em' }}>
      <Segment raised style={{ borderRadius: '12px', boxShadow: '0 4px 8px rgba(0,0,0,0.1)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
          <Header as='h2' style={{ margin: 0 }}>
            <Icon name='book' style={{ color: '#2185d0' }}/>
            <Header.Content>
              我的项目
              <Header.Subheader>管理您的创作项目</Header.Subheader>
            </Header.Content>
          </Header>
          
          <div style={{ display: 'flex', gap: '12px' }}>
            <Input 
              icon='search' 
              placeholder='搜索项目...' 
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              style={{ borderRadius: '20px', minWidth: '200px' }}
            />
            <Button 
              primary 
              icon 
              labelPosition='left'
              onClick={showCreateModal}
              style={{ borderRadius: '20px' }}
            >
              <Icon name='plus' />
              创建新项目
            </Button>
          </div>
        </div>
        
        {loading ? (
          <Dimmer active inverted>
            <Loader>加载中...</Loader>
          </Dimmer>
        ) : filteredProjects.length > 0 ? (
          <Transition.Group
            as={Grid}
            duration={300}
            columns={3}
            stackable
          >
            {filteredProjects.map(project => (
              <Grid.Column key={project.id}>
                <Card fluid style={{ borderRadius: '8px', boxShadow: '0 2px 6px rgba(0,0,0,0.08)', transition: 'transform 0.2s, box-shadow 0.2s' }} 
                      className="project-card"
                >
                  <Card.Content>
                    <Card.Header>
                      {project.title}
                      {project.genre && (
                        <Label size='tiny' color='blue' style={{ marginLeft: '8px', borderRadius: '20px', fontSize: '10px' }}>
                          {project.genre}
                        </Label>
                      )}
                    </Card.Header>
                    <Card.Meta style={{ marginTop: '6px', color: '#888' }}>
                      创建于 {formatDate(project.created_at)}
                      {project.last_edited_at && project.last_edited_at !== project.created_at && (
                        <span> · 更新于 {formatDate(project.last_edited_at)}</span>
                      )}
                    </Card.Meta>
                    <Card.Description style={{ marginTop: '12px', minHeight: '40px', color: '#666' }}>
                      {project.description || <span style={{ fontStyle: 'italic', color: '#aaa' }}>暂无描述</span>}
                    </Card.Description>
                  </Card.Content>
                  <Card.Content extra>
                    <div className="ui three buttons" style={{ marginLeft: '-5px', marginRight: '-5px' }}>
                      <Button basic color='blue' onClick={() => handleOpenProject(project.id)}>
                        <Icon name='edit outline' /> 编辑
                      </Button>
                      <Button basic color='teal' onClick={() => showEditModal(project)}>
                        <Icon name='setting' /> 设置
                      </Button>
                      <Button basic color='red' onClick={() => handleDeleteProject(project.id)}>
                        <Icon name='trash alternate outline' /> 删除
                      </Button>
                    </div>
                  </Card.Content>
                </Card>
              </Grid.Column>
            ))}
          </Transition.Group>
        ) : (
          <Message 
            icon 
            info
            style={{ boxShadow: 'none', borderRadius: '8px', marginTop: '2em' }}
          >
            <Icon name='info circle' />
            <Message.Content>
              <Message.Header>{searchTerm ? '没有匹配的项目' : '暂无项目'}</Message.Header>
              <p>{searchTerm ? '尝试使用其他关键词，或' : ''}点击"创建新项目"按钮开始您的创作之旅！</p>
            </Message.Content>
          </Message>
        )}
      </Segment>

      {/* 创建/编辑项目模态框 */}
      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        size='tiny'
        style={{ borderRadius: '12px' }}
      >
        <Modal.Header style={{ borderBottom: '1px solid #f0f0f0', padding: '20px 24px' }}>
          <Icon name={editingId ? 'edit' : 'plus'} style={{ marginRight: '10px' }} />{modalTitle}
        </Modal.Header>
        <Modal.Content style={{ padding: '24px' }}>
          <Form>
            <Form.Field style={{ marginBottom: '16px' }}>
              <label>项目名称</label>
              <Input
                fluid
                name='title'
                value={formData.title}
                onChange={handleInputChange}
                placeholder='请输入项目名称'
                required
              />
            </Form.Field>
            <Form.Field style={{ marginBottom: '16px' }}>
              <label>项目描述</label>
              <Form.TextArea
                name='description'
                value={formData.description}
                onChange={handleInputChange}
                placeholder='请输入项目描述（选填）'
                rows={3}
                style={{ resize: 'none' }}
              />
            </Form.Field>
            <Form.Field>
              <label>作品类型</label>
              <Input
                fluid
                name='genre'
                value={formData.genre}
                onChange={handleInputChange}
                placeholder='如：玄幻、科幻、都市等（选填）'
              />
            </Form.Field>
          </Form>
        </Modal.Content>
        <Modal.Actions style={{ background: '#f9f9f9', padding: '12px 24px', borderTop: '1px solid #f0f0f0' }}>
          <Button onClick={() => setModalOpen(false)} style={{ borderRadius: '4px' }}>
            取消
          </Button>
          <Button 
            primary 
            onClick={handleSubmit} 
            style={{ borderRadius: '4px' }}
          >
            {editingId ? '保存修改' : '创建项目'}
          </Button>
        </Modal.Actions>
      </Modal>
      
      {/* 添加项目卡片hover效果的CSS */}
      <style jsx global>{`
        .project-card:hover {
          transform: translateY(-3px);
          box-shadow: 0 4px 12px rgba(0,0,0,0.15) !important;
        }
        
        .ui.card > .extra.content {
          border-top: 1px solid #f0f0f0 !important;
          background: #fcfcfc;
        }
      `}</style>
    </Container>
  );
};

export default Dashboard; 