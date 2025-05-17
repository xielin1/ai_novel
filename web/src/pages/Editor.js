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
  Image,
  Progress
} from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../helpers';
import { UserContext } from '../context/User';
import '../styles/Editor.css';
import { useDropzone } from 'react-dropzone';

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
  const [saving, setSaving] = useState(false);
  const [aiModalOpen, setAiModalOpen] = useState(false);
  const [tooltipVisible, setTooltipVisible] = useState(false);
  const [promptModalOpen, setPromptModalOpen] = useState(false);
  const [customPrompt, setCustomPrompt] = useState('');
  const [showComparison, setShowComparison] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

  // 设置文件拖放区域
  const { getRootProps, getInputProps } = useDropzone({
    accept: {
      'text/plain': ['.txt'],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx']
    },
    maxFiles: 1,
    onDrop: (acceptedFiles) => {
      if (acceptedFiles.length > 0) {
        parseOutlineFile(acceptedFiles[0]);
      }
    }
  });

  // 文件解析处理函数
  const parseOutlineFile = async (file) => {
    if (!file) return;
    
    setUploading(true);
    const formData = new FormData();
    formData.append('file', file);
    
    try {
      const res = await API.post(`/api/outline/parse/${id}`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        },
        onUploadProgress: (e) => {
          const progress = Math.round((e.loaded / e.total) * 100);
          setUploadProgress(progress);
        }
      });
      
      const { success, message, data } = res.data;
      if (success) {
        setOutline(data.content);
        showSuccess('大纲文件解析成功');
      } else {
        showError(message || '文件解析失败');
      }
    } catch (error) {
      console.error('处理文件失败', error);
      showError(error.message || '处理文件失败');
    } finally {
      setUploading(false);
      setUploadProgress(0);
    }
  };

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
        wordLimit: aiSettings.wordLimit,
        customPrompt: customPrompt || undefined
      });
      
      const { success, message, data } = res.data;
      if (success) {
        setGeneratedContent(data.content);
        setShowComparison(true);
        showSuccess('大纲续写完成');
      } else {
        // 当API返回失败时，使用模拟数据
        console.warn('API请求失败，使用模拟数据：', message);
        const mockContent = generateMockContent(outline, aiSettings.style);
        setGeneratedContent(mockContent);
        setShowComparison(true);
        showSuccess('使用模拟数据完成续写（API请求失败）');
      }
    } catch (error) {
      console.error('AI生成失败', error);
      // 捕获到异常时，也使用模拟数据
      const mockContent = generateMockContent(outline, aiSettings.style);
      setGeneratedContent(mockContent);
      setShowComparison(true);
      showSuccess('使用模拟数据完成续写（调试模式）');
    } finally {
      setGenerating(false);
    }
  };

  // 生成模拟续写内容
  const generateMockContent = (originalContent, style) => {
    // 根据不同风格返回不同的模拟内容
    const mockContentByStyle = {
      default: `这是默认风格的模拟续写内容。

基于您的大纲，故事可以这样发展：
1. 主角遭遇意外，获得特殊能力
2. 在探索能力来源的过程中，结识重要伙伴
3. 发现隐藏的阴谋，决定挺身而出
4. 经历多次挫折和考验
5. 最终战胜反派，实现成长

这只是一个模拟生成的内容，用于调试界面功能。在实际使用中，AI会根据您的大纲生成更加贴合的续写内容。`,

      fantasy: `【玄幻奇幻风格模拟续写】

苍穹之上，紫气东来。主角觉醒了上古血脉，体内灵力涌动。

修炼功法《太虚龙经》第一层已然小成，但前路依旧漫长。神秘的藏经阁中，或许藏着血脉进化的秘密。

沧澜大陆五大宗门已经注意到主角身上散发的远古气息，暗中派人监视。特别是魔焰宗的宗主，更是将主角视为眼中钉。

接下来，主角需要：
1. 寻找上古秘境，提升修为
2. 结识志同道合的修行者，组建自己的势力
3. 揭开血脉的真相，了解自己的使命
4. 与黑暗势力展开对决

这是模拟生成的玄幻奇幻续写内容，仅用于调试。`,

      scifi: `【科幻风格模拟续写】

太空站Alpha-9发出了最后一条加密信息，随后通讯中断。

主角作为星际联盟特派调查员，驾驶最新型号的量子飞船前往调查。船载AI"星辰"提醒，太空站所在的星域最近出现了时空异常。

抵达目的地后，发现太空站处于一种"时间冻结"的状态，所有人员如同雕塑一般定格。站内量子计算机的屏幕上，闪烁着一行神秘代码。

深入调查发现的关键信息：
- 站内科学家在进行跨维度实验
- 有证据表明接触了未知智慧文明
- 实验日志中提到"虫洞稳定器"被激活

这只是模拟生成的科幻风格续写，用于界面调试。`,

      urban: `【都市风格模拟续写】

公司年会上，主角意外得罪了集团太子爷，次日就收到了降职通知。

正当一筹莫展之际，大学室友打来电话，邀请合伙创业。凭借专业特长，主角很快在新领域站稳脚跟。

与此同时，主角在社区志愿活动中结识了温柔善良的女医生，两人渐生情愫。不料，对方竟是当初那位太子爷的未婚妻。

商场如战场，面对前东家的打压和感情上的纠葛，主角需要：
1. 稳住创业团队，解决融资危机
2. 研发创新产品，打开市场
3. 理清复杂的感情关系
4. 找到事业与生活的平衡

这是模拟生成的都市文续写内容，仅用于调试功能。`,

      xianxia: `【仙侠修真风格模拟续写】

云海仙宗外，主角拜入掌门门下，成为记名弟子。修习《玉清心经》初见成效，体内已有三道灵脉觉醒。

在宗门秘境历练中，偶得一块残缺的古玉，内含神秘剑诀。练习时天降异象，引来宗门长老关注。

青灵峰上，主角邂逅了来自瑶池仙境的女修。她似乎对主角身世知晓内情，却欲言又止。

修真之路上的挑战：
1. 参加宗门大比，争取核心弟子身份
2. 探索古玉来历，领悟完整剑诀
3. 调查身世之谜，了解父母下落
4. 应对各方势力的暗中角逐

这是模拟生成的仙侠修真续写内容，用于界面功能调试。`,

      history: `【历史军事风格模拟续写】

乱世之中，群雄并起。主角出身将门，自幼习武，精通兵法。

北疆战报频传，边关告急。朝廷征调各路兵马，主角随父出征，初试锋芒。在一次伏击战中，力挽狂澜，救下大军。

回朝后，却发现朝中暗流涌动。奸臣当道，排挤异己。父亲遭受陷害，全族被贬边疆。

乱世之路：
1. 在边疆建立威信，招揽人才
2. 训练精锐部队，守卫边境
3. 寻找朝中盟友，为父洗刷冤屈
4. 应对敌国挑衅和内部势力争斗

风云际会，谁主沉浮？这是模拟生成的历史军事风格续写内容，仅用于调试。`
    };

    // 返回对应风格的模拟内容，如果没有匹配的风格则返回默认内容
    return mockContentByStyle[style] || mockContentByStyle.default;
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
    setShowComparison(false);
    
    showSuccess('已采用AI续写内容');
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

  // 打开提示词输入弹窗
  const handleOpenPromptModal = () => {
    setPromptModalOpen(true);
  };

  // 确认提示词输入
  const handleConfirmPrompt = () => {
    setPromptModalOpen(false);
    showSuccess('提示词已设置');
  };

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
                <Dropdown.Item icon='download' text='导出文档' onClick={handleExportOutline} />
                <Dropdown.Item icon='history' text='历史版本' onClick={handleViewVersionHistory} />
              </Dropdown.Menu>
            </Dropdown>
          </div>
        </div>

        {/* 主要内容区域 */}
        <div style={{padding: '20px'}}>
          {showComparison && generatedContent ? (
            <Grid stackable>
              <Grid.Row>
                <Grid.Column width={16}>
                  <Segment basic style={{padding: '0', marginBottom: '15px'}}>
                    <Header as='h4' style={{display: 'flex', alignItems: 'center', marginBottom: '15px'}}>
                      <Icon name='sync alternate' style={{color: '#2185d0'}} />
                      <Header.Content>
                        大纲对比视图
                        <Header.Subheader>对比原大纲和续写内容</Header.Subheader>
                      </Header.Content>
                    </Header>
                    <Button 
                      onClick={() => setShowComparison(false)} 
                      size='small'
                      style={{position: 'absolute', right: '0', top: '0'}}
                    >
                      <Icon name='close' /> 关闭对比
                    </Button>
                  </Segment>
                </Grid.Column>
              </Grid.Row>
              <Grid.Row>
                <Grid.Column width={8}>
                  <Segment raised style={{height: '100%'}}>
                    <Label attached='top' color='blue'>原始大纲</Label>
                    <div style={{
                      padding: '15px', 
                      marginTop: '10px', 
                      height: '50vh', 
                      overflowY: 'auto',
                      fontSize: '15px',
                      lineHeight: '1.6',
                      whiteSpace: 'pre-wrap'
                    }}>
                      {outline}
                    </div>
                  </Segment>
                </Grid.Column>
                <Grid.Column width={8}>
                  <Segment raised style={{height: '100%'}}>
                    <Label attached='top' color='teal'>续写内容</Label>
                    <div style={{
                      padding: '15px', 
                      marginTop: '10px', 
                      height: '50vh', 
                      overflowY: 'auto',
                      fontSize: '15px',
                      lineHeight: '1.6',
                      whiteSpace: 'pre-wrap',
                      background: '#f9f9f9'
                    }}>
                      {generatedContent}
                    </div>
                  </Segment>
                </Grid.Column>
              </Grid.Row>
              <Grid.Row>
                <Grid.Column width={16}>
                  <Button 
                    positive 
                    fluid
                    onClick={handleAdoptGenerated}
                    style={{borderRadius: '4px'}}
                  >
                    <Icon name='plus' /> 采用此续写内容
                  </Button>
                </Grid.Column>
              </Grid.Row>
            </Grid>
          ) : (
            <Grid stackable>
              <Grid.Row>
                <Grid.Column width={10}>
                  <Segment basic style={{padding: '0', height: '100%'}}>
                    {/* 文件上传区域 */}
                    <Segment
                      placeholder
                      {...getRootProps({ className: 'dropzone' })}
                      loading={uploading}
                      style={{ cursor: 'pointer', marginBottom: '15px' }}
                    >
                      <Header icon>
                        <Icon name='file outline' />
                        拖拽上传大纲文件或点击此处选择
                        <input {...getInputProps()} />
                      </Header>
                    </Segment>
                    
                    {uploading && (
                      <Progress
                        percent={uploadProgress}
                        success
                        progress='percent'
                        style={{marginTop: '-10px', marginBottom: '15px'}}
                      />
                    )}
                    
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
                    <div className="ai-tab-content" style={{marginTop: '15px'}}>
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
                          color="blue"
                          fluid
                          onClick={handleOpenPromptModal}
                          style={{marginTop: '15px', borderRadius: '4px'}}
                        >
                          <Icon name='comment' /> 自定义提示词
                        </Button>
                        
                        <Button
                          color="teal"
                          fluid
                          onClick={handleGenerateContent}
                          loading={generating}
                          disabled={generating}
                          style={{marginTop: '15px', borderRadius: '4px'}}
                        >
                          <Icon name='magic' /> 开始续写
                        </Button>
                        
                        {generating && (
                          <div style={{marginTop: '20px', textAlign: 'center'}}>
                            <Loader active inline="centered" size="small" />
                            <p style={{color: '#666', fontSize: '13px', marginTop: '10px'}}>AI正在续写中，请稍候...</p>
                          </div>
                        )}
                      </Form>
                    </div>
                  </Segment>
                </Grid.Column>
              </Grid.Row>
            </Grid>
          )}
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

      {/* 提示词输入弹窗 */}
      <Modal
        open={promptModalOpen}
        onClose={() => setPromptModalOpen(false)}
        size="small"
        style={{borderRadius: '8px'}}
      >
        <Modal.Header style={{borderBottom: '1px solid #f0f0f0', display: 'flex', alignItems: 'center'}}>
          <Icon name='comment' style={{marginRight: '10px'}} />自定义提示词
        </Modal.Header>
        <Modal.Content>
          <Form>
            <Form.Field>
              <label>输入自定义提示词：</label>
              <TextArea
                placeholder="例如：以幽默的风格续写，增加更多的对话..."
                value={customPrompt}
                onChange={(e, {value}) => setCustomPrompt(value)}
                style={{minHeight: '150px'}}
              />
            </Form.Field>
            <Message info>
              <Icon name='info circle' />
              <Message.Content>
                <p>提示词可以引导AI按照您期望的方向续写，例如指定特定的情节发展、人物特征或写作风格。</p>
              </Message.Content>
            </Message>
          </Form>
        </Modal.Content>
        <Modal.Actions style={{background: '#f9f9f9', padding: '15px', borderTop: '1px solid #f0f0f0'}}>
          <Button onClick={() => setPromptModalOpen(false)} style={{borderRadius: '4px'}}>
            取消
          </Button>
          <Button primary onClick={handleConfirmPrompt} style={{borderRadius: '4px'}}>
            确认
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