import { useState, useEffect } from 'react';
import './App.css';
import LoginPage from './pages/LoginPage';
import AdminUsersPage from './pages/AdminUsersPage'; 
import {
  BrowserRouter as Router, 
  Routes,                  
  Route,
  Navigate,                
  Outlet,                  
  useNavigate,             
  useLocation              
} from 'react-router-dom';

function ProtectedRoute() {
  const isAuthenticated = !!localStorage.getItem('token');
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <Outlet />; 
}

function AppNavbar() {
  const navigate = useNavigate();
  const isAuthenticated = !!localStorage.getItem('token');

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    <nav style={{ padding: '10px', background: '#eee', marginBottom: '20px' }}>
      {isAuthenticated && (
        <button onClick={handleLogout} style={{ float: 'right' }}>Cerrar Sesión</button>
      )}
      {/* Aquí podrían ir otros enlaces de navegación si los hubiera */}
    </nav>
  );
}

function App() {
  return (
    <Router>
      <div id="App">
        <AppNavbar /> 
        <Routes> 
          <Route path="/login" element={<LoginPage />} />
          
          <Route element={<ProtectedRoute />}>
            <Route path="/admin/users" element={<AdminUsersPage />} />
            {/* Aquí podrías añadir más rutas de administrador en el futuro */}
          </Route>

          <Route 
            path="/"
            element={
              !!localStorage.getItem('token') ? (
                <Navigate to="/admin/users" replace />
              ) : (
                <Navigate to="/login" replace />
              )
            }
          />
          
          {/* Opcional: Ruta para páginas no encontradas */}
          {/* <Route path="*" element={<div>Página no encontrada</div>} /> */}
        </Routes>
      </div>
    </Router>
  );
}

export default App;
