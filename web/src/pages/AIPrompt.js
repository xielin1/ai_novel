import React, { useState, useEffect } from 'react';
import { Container, Button, Form, Header, Segment, Loader, Dropdown, Divider } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';
import { useNavigate } from 'react-router-dom';

const AIPrompt = () => {
  const [systemPrompt, setSystemPrompt] = useState('你是一个有用的AI助手');
  const [userPrompt, setUserPrompt] = useState('');
  const [response, setResponse] = useState('');
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState([]);
  const [selectedModel, setSelectedModel] = useState('qwen-turbo');
  const [temperature, setTemperature] = useState(0.7);
  const [maxTokens, setMaxTokens] = useState(2000);
  const [loadingModels, setLoadingModels] = useState(false);
  
  const navigate = useNavigate();

  useEffect(() => {
    // 获取可用模型列表
    const fetchModels = async () => {
      setLoadingModels(true);
      try {
        const res = await API.get('/api/ai/models');
        if (res && res.data) {
          // 对模型进行分类和排序
          const modelList = res.data.data || [];
          
          // 按模型类型对模型进行分组
          const modelGroups = {
            'Qwen基础模型': [],
            'Qwen2系列': [],
            'Qwen2.5系列': [],
            'Qwen专业模型': [],
            '其它模型': []
          };
          
          // 对模型进行分类
          modelList.forEach(model => {
            const modelId = model.id.toLowerCase();
            if (modelId.startsWith('qwen2.5-')) {
              modelGroups['Qwen2.5系列'].push(model);
            } else if (modelId.startsWith('qwen2-')) {
              modelGroups['Qwen2系列'].push(model);
            } else if (modelId.startsWith('qwen-') || modelId.startsWith('qwen1.5-')) {
              modelGroups['Qwen基础模型'].push(model);
            } else if (modelId.includes('coder') || modelId.includes('math') || modelId.includes('vl-')) {
              modelGroups['Qwen专业模型'].push(model);
            } else {
              modelGroups['其它模型'].push(model);
            }
          });
          
          // 创建下拉菜单选项
          const dropdownOptions = [];
          
          // 将分组添加到选项中
          Object.entries(modelGroups)
            .filter(([_, models]) => models.length > 0)
            .forEach(([group, models]) => {
              // 添加分组标题
              dropdownOptions.push({
                key: group,
                text: group,
                value: group,
                disabled: true
              });
              
              // 添加组内模型
              models.forEach(model => {
                dropdownOptions.push({
                  key: model.id,
                  text: `  ${model.id}`,  // 添加缩进以显示层级关系
                  value: model.id,
                  description: model.owned_by
                });
              });
            });
          
          setModels(dropdownOptions);
          
          // 如果没有选择模型或默认模型不在列表中，设置一个合适的默认模型
          if (!selectedModel || !modelList.some(m => m.id === selectedModel)) {
            // 优先选择 qwen-turbo
            const defaultModel = modelList.find(m => m.id === 'qwen-turbo') || 
                                modelList.find(m => m.id.includes('turbo')) ||
                                modelList[0];
            if (defaultModel) {
              setSelectedModel(defaultModel.id);
            }
          }
        }
      } catch (error) {
        if (error.response && error.response.status === 401) {
          // 如果未授权，重定向到登录页
          navigate('/login');
        } else {
          showError(error.message || '获取模型列表失败');
        }
      } finally {
        setLoadingModels(false);
      }
    };

    fetchModels();
  }, [navigate, selectedModel]);

  const handleSubmit = async () => {
    if (!userPrompt.trim()) {
      showError('请输入提示内容');
      return;
    }

    setLoading(true);
    setResponse('');

    try {
      const res = await API.post('/api/ai/prompt', {
        system_prompt: systemPrompt,
        user_prompt: userPrompt,
        model: selectedModel,
        temperature: parseFloat(temperature),
        max_tokens: parseInt(maxTokens)
      });

      if (res && res.data) {
        setResponse(res.data.content);
        showSuccess('生成成功，共消耗 ' + res.data.tokens_used + ' 个tokens');
      }
    } catch (error) {
      showError(error.response?.data?.error || error.message || '生成失败');
    } finally {
      setLoading(false);
    }
  };

  const handleModelChange = (e, { value }) => {
    // 确保不会选中分组标题
    if (value && !models.find(m => m.value === value && m.disabled)) {
      setSelectedModel(value);
    }
  };

  return (
    <Container>
      <Header as='h2' content='AI 提示生成' style={{ marginTop: '2em' }} />
      <Segment>
        <Header as='h3' content='系统提示（定义AI行为和身份）' />
        <Form>
          <Form.TextArea
            rows={2}
            value={systemPrompt}
            onChange={(e) => setSystemPrompt(e.target.value)}
            placeholder='输入系统提示，定义AI的行为或背景'
          />
          
          <Form.Group widths='equal'>
            <Form.Field>
              <label>选择模型</label>
              <Dropdown
                selection
                search
                options={models}
                value={selectedModel}
                onChange={handleModelChange}
                placeholder='选择模型'
                loading={loadingModels}
                disabled={loadingModels}
              />
            </Form.Field>
            
            <Form.Field>
              <label>温度 ({temperature})</label>
              <input
                type='range'
                min='0'
                max='2'
                step='0.1'
                value={temperature}
                onChange={(e) => setTemperature(e.target.value)}
              />
            </Form.Field>
            
            <Form.Field>
              <label>最大Tokens</label>
              <input
                type='number'
                value={maxTokens}
                onChange={(e) => setMaxTokens(e.target.value)}
                min='50'
                max='4000'
                step='50'
              />
            </Form.Field>
          </Form.Group>

          <Header as='h3' content='用户提示（你的问题）' />
          <Form.TextArea
            rows={4}
            value={userPrompt}
            onChange={(e) => setUserPrompt(e.target.value)}
            placeholder='在此输入你的问题或提示'
          />

          <Button 
            primary 
            onClick={handleSubmit} 
            disabled={loading || !userPrompt.trim()}
          >
            {loading ? <Loader active inline='centered' size='small' /> : '生成'}
          </Button>
        </Form>
      </Segment>

      {response && (
        <Segment>
          <Header as='h3' content='AI 响应' />
          <Divider />
          <div style={{ 
            whiteSpace: 'pre-wrap', 
            padding: '1em', 
            backgroundColor: '#f8f8f9', 
            borderRadius: '0.3em',
            maxHeight: '500px',
            overflow: 'auto'
          }}>
            {response}
          </div>
        </Segment>
      )}
    </Container>
  );
};

export default AIPrompt; 