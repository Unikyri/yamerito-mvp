import React, { useState, useEffect, useCallback } from 'react';
import UserFormModal from '../components/UserFormModal'; // Importar el modal

const AVAILABLE_ROLES = ['Admin', 'Employee']; // Roles disponibles en PascalCase

function AdminUsersPage() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState(null); // Estado para el usuario en edición

  const styles = {
    container: {
      padding: '30px',
      fontFamily: '"Segoe UI", Tahoma, Geneva, Verdana, sans-serif',
      backgroundColor: '#f8f9fa', // Light background for the page
      minHeight: '100vh'
    },
    headerContainer: {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '30px'
    },
    title: {
      color: '#343a40',
      fontSize: '2rem',
      fontWeight: '500'
    },
    loading: {
      color: '#007bff',
      textAlign: 'center',
      fontSize: '1.2rem',
      marginTop: '50px'
    },
    error: {
      color: '#dc3545',
      backgroundColor: '#f8d7da',
      border: '1px solid #f5c6cb',
      padding: '15px',
      borderRadius: '4px',
      marginBottom: '20px'
    },
    userTable: {
      width: '100%',
      borderCollapse: 'collapse',
      marginTop: '20px',
      backgroundColor: '#fff',
      boxShadow: '0 2px 10px rgba(0, 0, 0, 0.075)',
      borderRadius: '8px',
      overflow: 'hidden' // Ensures border-radius is respected by child elements like th/td
    },
    tableHeader: {
      borderBottom: '2px solid #dee2e6', // Thicker bottom border for header
      padding: '12px 15px',
      textAlign: 'left',
      backgroundColor: '#e9ecef', // Slightly darker header
      color: '#495057',
      fontWeight: '600',
      textTransform: 'uppercase',
      fontSize: '0.85em'
    },
    tableCell: {
      borderBottom: '1px solid #e9ecef', // Lighter border for rows
      padding: '12px 15px',
      color: '#495057',
      verticalAlign: 'middle' // Align content vertically
    },
    tableRow: {
      // For alternating row colors, you'd typically use CSS :nth-child(even)
      // This is harder with inline styles directly on map, but can be simulated if needed.
    },
    actionButton: {
        marginRight: '8px',
        padding: '6px 12px',
        fontSize: '0.9em',
        cursor: 'pointer',
        borderRadius: '4px',
        border: '1px solid transparent',
        transition: 'background-color 0.2s ease, border-color 0.2s ease'
    },
    editButton: {
        borderColor: '#ffc107',
        backgroundColor: 'transparent',
        color: '#ffc107',
    },
    deleteButton: {
        borderColor: '#dc3545',
        backgroundColor: 'transparent',
        color: '#dc3545',
    },
    // Hover styles would ideally be CSS classes: 
    // .editButton:hover { backgroundColor: '#ffc107', color: '#212529' }
    // .deleteButton:hover { backgroundColor: '#dc3545', color: '#fff' }
    addUserButton: {
        padding: '10px 20px',
        backgroundColor: '#28a745', // Green
        color: 'white',
        border: 'none',
        borderRadius: '4px',
        cursor: 'pointer',
        fontSize: '1em',
        fontWeight: '500',
        transition: 'background-color 0.2s ease'
        // addButtonHover: { backgroundColor: '#218838' }
    }
  };

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    setError('');
    const token = localStorage.getItem('token');
    if (!token) {
      setError('No autorizado. Por favor, inicie sesión.');
      setLoading(false);
      return;
    }
    try {
      const response = await fetch('/api/v1/admin/users', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
      });
      const data = await response.json();
      if (response.ok) {
        setUsers(Array.isArray(data) ? data : (data.users || [])); // Asegurar que data.users es un array
      } else {
        setError(data.message || data.error || `Error: ${response.status}`);
        setUsers([]); // Limpiar usuarios en caso de error
      }
    } catch (err) {
      console.error('Error fetching users:', err);
      setError('Error de conexión al obtener usuarios.');
      setUsers([]); // Limpiar usuarios en caso de error
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleOpenModal = (userToEdit = null) => { 
    setEditingUser(userToEdit); 
    setError(''); 
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingUser(null); // Siempre limpiar editingUser al cerrar el modal
  };

  const handleCreateUser = async (userData) => {
    setError(''); // Clear previous errors
    console.log('Attempting to create user with payload:', userData);
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/v1/admin/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(userData),
      });

      if (!response.ok) {
        let errorMsg = `Error ${response.status}: ${response.statusText}`;
        console.log('Raw error response status:', response.status, 'statusText:', response.statusText);
        try {
          // Try to clone the response before reading it as JSON, so it can be read again if needed
          const errorData = await response.json();
          console.log('Parsed error data from backend:', errorData);
          if (errorData) {
            errorMsg = errorData.details ? 
              `Error: ${errorData.error}. Detalles: ${errorData.details}` :
              errorData.error || JSON.stringify(errorData);
          }
        } catch (jsonError) {
          console.error('Failed to parse error response as JSON:', jsonError);
          // If response.json() fails, try to get text for more insight
          try {
            const textError = await response.text();
            console.error('Error response as text:', textError);
            errorMsg = textError || errorMsg; // Use text error if available
          } catch (textParseError) {
            console.error('Failed to parse error response as text:', textParseError);
          }
        }
        console.log('Throwing error with message:', errorMsg);
        throw new Error(errorMsg);
      }

      // const newUser = await response.json(); // response already consumed if !response.ok and .json() was called
      // To get newUser, we'd need to handle the response.json() call more carefully if we also parse error JSON
      // For now, just refetching is safer.
      fetchUsers(); // Refetch all users to get the latest list including the new one
      handleCloseModal();
    } catch (err) {
      console.error('Error creating user (in final catch):', err);
      setError(err.message || 'Ocurrió un error al crear el usuario.');
    }
  };

  const handleUpdateUser = async (userId, userData) => {
    setError('');
    console.log(`Attempting to update user ${userId} with payload:`, userData);

    // No enviar la contraseña si está vacía (para no sobrescribir con una vacía)
    const payload = { ...userData };
    if (!payload.password) {
      delete payload.password;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/admin/users/${userId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        let errorMsg = `Error ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.json();
          errorMsg = errorData.details ? 
            `Error: ${errorData.error}. Detalles: ${errorData.details}` :
            errorData.error || JSON.stringify(errorData);
        } catch (jsonError) { /* no body or not JSON */ }
        throw new Error(errorMsg);
      }

      fetchUsers(); // Recargar la lista de usuarios
      handleCloseModal();
    } catch (err) {
      console.error('Error updating user:', err);
      setError(err.message || 'Ocurrió un error al actualizar el usuario.');
      // Mantener el modal abierto para que el usuario vea el error y corrija
    }
  };

  const handleDeleteUser = async (userId) => {
    if (!window.confirm('¿Estás seguro de que deseas eliminar este usuario? Esta acción no se puede deshacer.')) {
      return;
    }
    setError('');
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/v1/admin/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        let errorMsg = `Error ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.json();
          errorMsg = errorData.details ? 
            `Error: ${errorData.error}. Detalles: ${errorData.details}` :
            errorData.error || JSON.stringify(errorData);
        } catch (jsonError) {
          // No body or not JSON
        }
        throw new Error(errorMsg);
      }

      // alert('Usuario eliminado exitosamente'); // Opcional: usar un sistema de notificaciones más elegante
      fetchUsers(); // Recargar la lista de usuarios
    } catch (err) {
      console.error('Error deleting user:', err);
      setError(err.message || 'Ocurrió un error al eliminar el usuario.');
    }
  };

  if (loading && users.length === 0) return <p style={styles.loading}>Cargando usuarios...</p>;
  // Mostrar error solo si no hay usuarios y no está cargando, o si es un error específico del modal
  if (error && !isModalOpen && users.length === 0) return <p style={styles.error}>Error al cargar usuarios: {error}</p>;


  return (
    <div style={styles.container}>
      <div style={styles.headerContainer}>
        <h1 style={styles.title}>Gestión de Usuarios</h1>
        <button style={styles.addUserButton} onClick={() => handleOpenModal(null)}>{/* Llamar con null explícitamente para crear */}
          Añadir Nuevo Usuario
        </button>
      </div>
      
      {error && !isModalOpen && <p style={{...styles.error, textAlign: 'center', paddingBottom: '10px'}}>{error}</p>} {/* Mostrar errores generales aquí si no es del modal */}

      <UserFormModal 
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        onSubmit={editingUser ? (data) => handleUpdateUser(editingUser.id, data) : handleCreateUser}
        initialData={editingUser} // Pasar datos del usuario para edición
        availableRoles={AVAILABLE_ROLES}
        apiError={error} // Pass API error to the modal
      />

      <table style={styles.userTable}>
        <thead style={styles.tableHeader}>
          <tr>
            <th style={styles.tableCell}>ID</th>
            <th style={styles.tableCell}>Usuario</th>
            <th style={styles.tableCell}>Nombre Completo</th>
            <th style={styles.tableCell}>Rol</th>
            <th style={styles.tableCell}>Email</th>
            <th style={styles.tableCell}>Teléfono</th>
            <th style={styles.tableCell}>Cargo</th>
            <th style={styles.tableCell}>Acciones</th>
          </tr>
        </thead>
        <tbody>{/* Ensure no whitespace or text nodes here */}
          {users.length > 0 ? (
            users.map((user, index) => (
              // Aplicar estilo alterno de fila si se desea
              // const rowStyle = index % 2 === 0 ? styles.tableCell : {...styles.tableCell, backgroundColor: '#f8f9fa'};
              <tr key={user.id} style={styles.tableRow}>
                <td style={styles.tableCell}>{user.id}</td>
                <td style={styles.tableCell}>{user.username}</td>
                <td style={styles.tableCell}>{`${user.employee_details?.name || ''} ${user.employee_details?.last_name || ''}`.trim() || 'N/A'}</td> {/* Nombre completo */}
                <td style={styles.tableCell}>{user.role}</td>
                <td style={styles.tableCell}>{user.employee_details?.email || 'N/A'}</td>
                <td style={styles.tableCell}>{user.employee_details?.phone_number || 'N/A'}</td>
                <td style={styles.tableCell}>{user.employee_details?.position || 'N/A'}</td>
                <td style={styles.tableCell}>
                  <button 
                    style={{...styles.actionButton, ...styles.editButton}}
                    onClick={() => handleOpenModal(user)} /* Pasar usuario a editar */
                  >
                    Editar
                  </button>
                  <button 
                    style={{...styles.actionButton, ...styles.deleteButton}}
                    onClick={() => handleDeleteUser(user.id)} /* Conectado handleDeleteUser */
                  >
                    Eliminar
                  </button>
                </td>
              </tr>
            ))
          ) : (
            <tr>{/* This tr is for 'No hay usuarios' or 'Cargando' */}
              <td colSpan="8" style={{...styles.tableCell, textAlign: 'center' }}>{/* Changed colSpan to 8 to match headers */}
                {loading ? 'Cargando...' : (error && users.length === 0 ? '' : 'No hay usuarios para mostrar.')}
              </td>
            </tr>
          )}
        </tbody>{/* Ensure no whitespace or text nodes here */}
      </table>
    </div>
  );
}

export default AdminUsersPage;
