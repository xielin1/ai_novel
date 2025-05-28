import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button, Card, Container, Form, Header, Icon, Message, Segment, Statistic } from 'semantic-ui-react';
import { API, showError, showSuccess } from '../helpers';

const Payment = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [packageInfo, setPackageInfo] = useState(null);
  const [paymentMethod, setPaymentMethod] = useState('alipay');
  const [paymentUrl, setPaymentUrl] = useState('');

  useEffect(() => {
    const packageId = new URLSearchParams(location.search).get('package_id');
    if (packageId) {
      fetchPackageInfo(packageId);
    } else {
      showError('未选择套餐');
      navigate('/profile');
    }
  }, [location]);

  const fetchPackageInfo = async (packageId) => {
    try {
      setLoading(true);
      const res = await API.get(`/api/package/${packageId}`);
      const { success, data } = res.data;
      if (success) {
        setPackageInfo(data);
      } else {
        showError('获取套餐信息失败');
        navigate('/profile');
      }
    } catch (error) {
      console.error('获取套餐信息失败', error);
      showError('获取套餐信息失败');
      navigate('/profile');
    } finally {
      setLoading(false);
    }
  };

  const handlePayment = async () => {
    if (!packageInfo) return;

    try {
      setLoading(true);
      const res = await API.post('/api/payment/create', {
        package_id: packageInfo.id,
        payment_method: paymentMethod
      });

      const { success, data, message } = res.data;
      if (success) {
        setPaymentUrl(data.payment_url);
        // 如果是支付宝，直接跳转到支付页面
        if (paymentMethod === 'alipay') {
          window.location.href = data.payment_url;
        }
        showSuccess('订单创建成功');
      } else {
        showError(message || '创建订单失败');
      }
    } catch (error) {
      console.error('创建订单失败', error);
      showError('创建订单失败');
    } finally {
      setLoading(false);
    }
  };

  const renderPaymentInfo = () => {
    if (!packageInfo) return null;

    return (
      <Segment>
        <Header as='h3'>订单信息</Header>
        <Card fluid>
          <Card.Content>
            <Card.Header>{packageInfo.name}</Card.Header>
            <Card.Meta>
              {packageInfo.price} 元/{packageInfo.duration === 'monthly' ? '月' : '年'}
            </Card.Meta>
            <Card.Description>
              <p>{packageInfo.description}</p>
              <p>每月可获得 {packageInfo.monthly_tokens} Token</p>
            </Card.Description>
          </Card.Content>
        </Card>

        <Form>
          <Form.Field>
            <label>选择支付方式</label>
            <Form.Group>
              <Form.Radio
                label='支付宝'
                name='paymentMethod'
                value='alipay'
                checked={paymentMethod === 'alipay'}
                onChange={(e, { value }) => setPaymentMethod(value)}
              />
              <Form.Radio
                label='微信支付'
                name='paymentMethod'
                value='wechat'
                checked={paymentMethod === 'wechat'}
                onChange={(e, { value }) => setPaymentMethod(value)}
              />
            </Form.Group>
          </Form.Field>

          <Button
            primary
            fluid
            loading={loading}
            onClick={handlePayment}
          >
            立即支付
          </Button>
        </Form>

        {paymentMethod === 'wechat' && paymentUrl && (
          <Message info>
            <Message.Header>请使用微信扫码支付</Message.Header>
            <p>支付完成后，系统将自动更新您的套餐信息</p>
            <div style={{ textAlign: 'center', marginTop: '1em' }}>
              <img src={paymentUrl} alt="微信支付二维码" style={{ maxWidth: '200px' }} />
            </div>
          </Message>
        )}
      </Segment>
    );
  };

  return (
    <Container>
      <Header as='h2' icon textAlign='center' style={{ marginTop: '2em' }}>
        <Icon name='payment' circular />
        <Header.Content>套餐支付</Header.Content>
      </Header>

      {loading ? (
        <Message info>正在加载支付信息...</Message>
      ) : (
        renderPaymentInfo()
      )}
    </Container>
  );
};

export default Payment; 