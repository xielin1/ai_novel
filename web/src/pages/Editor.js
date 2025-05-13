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
  TextArea,
  Segment,
  Message,
  Dimmer,
  Loader,
  Dropdown,
  Input,
  Tab,
  Card
} from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';
import { UserContext } from '../context/User';
import '../styles/Editor.css';

const Editor = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [userState] = useContext(UserContext);
  
  const [project, setProject] = useState(null);
  const [outline, setOutline] = useState('');
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [generatedContent, setGeneratedContent] = useState('');
  const [aiSettings, setAiSettings] = useState({
    style: 'default',
    wordLimit: 1000
  });
  const [versions, setVersions] = useState([]);
  const [selectedVersion, setSelectedVersion] = useState(null);
  const [showVersionHistory, setShowVersionHistory] = useState(false);
  const [uploadFile, setUploadFile] = useState(null);
  const [saving, setSaving] = useState(false);
  const [aiModalOpen, setAiModalOpen] = useState(false);

  // 获取项目和大纲信息
  const fetchProjectAndOutline = async () => {
    setLoading(true);
    try {
      // 获取项目信息
      const projectRes = await API.get(`/api/projects/${id}`);
      const { success: projectSuccess, message: projectMessage, data: projectData } = projectRes.data;
      if (projectSuccess) {
        setProject(projectData);
      } else {
        showError(projectMessage || '获取项目信息失败');
        return;
      }

      // 获取大纲内容
      const outlineRes = await API.get(`/api/outlines/${id}`);
      const { success: outlineSuccess, data: outlineData } = outlineRes.data;
      if (outlineSuccess) {
        setOutline(outlineData?.content || '');
      } else if (outlineRes.data.status !== 404) {
        // 如果不是因为没找到(新项目)，则显示错误
        showError(outlineRes.data.message || '获取大纲内容失败');
      }

      // 获取版本历史
      const versionsRes = await API.get(`/api/versions/${id}`);
      const { success: versionsSuccess, data: versionsData } = versionsRes.data;
      if (versionsSuccess) {
        setVersions(versionsData || []);
      }
    } catch (error) {
      console.error('获取项目和大纲信息失败', error);
      showError(error.message || '获取项目和大纲信息失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (id) {
      fetchProjectAndOutline();
    } else {
      setLoading(false);
      showError('项目ID无效');
      navigate('/dashboard');
    }
  }, [id]);

  // 返回仪表盘
  const handleBackToDashboard = () => {
    navigate('/dashboard');
  };

  // 保存大纲
  const handleSaveOutline = async () => {
    if (!outline.trim()) {
      showError('大纲内容不能为空');
      return;
    }

    setSaving(true);
    try {
      const res = await API.post(`/api/outlines/${id}`, { content: outline });
      const { success, message } = res.data;
      if (success) {
        showSuccess('大纲保存成功');
        fetchProjectAndOutline(); // 刷新版本历史
        
        // 更新项目最后编辑时间
        if (project) {
          setProject({
            ...project,
            last_edited_at: new Date().toISOString()
          });
        }
      } else {
        showError(message || '保存失败');
      }
    } catch (error) {
      console.error('保存失败', error);
      showError(error.message || '保存失败，请稍后重试');
    } finally {
      setSaving(false);
    }
  };

  // 处理AI续写
  const handleGenerateContent = async () => {
    if (!outline.trim()) {
      showError('请先输入或上传大纲内容');
      return;
    }

    setGenerating(true);
    try {
      const res = await API.post(`/api/ai/generate/${id}`, {
        content: outline,
        style: aiSettings.style,
        wordLimit: aiSettings.wordLimit
      });
      
      const { success, message, data } = res.data;
      if (success) {
        setGeneratedContent(data.content);
        showSuccess('大纲续写完成');
      } else {
        showError(message || 'AI生成失败');
      }
    } catch (error) {
      console.error('AI生成失败', error);
      showError(error.message || 'AI生成失败，请稍后重试');
    } finally {
      setGenerating(false);
    }
  };

  // 采用AI生成的内容
  const handleAdoptGenerated = () => {
    if (!generatedContent) {
      showError('没有生成的内容可采用');
      return;
    }
    
    // 将生成的内容追加到当前大纲
    const updatedOutline = outline + '\n\n' + generatedContent;
    setOutline(updatedOutline);
    setGeneratedContent('');
    
    showSuccess('已采用AI续写内容');
  };

  // 处理文件选择
  const handleFileChange = (e) => {
    setUploadFile(e.target.files[0]);
  };

  // 上传大纲文件
  const handleFileUpload = async () => {
    if (!uploadFile) {
      showError('请先选择文件');
      return;
    }

    const formData = new FormData();
    formData.append('file', uploadFile);

    try {
      const res = await API.post(`/api/upload/outline/${id}`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });
      
      const { success, message, data } = res.data;
      if (success) {
        setOutline(data.content);
        showSuccess('文件上传成功');
      } else {
        showError(message || '上传失败');
      }
    } catch (error) {
      console.error('处理上传失败', error);
      showError(error.message || '处理上传失败');
    }
  };

  // 查看版本历史
  const handleViewVersionHistory = () => {
    setShowVersionHistory(true);
  };

  // 选择历史版本
  const handleSelectVersion = (version) => {
    setSelectedVersion(version);
    setOutline(version.content);
    setShowVersionHistory(false);
    showSuccess(`已恢复到版本 ${version.version_number}`);
  };

  // 导出大纲
  const handleExportOutline = async () => {
    try {
      const res = await API.post(`/api/exports/${id}`, {
        format: 'txt'
      });
      
      const { success, message, data } = res.data;
      if (success && data.file_url) {
        // 创建下载链接
        const link = document.createElement('a');
        link.href = data.file_url;
        link.download = `${project.title || '大纲'}.txt`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      } else {
        showError(message || '导出失败');
      }
    } catch (error) {
      console.error('导出失败', error);
      showError(error.message || '导出失败，请稍后重试');
    }
  };

  const styleOptions = [
    { key: 'default', text: '默认', value: 'default' },
    { key: 'fantasy', text: '玄幻奇幻', value: 'fantasy' },
    { key: 'scifi', text: '科幻', value: 'scifi' },
    { key: 'urban', text: '都市', value: 'urban' },
    { key: 'xianxia', text: '仙侠修真', value: 'xianxia' },
    { key: 'history', text: '历史军事', value: 'history' }
  ];

  const panes = [
    {
      menuItem: '生成设置',
      render: () => (
        <Tab.Pane>
          <Form>
            <Form.Field>
              <label>续写风格</label>
              <Dropdown
                fluid
                selection
                options={styleOptions}
                value={aiSettings.style}
                onChange={(e, { value }) => setAiSettings({...aiSettings, style: value})}
              />
            </Form.Field>
            
            <Form.Field>
              <label>字数限制</label>
              <Input
                type="number"
                min={100}
                max={5000}
                value={aiSettings.wordLimit}
                onChange={(e, { value }) => setAiSettings({...aiSettings, wordLimit: parseInt(value)})}
                fluid
              />
            </Form.Field>
            
            <Button
              primary
              fluid
              onClick={handleGenerateContent}
              loading={generating}
              disabled={generating}
            >
              <Icon name='send' /> 开始续写
            </Button>
          </Form>
        </Tab.Pane>
      )
    },
    {
      menuItem: '续写结果',
      render: () => (
        <Tab.Pane>
          {generating ? (
            <div className="generating-indicator">
              <Loader active inline="centered" />
              <p style={{ marginTop: 16, textAlign: 'center' }}>AI正在续写中，请稍候...</p>
            </div>
          ) : generatedContent ? (
            <div className="generated-content">
              <Form>
                <TextArea
                  value={generatedContent}
                  readOnly
                  style={{ minHeight: 250 }}
                />
              </Form>
              <Divider />
              <Button
                primary
                fluid
                onClick={handleAdoptGenerated}
              >
                <Icon name='plus' /> 采用此内容
              </Button>
            </div>
          ) : (
            <Message info>
              <Message.Header>尚未生成续写内容</Message.Header>
              <p>请在"生成设置"选项卡中设置参数并点击"开始续写"</p>
            </Message>
          )}
        </Tab.Pane>
      )
    }
  ];

  if (loading) {
    return (
      <Dimmer active>
        <Loader size="large">加载中...</Loader>
      </Dimmer>
    );
  }

  return (
    <Container fluid style={{ padding: '2em' }}>
      <Grid>
        <Grid.Row>
          <Grid.Column width={16}>
            <Segment clearing>
              <Header as='h2' floated='left'>
                <Button 
                  icon 
                  labelPosition='left'
                  onClick={handleBackToDashboard}
                >
                  <Icon name='arrow left' />
                  返回
                </Button>
                {project ? project.title : '编辑项目'}
              </Header>
              <Button.Group floated='right'>
                <Button onClick={() => document.getElementById('fileInput').click()}>
                  <Icon name='upload' /> 上传
                </Button>
                <input
                  id="fileInput"
                  type="file"
                  accept=".txt,.docx"
                  style={{ display: 'none' }}
                  onChange={handleFileChange}
                />
                {uploadFile && (
                  <Button positive onClick={handleFileUpload}>
                    <Icon name='check' /> 确认上传
                  </Button>
                )}
                <Button onClick={handleExportOutline}>
                  <Icon name='download' /> 导出
                </Button>
                <Button onClick={handleViewVersionHistory}>
                  <Icon name='history' /> 历史
                </Button>
                <Button primary onClick={handleSaveOutline}>
                  <Icon name='save' /> 保存
                </Button>
              </Button.Group>
            </Segment>
          </Grid.Column>
        </Grid.Row>

        <Grid.Row>
          <Grid.Column width={10}>
            <Segment>
              <Header as='h3'>大纲编辑</Header>
              <Form>
                <TextArea
                  placeholder="在此输入或上传您的网文大纲..."
                  value={outline}
                  onChange={(e, { value }) => setOutline(value)}
                  style={{ minHeight: 500 }}
                />
              </Form>
            </Segment>
          </Grid.Column>
          
          <Grid.Column width={6}>
            <Segment>
              <Header as='h3'>AI续写</Header>
              <Tab panes={panes} />
            </Segment>
          </Grid.Column>
        </Grid.Row>
      </Grid>

      {/* 历史版本弹窗 */}
      <Modal open={showVersionHistory} onClose={() => setShowVersionHistory(false)} size="large">
        <Modal.Header>历史版本</Modal.Header>
        <Modal.Content scrolling>
          {versions.length > 0 ? (
            <Card.Group>
              {versions.map((version) => (
                <Card fluid key={version.id} onClick={() => handleSelectVersion(version)}>
                  <Card.Content>
                    <Card.Header>
                      版本 {version.version_number}
                      {version.is_ai_generated && (
                        <span style={{ marginLeft: '1em', fontSize: '0.8em', color: 'grey' }}>
                          (AI生成)
                        </span>
                      )}
                    </Card.Header>
                    <Card.Meta>
                      {new Date(version.created_at).toLocaleString()}
                    </Card.Meta>
                    <Card.Description>
                      {version.content.substring(0, 100)}
                      {version.content.length > 100 ? '...' : ''}
                    </Card.Description>
                  </Card.Content>
                  <Card.Content extra>
                    <Button basic color="blue" size="small">
                      <Icon name='undo' /> 恢复此版本
                    </Button>
                  </Card.Content>
                </Card>
              ))}
            </Card.Group>
          ) : (
            <Message info>
              <Message.Header>暂无历史版本</Message.Header>
              <p>保存大纲后将在此显示历史版本</p>
            </Message>
          )}
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setShowVersionHistory(false)}>
            关闭
          </Button>
        </Modal.Actions>
      </Modal>

      {/* AI续写模态框 */}
      <Modal
        open={aiModalOpen}
        onClose={() => setAiModalOpen(false)}
      >
        <Modal.Header>AI续写设置</Modal.Header>
        <Modal.Content>
          <Form>
            <Form.Select
              label='续写风格'
              name='style'
              options={[
                { key: 'default', text: '默认', value: 'default' },
                { key: 'fantasy', text: '玄幻', value: 'fantasy' },
                { key: 'scifi', text: '科幻', value: 'scifi' },
                { key: 'urban', text: '都市', value: 'urban' },
                { key: 'xianxia', text: '仙侠', value: 'xianxia' },
                { key: 'history', text: '历史', value: 'history' }
              ]}
              value={aiSettings.style}
              onChange={(e, { value }) => setAiSettings({ ...aiSettings, style: value })}
            />
            <Form.Input
              label='生成字数'
              name='wordLimit'
              type='number'
              value={aiSettings.wordLimit}
              onChange={(e, { value }) => setAiSettings({ ...aiSettings, wordLimit: value })}
              min='100'
              max='5000'
            />
            <Message>
              <Message.Header>提示</Message.Header>
              <p>AI续写将基于您当前的大纲内容生成后续文字，生成的内容将追加到现有内容之后。</p>
            </Message>
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setAiModalOpen(false)}>
            取消
          </Button>
          <Button 
            primary 
            onClick={handleGenerateContent}
            loading={generating}
            disabled={generating}
          >
            开始生成
          </Button>
        </Modal.Actions>
      </Modal>
    </Container>
  );
};

export default Editor; 