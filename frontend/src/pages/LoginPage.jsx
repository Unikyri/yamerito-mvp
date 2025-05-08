import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom'; 
// Podríamos añadir un archivo CSS específico para LoginPage más adelante
// import './LoginPage.css'; 

function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate(); 
  const location = useLocation(); 

  // Determinar a dónde redirigir después del login
  const from = location.state?.from?.pathname || "/admin/users"; 

  const handleSubmit = async (event) => {
    event.preventDefault();
    setError(''); // Limpiar errores previos
    setLoading(true);

    console.log('Attempting login with:', { username, password });

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
        console.log('Login successful:', data);
        localStorage.setItem('token', data.token); // Guardar el token
        
        // Redirigir usando navigate
        navigate(from, { replace: true });

        // Ya no necesitamos el alert o limpiar campos aquí, la redirección se encarga
        // alert('¡Login exitoso! Token guardado. Redirección pendiente.'); 
        // setUsername(''); 
        // setPassword('');
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
    <div style={styles.container}>
      <div style={styles.loginBox}>
        <h2 style={styles.title}>Iniciar Sesión</h2>
        <form onSubmit={handleSubmit} style={styles.form}>
          <div style={styles.inputGroup}>
            <label htmlFor="username" style={styles.label}>Nombre de Usuario:</label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              style={styles.input}
            />
          </div>
          <div style={styles.inputGroup}>
            <label htmlFor="password" style={styles.label}>Contraseña:</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              style={styles.input}
            />
          </div>
          {error && <p style={styles.errorMessage}>{error}</p>}
          <button type="submit" disabled={loading} style={styles.button}>
            {loading ? 'Ingresando...' : 'Ingresar'}
          </button>
        </form>
      </div>
    </div>
  );
}

// Estilos básicos en línea para empezar. 
// Podríamos moverlos a un archivo CSS para una mejor organización.
const styles = {
  container: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh', // Que ocupe toda la altura de la vista
    backgroundColor: '#f0f2f5', // Un fondo suave
    fontFamily: 'Arial, sans-serif',
  },
  loginBox: {
    padding: '40px',
    backgroundColor: '#ffffff',
    borderRadius: '8px',
    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
    width: '100%',
    maxWidth: '400px',
    textAlign: 'center',
  },
  title: {
    marginBottom: '24px',
    color: '#333',
    fontSize: '24px',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
  },
  inputGroup: {
    marginBottom: '20px',
    textAlign: 'left',
  },
  label: {
    display: 'block',
    marginBottom: '8px',
    color: '#555',
    fontSize: '14px',
  },
  input: {
    width: '100%',
    padding: '10px',
    border: '1px solid #ddd',
    borderRadius: '4px',
    boxSizing: 'border-box', // Para que el padding no afecte el ancho total
    fontSize: '16px',
  },
  button: {
    padding: '12px',
    backgroundColor: '#007bff', // Azul primario
    color: 'white',
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '16px',
    transition: 'background-color 0.2s',
  },
  // Estilo para el botón cuando está desactivado o en hover (se puede agregar con :hover en CSS)
  // buttonDisabled: {
  //   backgroundColor: '#0056b3',
  // },
  errorMessage: {
    color: 'red',
    marginBottom: '16px',
    fontSize: '14px',
  }
};

export default LoginPage;
