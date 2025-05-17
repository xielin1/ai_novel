import React, { useContext, useEffect, useState } from 'react';
import { Button, Container, Grid, Header, Image, Segment, Icon } from 'semantic-ui-react';
import { API, showError } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { useNavigate } from 'react-router-dom';
import './style.css';

const Home = () => {
  const [statusState] = useContext(StatusContext);
  const homePageLink = localStorage.getItem('home_page_link') || '';
  const [homePageContent, setHomePageContent] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  const displayNotice = async () => {
    const res = await API.get('/api/notice');
    const { success, message, data } = res.data;
    if (success) {
      let oldNotice = localStorage.getItem('notice');
      if (data !== oldNotice && data !== '') {
        showError(data);
        localStorage.setItem('notice', data);
      }
    } else {
      showError(message);
    }
  };

  const fetchHomePageContent = async () => {
    try {
      setLoading(true);
      const res = await API.get('/api/homepage');
      const { success, data } = res.data;
      if (success && data) {
        try {
          // 尝试解析JSON配置
          const contentObj = JSON.parse(data);
          setHomePageContent(contentObj);
        } catch (e) {
          // 如果解析失败，使用默认配置
          setHomePageContent(getDefaultConfig());
        }
      } else {
        setHomePageContent(getDefaultConfig());
      }
    } catch (error) {
      console.error('获取首页配置失败', error);
      setHomePageContent(getDefaultConfig());
    } finally {
      setLoading(false);
    }
  };

  // 默认配置
  const getDefaultConfig = () => {
    return {
      title: "AI网文大纲续写工具",
      subtitle: "让创作更轻松，让灵感不中断",
      description: "这是一个帮助作者快速生成、续写小说大纲的AI工具，告别写作瓶颈，让创作如行云流水。",
      backgroundImage: "https://source.unsplash.com/random/1600x900/?write,landscape",
      features: [
        {
          title: "智能续写",
          description: "基于前文内容智能分析，生成符合情节发展的续写内容",
          icon: "magic"
        },
        {
          title: "多种风格",
          description: "支持玄幻、科幻、都市等多种风格的续写，满足不同创作需求",
          icon: "paint brush"
        },
        {
          title: "版本管理",
          description: "保存多个版本的大纲，随时回溯，对比不同版本的内容",
          icon: "history"
        },
        {
          title: "导出功能",
          description: "一键导出大纲内容，支持多种格式，方便后续创作",
          icon: "file alternate"
        }
      ]
    };
  };

  useEffect(() => {
    displayNotice();
    fetchHomePageContent();
  }, []);

  if (loading) {
    return <div className="home-loading">加载中...</div>;
  }

  // 修改逻辑：使用后端设置的UseExternalHome
  const useExternalHome = statusState?.status?.use_external_home || false;
  if (homePageLink !== '' && useExternalHome) {
    return (
      <iframe
        src={homePageLink}
        style={{ width: '100%', height: '100vh', border: 'none' }}
        title="homepage"
      />
    );
  }

  // 处理按钮点击
  const handleWriteFragment = () => {
    navigate('/ai/prompt');
  };

  const handleWriteOutline = () => {
    navigate('/dashboard');
  };

  return (
    <div className="home-container">
      {/* 英雄区域 */}
      <Segment 
        className="hero-section" 
        style={{ 
          backgroundImage: `linear-gradient(rgba(0, 0, 0, 0.5), rgba(0, 0, 0, 0.7)), url(${homePageContent.backgroundImage})` 
        }}
        vertical
        textAlign='center'
        padded='very'
      >
        <Container text>
          <Header
            as='h1'
            content={homePageContent.title}
            className="hero-title"
          />
          <Header
            as='h2'
            content={homePageContent.subtitle}
            className="hero-subtitle"
          />
          <p className="hero-description">{homePageContent.description}</p>
          <div className="hero-buttons">
            <Button size='huge' primary onClick={handleWriteFragment}>
              <Icon name='pencil' />
              续写片段
            </Button>
            <Button size='huge' secondary onClick={handleWriteOutline}>
              <Icon name='file alternate outline' />
              续写大纲
            </Button>
          </div>
        </Container>
      </Segment>

      {/* 特性展示 */}
      <Segment vertical className="features-section" padded='very'>
        <Container>
          <Header as='h2' textAlign='center' className="section-title">
            强大功能
          </Header>
          <Grid stackable columns={4} className="features-grid">
            {homePageContent.features.map((feature, index) => (
              <Grid.Column key={index}>
                <div className="feature-item">
                  <Icon name={feature.icon} size='huge' className="feature-icon" />
                  <Header as='h3'>{feature.title}</Header>
                  <p>{feature.description}</p>
                </div>
              </Grid.Column>
            ))}
          </Grid>
        </Container>
      </Segment>

      {/* 使用流程 */}
      <Segment vertical className="how-it-works" padded='very'>
        <Container text>
          <Header as='h2' textAlign='center' className="section-title">
            如何使用
          </Header>
          <div className="steps">
            <div className="step">
              <div className="step-number">1</div>
              <div className="step-content">
                <h3>创建项目</h3>
                <p>在个人仪表盘中创建新项目，设置项目类型和基本信息</p>
              </div>
            </div>
            <div className="step">
              <div className="step-number">2</div>
              <div className="step-content">
                <h3>编写初始内容</h3>
                <p>在编辑器中输入初始大纲或片段内容</p>
              </div>
            </div>
            <div className="step">
              <div className="step-number">3</div>
              <div className="step-content">
                <h3>AI智能续写</h3>
                <p>选择合适的续写风格，点击续写按钮，AI自动生成后续内容</p>
              </div>
            </div>
            <div className="step">
              <div className="step-number">4</div>
              <div className="step-content">
                <h3>修改与完善</h3>
                <p>根据需要编辑AI生成的内容，保存或导出最终成果</p>
              </div>
            </div>
          </div>
        </Container>
      </Segment>

      {/* 号召性用语 */}
      <Segment vertical className="cta-section" padded='very'>
        <Container text textAlign='center'>
          <Header as='h2' className="cta-title">
            开始您的创作之旅
          </Header>
          <p className="cta-description">
            告别创作瓶颈，让AI成为您的得力助手，现在就开始使用吧！
          </p>
          <div className="cta-buttons">
            <Button size='huge' primary onClick={handleWriteFragment}>
              <Icon name='pencil' />
              续写片段
            </Button>
            <Button size='huge' secondary onClick={handleWriteOutline}>
              <Icon name='file alternate outline' />
              续写大纲
            </Button>
          </div>
        </Container>
      </Segment>
    </div>
  );
};

export default Home;
