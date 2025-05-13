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
  Card,
  Label,
  Popup,
  Transition,
  Image
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
  const [tooltipVisible, setTooltipVisible] = useState(false);

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
      } else {
        console.log(outlineRes.data);
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
        
        // 显示保存提示
        setTooltipVisible(true);
        setTimeout(() => setTooltipVisible(false), 2000);
        
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
        setUploadFile(null); // 重置上传文件状态
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

  // 格式化日期显示
  const formatDate = dateString => {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const styleOptions = [
    { key: 'default', text: '默认', value: 'default' },
    { key: 'fantasy', text: '玄幻奇幻', value: 'fantasy', color: 'purple' },
    { key: 'scifi', text: '科幻', value: 'scifi', color: 'blue' },
    { key: 'urban', text: '都市', value: 'urban', color: 'teal' },
    { key: 'xianxia', text: '仙侠修真', value: 'xianxia', color: 'green' },
    { key: 'history', text: '历史军事', value: 'history', color: 'brown' }
  ];

  const panes = [
    {
      menuItem: { key: 'settings', icon: 'settings', content: '生成设置' },
      render: () => (
        <Tab.Pane className="ai-tab-content">
          <Form>
            <Form.Field>
              <label>续写风格</label>
              <Dropdown
                fluid
                selection
                options={styleOptions.map(option => ({
                  ...option,
                  text: <span>
                    {option.color && <Label circular empty color={option.color} style={{marginRight: '8px'}} />}
                    {option.text}
                  </span>
                }))}
                value={aiSettings.style}
                onChange={(e, { value }) => setAiSettings({...aiSettings, style: value})}
              />
            </Form.Field>
            
            <Form.Field>
              <label>字数限制 ({aiSettings.wordLimit}字)</label>
              <Input
                type="range"
                min={100}
                max={5000}
                step={100}
                value={aiSettings.wordLimit}
                onChange={(e, { value }) => setAiSettings({...aiSettings, wordLimit: parseInt(value)})}
                fluid
              />
              <div style={{display: 'flex', justifyContent: 'space-between', fontSize: '12px', color: '#666', marginTop: '5px'}}>
                <span>100字</span>
                <span>5000字</span>
              </div>
            </Form.Field>
            
            <Button
              color="teal"
              fluid
              onClick={handleGenerateContent}
              loading={generating}
              disabled={generating}
              style={{marginTop: '20px', borderRadius: '4px'}}
            >
              <Icon name='magic' /> 开始续写
            </Button>
          </Form>
        </Tab.Pane>
      )
    },
    {
      menuItem: { key: 'result', icon: 'file text', content: '续写结果' },
      render: () => (
        <Tab.Pane className="ai-tab-content">
          {generating ? (
            <div className="generating-indicator">
              <Loader active inline="centered" />
              <div style={{ marginTop: '20px', textAlign: 'center' }}>
                <p style={{color: '#666'}}>AI正在续写中，请稍候...</p>
                <p style={{fontSize: '12px', color: '#999', marginTop: '10px'}}>根据内容长度，这可能需要几秒钟时间</p>
              </div>
            </div>
          ) : generatedContent ? (
            <div className="generated-content">
              <Segment raised style={{background: '#f9f9f9', borderRadius: '8px'}}>
                <Label attached='top' color='teal'>AI 续写结果</Label>
                <div style={{padding: '10px', marginTop: '10px', maxHeight: '300px', overflowY: 'auto', lineHeight: '1.6'}}>
                  {generatedContent.split('\n').map((line, i) => (
                    <p key={i}>{line || <br/>}</p>
                  ))}
                </div>
              </Segment>
              <Button
                positive
                fluid
                onClick={handleAdoptGenerated}
                style={{marginTop: '15px', borderRadius: '4px'}}
              >
                <Icon name='plus' /> 采用此内容
              </Button>
            </div>
          ) : (
            <Message info icon>
              <Icon name='info circle' />
              <Message.Content>
                <Message.Header>尚未生成续写内容</Message.Header>
                <p>请在"生成设置"选项卡中设置参数并点击"开始续写"</p>
              </Message.Content>
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
    <Container fluid style={{ padding: '1.5em' }}>
      <Segment raised style={{ borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.1)', padding: '0' }}>
        {/* 顶部导航条 */}
        <div style={{
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center', 
          padding: '15px 20px',
          borderBottom: '1px solid #f0f0f0',
          background: '#fcfcfc',
          borderRadius: '8px 8px 0 0'
        }}>
          <div style={{display: 'flex', alignItems: 'center'}}>
            <Button 
              icon 
              basic
              size="small"
              onClick={handleBackToDashboard}
              style={{marginRight: '12px', boxShadow: 'none'}}
            >
              <Icon name='arrow left' />
            </Button>
            <div>
              <Header as='h3' style={{margin: '0'}}>
                {project ? project.title : '编辑项目'}
                {project && project.genre && (
                  <Label size='tiny' color='blue' style={{marginLeft: '8px', borderRadius: '20px', fontSize: '10px'}}>
                    {project.genre}
                  </Label>
                )}
              </Header>
              <div style={{fontSize: '12px', color: '#888', marginTop: '4px'}}>
                {project && project.last_edited_at ? `上次编辑于 ${formatDate(project.last_edited_at)}` : '新项目'}
              </div>
            </div>
          </div>
          
          <div>
            <Popup
              open={tooltipVisible}
              content='大纲已保存'
              position='bottom center'
              inverted
              style={{opacity: 0.9}}
              trigger={
                <Button 
                  primary 
                  onClick={handleSaveOutline} 
                  loading={saving}
                  disabled={saving}
                  style={{borderRadius: '4px', marginRight: '5px'}}
                >
                  <Icon name='save' /> 保存
                </Button>
              }
            />
            <Dropdown
              trigger={
                <Button basic icon style={{boxShadow: 'none'}}>
                  <Icon name='ellipsis vertical' />
                </Button>
              }
              direction='left'
              icon={null}
            >
              <Dropdown.Menu>
                <Dropdown.Item icon='upload' text='上传文件' onClick={() => document.getElementById('fileInput').click()} />
                <input
                  id="fileInput"
                  type="file"
                  accept=".txt,.docx"
                  style={{ display: 'none' }}
                  onChange={handleFileChange}
                />
                {uploadFile && (
                  <Dropdown.Item icon='check' text={`上传 "${uploadFile.name}"`} onClick={handleFileUpload} />
                )}
                <Dropdown.Item icon='download' text='导出文档' onClick={handleExportOutline} />
                <Dropdown.Item icon='history' text='历史版本' onClick={handleViewVersionHistory} />
              </Dropdown.Menu>
            </Dropdown>
          </div>
        </div>

        {/* 主要内容区域 */}
        <div style={{padding: '20px'}}>
          <Grid stackable>
            <Grid.Row>
              <Grid.Column width={10}>
                <Segment basic style={{padding: '0', height: '100%'}}>
                  <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px'}}>
                    <Header as='h4' style={{margin: '0'}}>
                      <Icon name='edit' style={{color: '#2185d0'}} />
                      <Header.Content>
                        大纲编辑
                        <Header.Subheader>在这里编写或编辑您的作品大纲</Header.Subheader>
                      </Header.Content>
                    </Header>
                    <div style={{fontSize: '12px', color: versions.length > 0 ? '#888' : '#ccc', cursor: versions.length > 0 ? 'pointer' : 'default'}} onClick={versions.length > 0 ? handleViewVersionHistory : undefined}>
                      <Icon name='history' /> {versions.length} 个历史版本
                    </div>
                  </div>
                  <Form style={{height: 'calc(100% - 40px)'}}>
                    <TextArea
                      placeholder="在此开始编写您的网文大纲..."
                      value={outline}
                      onChange={(e, { value }) => setOutline(value)}
                      style={{ 
                        minHeight: '70vh', 
                        padding: '15px',
                        fontSize: '15px',
                        lineHeight: '1.6',
                        border: '1px solid #e0e0e0',
                        borderRadius: '6px',
                        boxShadow: 'inset 0 1px 3px rgba(0,0,0,0.05)'
                      }}
                    />
                  </Form>
                </Segment>
              </Grid.Column>
              
              <Grid.Column width={6}>
                <Segment raised style={{borderRadius: '8px', boxShadow: '0 2px 6px rgba(0,0,0,0.08)'}}>
                  <Header as='h4' style={{display: 'flex', alignItems: 'center'}}>
                    <Icon name='magic' style={{color: '#00b5ad'}} />
                    <Header.Content>
                      AI续写助手
                      <Header.Subheader>使用人工智能帮助您续写内容</Header.Subheader>
                    </Header.Content>
                  </Header>
                  <Tab 
                    menu={{ secondary: true, pointing: true, attached: 'top', tabular: true, style: {borderBottom: '1px solid #f0f0f0', paddingBottom: '0'} }} 
                    panes={panes} 
                    className="editor-tabs"
                  />
                </Segment>
              </Grid.Column>
            </Grid.Row>
          </Grid>
        </div>
      </Segment>

      {/* 历史版本弹窗 */}
      <Modal 
        open={showVersionHistory} 
        onClose={() => setShowVersionHistory(false)} 
        size="large"
        style={{borderRadius: '8px'}}
      >
        <Modal.Header style={{borderBottom: '1px solid #f0f0f0', display: 'flex', alignItems: 'center'}}>
          <Icon name='history' style={{marginRight: '10px'}} />历史版本
        </Modal.Header>
        <Modal.Content scrolling style={{padding: '0'}}>
          {versions.length > 0 ? (
            <div style={{padding: '10px'}}>
              <Card.Group>
                {versions.map((version) => (
                  <Card 
                    fluid 
                    key={version.id} 
                    onClick={() => handleSelectVersion(version)}
                    style={{
                      cursor: 'pointer', 
                      borderRadius: '6px',
                      boxShadow: '0 2px 4px rgba(0,0,0,0.05)',
                      transition: 'transform 0.2s, box-shadow 0.2s'
                    }}
                    className="version-card"
                  >
                    <Card.Content>
                      <Card.Header style={{display: 'flex', alignItems: 'center'}}>
                        版本 {version.version_number}
                        {version.is_ai_generated && (
                          <Label color='teal' size='tiny' style={{marginLeft: '8px', borderRadius: '20px'}}>
                            <Icon name='magic' />AI生成
                          </Label>
                        )}
                      </Card.Header>
                      <Card.Meta style={{marginTop: '5px', fontSize: '12px'}}>
                        {formatDate(version.created_at)}
                        {version.is_ai_generated && version.ai_style && (
                          <span> · 风格：{styleOptions.find(opt => opt.value === version.ai_style)?.text || version.ai_style}</span>
                        )}
                      </Card.Meta>
                      <Card.Description style={{
                        marginTop: '10px', 
                        background: '#f9f9f9', 
                        padding: '10px', 
                        borderRadius: '4px', 
                        fontSize: '14px', 
                        lineHeight: '1.5',
                        color: '#666'
                      }}>
                        {version.content.substring(0, 150)}
                        {version.content.length > 150 ? '...' : ''}
                      </Card.Description>
                    </Card.Content>
                    <Card.Content extra style={{background: '#fcfcfc', borderTop: '1px solid #f0f0f0'}}>
                      <Button basic color="blue" size="small" fluid>
                        <Icon name='undo' /> 恢复此版本
                      </Button>
                    </Card.Content>
                  </Card>
                ))}
              </Card.Group>
            </div>
          ) : (
            <Message icon info style={{margin: '20px'}}>
              <Icon name='info circle' />
              <Message.Content>
                <Message.Header>暂无历史版本</Message.Header>
                <p>保存大纲后将在此显示历史版本</p>
              </Message.Content>
            </Message>
          )}
        </Modal.Content>
        <Modal.Actions style={{background: '#f9f9f9', padding: '15px', borderTop: '1px solid #f0f0f0'}}>
          <Button onClick={() => setShowVersionHistory(false)} style={{borderRadius: '4px'}}>
            关闭
          </Button>
        </Modal.Actions>
      </Modal>

      {/* 添加样式 */}
      <style jsx="true" global="true">{`
        .version-card:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 8px rgba(0,0,0,0.1) !important;
        }
        
        .ai-tab-content {
          padding: 15px !important;
          min-height: 350px;
        }
        
        .editor-tabs .ui.pointing.secondary.menu {
          margin-left: -5px;
          margin-right: -5px;
        }
        
        .editor-tabs .ui.pointing.secondary.menu .item {
          margin: 0;
          padding: 10px 15px;
          border-bottom-width: 3px;
        }
        
        textarea {
          font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif !important;
          resize: none !important;
        }
      `}</style>
    </Container>
  );
};

export default Editor; 