import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button, Container, Form, Header, Radio, Segment } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';

const Payment = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [packageId, setPackageId] = useState(null);
  const [packageInfo, setPackageInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState('alipay');

  useEffect(() => {
    // 从URL参数获取套餐ID
    const queryParams = new URLSearchParams(location.search);
    const pkgId = queryParams.get('package_id');
    if (pkgId) {
      setPackageId(Number(pkgId));
      fetchPackageInfo(Number(pkgId));
    } else {
      navigate('/profile');
    }
  }, [location]);

  const fetchPackageInfo = async (id) => {
    try {
      setLoading(true);
      const res = await API.get('/api/packages');
      const { success, data } = res.data;
      if (success) {
        const targetPackage = data.find(pkg => pkg.id === id);
        if (targetPackage) {
          setPackageInfo(targetPackage);
        } else {
          showError('未找到指定套餐');
          navigate('/profile');
        }
      }
    } catch (error) {
      console.error('获取套餐信息失败', error);
    } finally {
      setLoading(false);
    }
  };

  const handlePaymentMethodChange = (e, { value }) => {
    setPaymentMethod(value);
  };

  const handleSubmit = async () => {
    if (!packageId) return;

    try {
      setLoading(true);
      const res = await API.post('/api/packages/subscribe', {
        package_id: packageId,
        payment_method: paymentMethod
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess('订阅成功');
        navigate('/profile');
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('订阅失败', error);
    } finally {
      setLoading(false);
    }
  };

  if (!packageInfo) {
    return <Container text style={{ marginTop: '7em' }}>加载中...</Container>;
  }

  return (
    <Container text style={{ marginTop: '7em' }}>
      <Header as="h2">确认订阅</Header>
      <Segment>
        <Header as="h3">{packageInfo.name}</Header>
        <p><strong>价格:</strong> {packageInfo.price} 元/{packageInfo.duration === 'monthly' ? '月' : '永久'}</p>
        <p><strong>每月Token:</strong> {packageInfo.monthly_tokens}</p>
        <p><strong>描述:</strong> {packageInfo.description}</p>
        
        <Form>
          <Form.Field>
            <label>选择支付方式</label>
          </Form.Field>
          <Form.Field>
            <Radio
              label='支付宝'
              name='paymentMethod'
              value='alipay'
              checked={paymentMethod === 'alipay'}
              onChange={handlePaymentMethodChange}
            />
          </Form.Field>
          <Form.Field>
            <Radio
              label='微信支付'
              name='paymentMethod'
              value='wechat'
              checked={paymentMethod === 'wechat'}
              onChange={handlePaymentMethodChange}
            />
          </Form.Field>
          <Button 
            primary 
            onClick={handleSubmit} 
            loading={loading}
            disabled={loading}
          >
            确认支付
          </Button>
          <Button onClick={() => navigate('/profile')}>
            取消
          </Button>
        </Form>
      </Segment>
    </Container>
  );
};

export default Payment; 