import React, { useState, useEffect, useContext } from 'react';
import { 
  Layout, Button, Card, Typography, Tabs, Form, Input, 
  Select, InputNumber, Divider, Upload, Spin,
  Space, Tooltip, Modal
} from 'antd';
import { 
  LeftOutlined, SaveOutlined, FileAddOutlined,
  UploadOutlined, DownloadOutlined, HistoryOutlined,
  RollbackOutlined, SendOutlined
} from '@ant-design/icons';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';
import { UserContext } from '../context/User';
import '../styles/Editor.css';

const { Header, Content } = Layout;
const { Title, Text } = Typography;
const { TabPane } = Tabs;
const { TextArea } = Input;
const { Option } = Select;

const Editor = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [userState] = useContext(UserContext);
  
  const [project, setProject] = useState({});
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
    fetchProjectAndOutline();
  }, [id]);

  // 返回仪表盘
  const handleBackToDashboard = () => {
    navigate('/dashboard');
  };

  // 保存大纲
  const handleSaveOutline = async () => {
    try {
      const res = await API.post(`/api/outlines/${id}`, { content: outline });
      const { success, message } = res.data;
      if (success) {
        showSuccess('大纲保存成功');
        fetchProjectAndOutline(); // 刷新版本历史
      } else {
        showError(message || '保存失败');
      }
    } catch (error) {
      console.error('保存失败', error);
      showError(error.message || '保存失败，请稍后重试');
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

  // 上传大纲文件
  const handleFileUpload = async (info) => {
    if (info.file.status === 'done') {
      try {
        const res = info.file.response;
        if (res.success) {
          setOutline(res.data.content);
          showSuccess('文件上传成功');
        } else {
          showError(res.message || '上传失败');
        }
      } catch (error) {
        console.error('处理上传失败', error);
        showError(error.message || '处理上传失败');
      }
    } else if (info.file.status === 'error') {
      showError('文件上传失败');
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

  // 文件上传组件的配置
  const uploadProps = {
    name: 'file',
    action: `/api/upload/outline/${id}`,
    headers: {
      Authorization: userState.user ? `Bearer ${userState.user.token}` : ''
    },
    onChange: handleFileUpload,
    accept: '.txt,.docx'
  };

  if (loading) {
    return (
      <div className="loading-container">
        <Spin size="large" />
        <Text style={{ marginTop: 16 }}>加载中...</Text>
      </div>
    );
  }

  return (
    <Layout className="editor-layout">
      <Header className="editor-header">
        <div className="header-left">
          <Button 
            type="link" 
            icon={<LeftOutlined />} 
            onClick={handleBackToDashboard}
          >
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>
            {project.title || '未命名项目'}
          </Title>
        </div>
        <div className="header-right">
          <Space>
            <Upload {...uploadProps}>
              <Tooltip title="上传大纲文件">
                <Button icon={<UploadOutlined />}>上传</Button>
              </Tooltip>
            </Upload>
            <Tooltip title="导出大纲">
              <Button icon={<DownloadOutlined />} onClick={handleExportOutline}>
                导出
              </Button>
            </Tooltip>
            <Tooltip title="查看历史版本">
              <Button icon={<HistoryOutlined />} onClick={handleViewVersionHistory}>
                历史
              </Button>
            </Tooltip>
            <Tooltip title="保存大纲">
              <Button 
                type="primary" 
                icon={<SaveOutlined />} 
                onClick={handleSaveOutline}
              >
                保存
              </Button>
            </Tooltip>
          </Space>
        </div>
      </Header>
      
      <Content className="editor-content">
        <div className="editor-container">
          <Card className="outline-editor-card" title="大纲编辑">
            <TextArea
              className="outline-editor"
              value={outline}
              onChange={(e) => setOutline(e.target.value)}
              placeholder="在此输入或上传您的网文大纲..."
              autoSize={{ minRows: 20, maxRows: 30 }}
            />
          </Card>
          
          <Card className="ai-output-card" title="AI续写">
            <Tabs defaultActiveKey="generate">
              <TabPane tab="生成设置" key="generate">
                <Form layout="vertical">
                  <Form.Item label="续写风格">
                    <Select
                      value={aiSettings.style}
                      onChange={(value) => setAiSettings({...aiSettings, style: value})}
                    >
                      <Option value="default">默认</Option>
                      <Option value="fantasy">玄幻奇幻</Option>
                      <Option value="scifi">科幻</Option>
                      <Option value="urban">都市</Option>
                      <Option value="xianxia">仙侠修真</Option>
                      <Option value="history">历史军事</Option>
                    </Select>
                  </Form.Item>
                  
                  <Form.Item label="字数限制">
                    <InputNumber
                      min={100}
                      max={5000}
                      value={aiSettings.wordLimit}
                      onChange={(value) => setAiSettings({...aiSettings, wordLimit: value})}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                  
                  <Form.Item>
                    <Button
                      type="primary"
                      icon={<SendOutlined />}
                      onClick={handleGenerateContent}
                      loading={generating}
                      block
                    >
                      开始续写
                    </Button>
                  </Form.Item>
                </Form>
              </TabPane>
              
              <TabPane tab="续写结果" key="result">
                {generating ? (
                  <div className="generating-indicator">
                    <Spin />
                    <Text style={{ marginTop: 16 }}>AI正在续写中，请稍候...</Text>
                  </div>
                ) : generatedContent ? (
                  <div className="generated-content">
                    <TextArea
                      value={generatedContent}
                      readOnly
                      autoSize={{ minRows: 15, maxRows: 20 }}
                    />
                    <Divider />
                    <Button
                      type="primary"
                      icon={<FileAddOutlined />}
                      onClick={handleAdoptGenerated}
                      block
                    >
                      采用此内容
                    </Button>
                  </div>
                ) : (
                  <div className="no-result">
                    <Text type="secondary">
                      尚未生成续写内容，请在"生成设置"选项卡中设置参数并点击"开始续写"
                    </Text>
                  </div>
                )}
              </TabPane>
            </Tabs>
          </Card>
        </div>
      </Content>
      
      {/* 历史版本弹窗 */}
      <Modal
        title="历史版本"
        visible={showVersionHistory}
        onCancel={() => setShowVersionHistory(false)}
        footer={null}
        width={700}
      >
        {versions.length > 0 ? (
          <div className="version-list">
            {versions.map((version) => (
              <Card 
                key={version.id} 
                className="version-item"
                hoverable
                onClick={() => handleSelectVersion(version)}
              >
                <div className="version-header">
                  <div>
                    <Text strong>版本 {version.version_number}</Text>
                    {version.is_ai_generated && (
                      <Text type="secondary" style={{ marginLeft: 8 }}>
                        (AI生成)
                      </Text>
                    )}
                  </div>
                  <Text type="secondary">
                    {new Date(version.created_at).toLocaleString()}
                  </Text>
                </div>
                <div className="version-preview">
                  {version.content.substring(0, 100)}
                  {version.content.length > 100 ? '...' : ''}
                </div>
                <Button 
                  type="text" 
                  icon={<RollbackOutlined />} 
                  size="small"
                >
                  恢复此版本
                </Button>
              </Card>
            ))}
          </div>
        ) : (
          <div className="no-versions">
            <Text type="secondary">暂无历史版本</Text>
          </div>
        )}
      </Modal>
    </Layout>
  );
};

export default Editor; 