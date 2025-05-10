import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import './LoginPage.css';
import './font.css'; // Importamos la fuente

// En Vite/React, importamos las imágenes directamente
// IMPORTANTE: Asegúrate de copiar manualmente las imágenes a estas ubicaciones si no están ahí
// El error de carga anterior sugiere que Vite no puede encontrar estos archivos en esas rutas.
// Si las imágenes están en otra ubicación, ajusta estas rutas
import logoImg from '../assets/images/LogoLogin.png';
import tablonImg from '../assets/images/TablonLogin.png';

function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const from = location.state?.from?.pathname || "/admin/users";

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      const data = await response.json();

      if (response.ok) {
        localStorage.setItem('token', data.token);
        navigate(from, { replace: true });
      } else {
        setError(data.error || `Error: ${response.status} - ${response.statusText}`);
      }
    } catch (err) {
      console.error('Login API call error:', err);
      setError('Error de conexión. Inténtalo de nuevo más tarde.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      {/* Fondo con imagen que cubre toda la pantalla */}
      <div className="background-image"></div>
      
      {/* Contenido principal */}
      <div className="login-content">
        {/* Título YaMerito (ahora posicionado entre el sombrero y el tablón) */}
        <h1 className="login-title">YaMerito</h1>
        
        <div className="login-board">
          <img 
            src={tablonImg} 
            alt="Tablón de login"
            className="login-board-image"
          />
          
          {/* Formulario superpuesto */}
          <form onSubmit={handleSubmit} className="login-form">
            <div className="username-field">
              <input
                type="text"
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Usuario"
                required
              />
            </div>
            
            <div className="password-field">
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Contraseña"
                required
              />
            </div>
            
            <button type="submit" disabled={loading} className="login-button">
              <span>{loading ? 'Ingresando...' : 'Iniciar sesión'}</span>
            </button>
          </form>
        </div>
        
        {error && <div className="error-message">{error}</div>}
      </div>
    </div>
  );
}

export default LoginPage;
