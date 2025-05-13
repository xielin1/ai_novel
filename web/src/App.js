import React, { lazy, Suspense, useContext, useEffect, useState } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/lib/locale/zh_CN';
import Loading from './components/Loading';
import User from './pages/User';
import { PrivateRoute } from './components/PrivateRoute';
import RegisterForm from './components/RegisterForm';
import LoginForm from './components/LoginForm';
import NotFound from './pages/NotFound';
import Setting from './pages/Setting';
import EditUser from './pages/User/EditUser';
import AddUser from './pages/User/AddUser';
import { API, showError, showNotice } from './helpers';
import PasswordResetForm from './components/PasswordResetForm';
import GitHubOAuth from './components/GitHubOAuth';
import PasswordResetConfirm from './components/PasswordResetConfirm';
import { UserContext } from './context/User';
import { StatusContext } from './context/Status';
import File from './pages/File';
import AIPrompt from './pages/AIPrompt';
// 导入新添加的页面组件，移除Login组件
import Dashboard from './pages/Dashboard';
import Editor from './pages/Editor';
import './App.css';

const Home = lazy(() => import('./pages/Home'));
const About = lazy(() => import('./pages/About'));

// 自定义私有路由组件，用于新添加的页面
const CustomPrivateRoute = ({ children }) => {
  const isAuthenticated = localStorage.getItem('token') !== null;
  return isAuthenticated ? children : <Navigate to="/login" />;
};

function App() {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [loading, setLoading] = useState(false);

  const loadUser = () => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      userDispatch({ type: 'login', payload: data });
    }
  };
  
  const loadStatus = async () => {
    try {
      const res = await API.get('/api/status');
      const { success, data } = res.data;
      if (success) {

        localStorage.setItem('status', JSON.stringify(data));
        statusDispatch({ type: 'set', payload: data });
        localStorage.setItem('system_name', data.system_name);
        localStorage.setItem('footer_html', data.footer_html);
        localStorage.setItem('home_page_link', data.home_page_link);
        if (
          data.version !== process.env.REACT_APP_VERSION &&
          data.version !== 'v0.0.0' &&
          process.env.REACT_APP_VERSION !== ''
        ) {
          showNotice(
            `新版本可用：${data.version}，请使用快捷键 Shift + F5 刷新页面`
          );
        }
      } else {
        showError('无法正常连接至服务器！');
      }
    } catch (error) {
      console.error('加载状态失败', error);
    }
  };

  useEffect(() => {
    loadUser();
    loadStatus().then();
    
    // 模拟初始化加载（新功能）
    setLoading(true);
    const timer = setTimeout(() => {
      setLoading(false);
    }, 1000);

    return () => clearTimeout(timer);
  }, []);

  return (
    <ConfigProvider locale={zhCN}>
      <Routes>
        {/* 原有路由配置 */}
        <Route
          path='/'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <Home />
            </Suspense>
          }
        />
        <Route
          path='/file'
          element={
            <PrivateRoute>
              <File />
            </PrivateRoute>
          }
        />
        <Route
          path='/ai/prompt'
          element={
            <PrivateRoute>
              <AIPrompt />
            </PrivateRoute>
          }
        />
        <Route
          path='/user'
          element={
            <PrivateRoute>
              <User />
            </PrivateRoute>
          }
        />
        <Route
          path='/user/edit/:id'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <EditUser />
            </Suspense>
          }
        />
        <Route
          path='/user/edit'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <EditUser />
            </Suspense>
          }
        />
        <Route
          path='/user/add'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <AddUser />
            </Suspense>
          }
        />
        <Route
          path='/user/reset'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <PasswordResetConfirm />
            </Suspense>
          }
        />
        <Route
          path='/login'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <LoginForm />
            </Suspense>
          }
        />
        <Route
          path='/register'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <RegisterForm />
            </Suspense>
          }
        />
        <Route
          path='/reset'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <PasswordResetForm />
            </Suspense>
          }
        />
        <Route
          path='/oauth/github'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <GitHubOAuth />
            </Suspense>
          }
        />
        <Route
          path='/setting'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>}>
                <Setting />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/about'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <About />
            </Suspense>
          }
        />
        
        {/* 新添加的路由配置，使用原有的PrivateRoute组件 */}
        <Route 
          path="/dashboard" 
          element={
            <PrivateRoute>
              <Dashboard />
            </PrivateRoute>
          } 
        />
        <Route 
          path="/editor/:id" 
          element={
            <PrivateRoute>
              <Editor />
            </PrivateRoute>
          } 
        />
        
        {/* 捕获所有未匹配的路由 */}
        <Route path='*' element={<NotFound />} />
      </Routes>
    </ConfigProvider>
  );
}

export default App;
