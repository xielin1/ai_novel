import React, { useContext, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';

import {
  Button,
  Container,
  Dropdown,
  Icon,
  Menu,
  Segment,
} from 'semantic-ui-react';
import { API, getSystemName, isAdmin, isMobile, showSuccess } from '../helpers';
import '../index.css';

// Header Buttons
const headerButtons = [
  {
    name: '首页',
    to: '/',
    icon: 'home',
  },
  {
    name: '我的作品',
    to: '/dashboard',
    icon: 'book',
    requireLogin: true,
  },
  {
    name: '网文续写',
    to: '/ai/prompt',
    icon: 'magic',
  },
  {
    name: '文件',
    to: '/file',
    icon: 'file',
    admin: true,
  },
  {
    name: '用户',
    to: '/user',
    icon: 'user',
    admin: true,
  },
  {
    name: '设置',
    to: '/setting',
    icon: 'setting',
  },
  {
    name: '关于',
    to: '/about',
    icon: 'info circle',
  },
];

const Header = () => {
  const [userState, userDispatch] = useContext(UserContext);
  let navigate = useNavigate();

  const [showSidebar, setShowSidebar] = useState(false);
  const systemName = getSystemName();

  async function logout() {
    setShowSidebar(false);
    await API.get('/api/user/logout');
    showSuccess('注销成功!');
    userDispatch({ type: 'logout' });
    localStorage.removeItem('user');
    navigate('/login');
  }

  const toggleSidebar = () => {
    setShowSidebar(!showSidebar);
  };

  const renderButtons = (isMobile) => {
    return headerButtons.map((button) => {
      if (button.admin && !isAdmin()) return <></>;
      if (button.requireLogin && !userState.user) return <></>;
      
      if (isMobile) {
        return (
          <Menu.Item
            key={button.name}
            onClick={() => {
              navigate(button.to);
              setShowSidebar(false);
            }}
          >
            {button.name}
          </Menu.Item>
        );
      }
      return (
        <Menu.Item key={button.name} as={Link} to={button.to}>
          <Icon name={button.icon} />
          {button.name}
        </Menu.Item>
      );
    });
  };

  if (isMobile()) {
    return (
      <>
        <Menu
          borderless
          size='large'
          style={
            showSidebar
              ? {
                  borderBottom: 'none',
                  marginBottom: '0',
                  borderTop: 'none',
                  height: '51px',
                }
              : { borderTop: 'none', height: '52px' }
          }
        >
          <Container>
            <Menu.Item as={Link} to='/'>
              <img
                src='/logo.png'
                alt='logo'
                style={{ marginRight: '0.75em' }}
              />
              <div style={{ fontSize: '20px' }}>
                <b>{systemName}</b>
              </div>
            </Menu.Item>
            <Menu.Menu position='right'>
              <Menu.Item onClick={toggleSidebar}>
                <Icon name={showSidebar ? 'close' : 'sidebar'} />
              </Menu.Item>
            </Menu.Menu>
          </Container>
        </Menu>
        {showSidebar ? (
          <Segment style={{ marginTop: 0, borderTop: '0' }}>
            <Menu secondary vertical style={{ width: '100%', margin: 0 }}>
              {renderButtons(true)}
              <Menu.Item>
                {userState.user ? (
                  <>
                    <Button
                      onClick={() => {
                        setShowSidebar(false);
                        navigate('/profile');
                      }}
                      style={{ marginRight: '8px' }}
                    >
                      用户中心
                    </Button>
                    <Button onClick={logout}>注销</Button>
                  </>
                ) : (
                  <>
                    <Button
                      onClick={() => {
                        setShowSidebar(false);
                        navigate('/login');
                      }}
                    >
                      登录
                    </Button>
                    <Button
                      onClick={() => {
                        setShowSidebar(false);
                        navigate('/register');
                      }}
                      primary
                    >
                      注册
                    </Button>
                    <Button
                      onClick={() => {
                        setShowSidebar(false);
                        navigate('/login');
                      }}
                      color='teal'
                      style={{ marginTop: '8px' }}
                    >
                      <Icon name='pencil' /> 开始创作
                    </Button>
                  </>
                )}
              </Menu.Item>
            </Menu>
          </Segment>
        ) : (
          <></>
        )}
      </>
    );
  }

  return (
    <>
      <Menu borderless style={{ borderTop: 'none' }}>
        <Container>
          <Menu.Item as={Link} to='/' className={'hide-on-mobile'}>
            <img src='/logo.png' alt='logo' style={{ marginRight: '0.75em' }} />
            <div style={{ fontSize: '20px' }}>
              <b>{systemName}</b>
            </div>
          </Menu.Item>
          {renderButtons(false)}
          <Menu.Menu position='right'>
            {userState.user ? (
              <Dropdown
                text={userState.user.username}
                pointing
                className='link item'
              >
                <Dropdown.Menu>
                  <Dropdown.Item as={Link} to='/profile'>
                    <Icon name='user circle' />
                    用户中心
                  </Dropdown.Item>
                  <Dropdown.Item onClick={logout}>
                    <Icon name='sign-out' />
                    注销
                  </Dropdown.Item>
                </Dropdown.Menu>
              </Dropdown>
            ) : (
              <>
                <Menu.Item
                  name='登录'
                  as={Link}
                  to='/login'
                  className='btn btn-link'
                />
                <Menu.Item>
                  <Button as={Link} to='/register' primary>
                    注册
                  </Button>
                </Menu.Item>
                <Menu.Item>
                  <Button as={Link} to='/login' color='teal'>
                    <Icon name='pencil' /> 开始创作
                  </Button>
                </Menu.Item>
              </>
            )}
          </Menu.Menu>
        </Container>
      </Menu>
    </>
  );
};

export default Header;
