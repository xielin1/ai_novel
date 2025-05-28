import React, { useEffect, useState, useContext } from 'react';
import { Button, Card, Container, Divider, Form, Grid, Header, Icon, Input, Label, Message, Segment, Statistic } from 'semantic-ui-react';
import { API, showError, showSuccess, copy } from '../helpers';
import { UserContext } from '../context/User';
import { useNavigate } from 'react-router-dom';

const UserProfile = () => {
  const [userState] = useContext(UserContext);
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [packageInfo, setPackageInfo] = useState(null);
  const [packages, setPackages] = useState([]);
  const [referralCode, setReferralCode] = useState('');
  const [referralStats, setReferralStats] = useState({ total_referred: 0, total_tokens_earned: 0 });
  const [inputReferralCode, setInputReferralCode] = useState('');
  const [submittingReferralCode, setSubmittingReferralCode] = useState(false);

  // 免费版套餐信息
  const freePackage = {
    id: 'free',
    name: '免费版',
    price: '0',
    duration: 'monthly',
    description: '基础功能免费体验',
    monthly_tokens: 500,
    features: ['基础AI续写功能', '每月500个免费Token', '社区支持']
  };

  useEffect(() => {
    fetchUserPackage();
    fetchPackages();
    fetchReferralCode();
  }, []);

  const fetchUserPackage = async () => {
    try {
      setLoading(true);
      const res = await API.get('/api/user/package');
      const { success, data } = res.data;
      if (success) {
        setPackageInfo(data);
      } else {
        // 如果用户没有套餐，设置为免费版
        setPackageInfo({
          package: freePackage,
          subscription_status: '有效',
          start_date: '注册日期',
          expiry_date: '永久有效',
          auto_renew: false
        });
      }
    } catch (error) {
      console.error('获取用户套餐信息失败', error);
      // 发生错误时也设置为免费版
      setPackageInfo({
        package: freePackage,
        subscription_status: '有效',
        start_date: '注册日期',
        expiry_date: '永久有效',
        auto_renew: false
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchPackages = async () => {
    try {
      const res = await API.get('/api/package/all');
      const { success, data } = res.data;
      if (success) {
        setPackages(data.packages);
      }
    } catch (error) {
      console.error('获取套餐列表失败', error);
    }
  };

  const fetchReferralCode = async () => {
    try {
      const res = await API.get('/api/user/referral-code');
      const { success, data } = res.data;
      if (success) {
        setReferralCode(data.referral_code);
        setReferralStats({
          total_referred: data.total_referred,
          total_tokens_earned: data.total_tokens_earned
        });
      }
    } catch (error) {
      console.error('获取推荐码失败', error);
    }
  };

  const handleCopyReferralCode = async () => {
    if (referralCode) {
      await copy(referralCode);
      showSuccess('推荐码已复制到剪贴板');
    }
  };

  const handleCopyReferralLink = async () => {
    if (referralCode) {
      const referralLink = `${window.location.origin}/register?ref=${referralCode}`;
      await copy(referralLink);
      showSuccess('推荐链接已复制到剪贴板');
    }
  };

  const handleUseReferralCode = async () => {
    if (!inputReferralCode) {
      showError('请输入推荐码');
      return;
    }
    
    try {
      setSubmittingReferralCode(true);
      const res = await API.post('/api/user/referral', {
        referralCode: inputReferralCode
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess('推荐码使用成功');
        setInputReferralCode('');
        // 刷新Token余额
        fetchUserPackage();
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('使用推荐码失败', error);
    } finally {
      setSubmittingReferralCode(false);
    }
  };

  const handleSubscribe = (packageId) => {
    navigate(`/payment?package_id=${packageId}`);
  };

  const renderPackageStatus = () => {
    if (loading) {
      return <Message info>正在加载套餐信息...</Message>;
    }

    if (!packageInfo) {
      return (
        <Message warning>
          <Message.Header>您还没有订阅套餐</Message.Header>
          <p>订阅套餐可获得更多的Token用于AI续写</p>
        </Message>
      );
    }

    const isFreePlan = packageInfo.package.id === 'free';

    return (
      <Segment>
        <Header as='h3'>当前套餐信息</Header>
        <Grid columns={2} stackable>
          <Grid.Column>
            <Statistic>
              <Statistic.Value>{packageInfo.package.name}</Statistic.Value>
              <Statistic.Label>套餐类型</Statistic.Label>
            </Statistic>
          </Grid.Column>
          <Grid.Column>
            <Statistic>
              <Statistic.Value>{packageInfo.package.monthly_tokens}</Statistic.Value>
              <Statistic.Label>每月Token</Statistic.Label>
            </Statistic>
          </Grid.Column>
        </Grid>
        <Divider />
        <p><strong>订阅状态:</strong> {packageInfo.subscription_status}</p>
        {!isFreePlan && (
          <>
            <p><strong>开始日期:</strong> {packageInfo.start_date}</p>
            <p><strong>到期日期:</strong> {packageInfo.expiry_date}</p>
            <p><strong>自动续费:</strong> {packageInfo.auto_renew ? '是' : '否'}</p>
            {packageInfo.auto_renew && (
              <Button color='yellow' onClick={() => cancelRenewal()} size='small'>
                取消自动续费
              </Button>
            )}
          </>
        )}
        {isFreePlan ? (
          <Message info>
            <p>您正在使用免费版套餐，升级到高级套餐可获得更多Token和高级功能！</p>
          </Message>
        ) : (
          <Button color='blue' size='small'>
            升级套餐
          </Button>
        )}
      </Segment>
    );
  };

  const cancelRenewal = async () => {
    try {
      const res = await API.post('/api/package/cancel-renewal');
      const { success, message } = res.data;
      if (success) {
        showSuccess('已取消自动续费');
        fetchUserPackage(); // 刷新套餐信息
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('取消自动续费失败', error);
    }
  };

  return (
    <Container>
      <Header as='h2' icon textAlign='center' style={{ marginTop: '2em' }}>
        <Icon name='user' circular />
        <Header.Content>用户中心</Header.Content>
      </Header>

      <Grid stackable columns={2}>
        <Grid.Column>
          {renderPackageStatus()}
          
          {(!packageInfo || packages.length > 0) && (
            <Segment>
              <Header as='h3'>
                {packageInfo && packageInfo.package.id !== 'free' ? '升级套餐' : '选择套餐'}
              </Header>
              <Card.Group>
                {packages.map(pkg => (
                  <Card key={pkg.id}>
                    <Card.Content>
                      <Card.Header>{pkg.name}</Card.Header>
                      <Card.Meta>{pkg.price} 元/{pkg.duration === 'monthly' ? '月' : '年'}</Card.Meta>
                      <Card.Description>
                        <p>{pkg.description}</p>
                        <p>每月可获得 {pkg.monthly_tokens} Token</p>
                      </Card.Description>
                    </Card.Content>
                    <Card.Content extra>
                      <div>
                        {pkg.features.map((feature, index) => (
                          <Label key={index} basic>
                            <Icon name='check' /> {feature}
                          </Label>
                        ))}
                      </div>
                    </Card.Content>
                    <Button 
                      attached='bottom' 
                      primary
                      onClick={() => handleSubscribe(pkg.id)}
                    >
                      {packageInfo && packageInfo.package.id !== 'free' ? '升级到此套餐' : '订阅此套餐'}
                    </Button>
                  </Card>
                ))}
              </Card.Group>
            </Segment>
          )}
        </Grid.Column>

        <Grid.Column>
          <Segment>
            <Header as='h3'>我的推荐码</Header>
            {referralCode ? (
              <>
                <Grid columns={2} stackable>
                  <Grid.Column>
                    <Statistic size='small'>
                      <Statistic.Value>{referralStats.total_referred}</Statistic.Value>
                      <Statistic.Label>推荐人数</Statistic.Label>
                    </Statistic>
                  </Grid.Column>
                  <Grid.Column>
                    <Statistic size='small'>
                      <Statistic.Value>{referralStats.total_tokens_earned}</Statistic.Value>
                      <Statistic.Label>获得Token</Statistic.Label>
                    </Statistic>
                  </Grid.Column>
                </Grid>
                <Divider />
                <p>分享您的推荐码给好友，每当有新用户使用您的推荐码注册，您将获得额外的Token奖励！</p>
                <Input
                  action={
                    <Button color='teal' onClick={handleCopyReferralCode}>
                      复制推荐码
                    </Button>
                  }
                  fluid
                  value={referralCode}
                  readOnly
                />
                <Button 
                  fluid 
                  style={{ marginTop: '1em' }} 
                  onClick={handleCopyReferralLink}
                >
                  复制推荐链接
                </Button>
              </>
            ) : (
              <Message>加载推荐码中...</Message>
            )}
          </Segment>

          <Segment>
            <Header as='h3'>使用推荐码</Header>
            <p>首次输入好友的推荐码，双方都可获得Token奖励！</p>
            <Form>
              <Form.Field>
                <Input
                  placeholder='请输入推荐码'
                  value={inputReferralCode}
                  onChange={(e) => setInputReferralCode(e.target.value)}
                  action={
                    <Button 
                      color='green' 
                      onClick={handleUseReferralCode}
                      loading={submittingReferralCode}
                      disabled={submittingReferralCode}
                    >
                      提交
                    </Button>
                  }
                />
              </Form.Field>
            </Form>
          </Segment>
        </Grid.Column>
      </Grid>
    </Container>
  );
};

export default UserProfile; 