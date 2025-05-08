import React, { useState, useEffect } from 'react';

function UserFormModal({ isOpen, onClose, onSubmit, initialData, availableRoles, apiError }) {
  const [formData, setFormData] = useState({});
  const [formError, setFormError] = useState('');

  useEffect(() => {
    const defaultEmployeeDetails = {
      name: '',
      last_name: '',
      email: '',
      phone_number: '',
      position: '',
    };

    if (initialData) {
      // Normalizar rol de initialData a PascalCase
      let normalizedRole = initialData.role;
      if (normalizedRole) {
        normalizedRole = normalizedRole.charAt(0).toUpperCase() + normalizedRole.slice(1).toLowerCase();
        if (normalizedRole !== 'Admin' && normalizedRole !== 'Employee') {
            // Si después de normalizar no es uno de los roles válidos, usar el primer rol disponible como default
            // o dejarlo como está si es un caso inesperado pero se quiere conservar.
            // Por seguridad, lo ajustamos a uno válido o al default.
            normalizedRole = availableRoles && availableRoles.length > 0 ? availableRoles[0] : 'Employee'; 
        }
      }

      setFormData({
        ...initialData,
        role: normalizedRole || (availableRoles && availableRoles.length > 0 ? availableRoles[0] : 'Employee'),
        // Asegurar que employee_details sea un objeto, incluso si initialData.employee_details es null o undefined
        employee_details: initialData.employee_details 
          ? { ...defaultEmployeeDetails, ...initialData.employee_details } 
          : defaultEmployeeDetails,
        password: '', // Limpiar la contraseña al cargar para edición; el usuario debe reingresarla si quiere cambiarla
      });
    } else {
      setFormData({
        username: '',
        password: '',
        role: availableRoles && availableRoles.length > 0 ? availableRoles[0] : 'Employee',
        employee_details: defaultEmployeeDetails,
      });
    }
    setFormError('');
  }, [initialData, isOpen, availableRoles]);

  // Effect to clear formError if a new apiError is received
  useEffect(() => {
    if (apiError) {
      setFormError(''); // Clear local form errors if there's a new API error
    }
  }, [apiError]);

  if (!isOpen) return null;

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleEmployeeDetailsChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      employee_details: {
        ...prev.employee_details,
        [name]: value,
      },
    }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    setFormError('');
    if (!formData.username) {
      setFormError('El nombre de usuario es obligatorio.');
      return;
    }
    // La contraseña solo es obligatoria para nuevos usuarios (cuando no hay initialData)
    if (!initialData && !formData.password) {
      setFormError('La contraseña es obligatoria para nuevos usuarios.');
      return;
    }
    // Si hay initialData (edición) y se ingresó una contraseña, validar su longitud
    if (initialData && formData.password && formData.password.length < 8) {
        setFormError('La nueva contraseña debe tener al menos 8 caracteres.');
        return;
    }

    if (formData.employee_details.email && !/\S+@\S+\.\S+/.test(formData.employee_details.email)) {
      setFormError('Por favor, introduce un email válido.');
      return;
    }

    // Normalizar rol a PascalCase antes de enviar
    let finalRole = formData.role;
    if (finalRole) {
        finalRole = finalRole.charAt(0).toUpperCase() + finalRole.slice(1).toLowerCase();
    }

    const finalPayload = { 
      username: formData.username,
      role: finalRole, 
    };

    // Ensure password is only included if provided (especially for new users or if changed for existing)
    if (formData.password) {
        finalPayload.password = formData.password;
    } else {
        // If it's an existing user and password is blank, don't send the password field
        if (initialData) {
            delete finalPayload.password;
        }
    }

    // Clean employee_details
    const cleanDetails = {};
    let hasEmployeeDetails = false;
    if (formData.employee_details) {
        for (const key in formData.employee_details) {
            // Check if the property belongs to the object and is not empty
            if (Object.prototype.hasOwnProperty.call(formData.employee_details, key) && 
                formData.employee_details[key] !== null && 
                formData.employee_details[key] !== '') {
                cleanDetails[key] = formData.employee_details[key];
                hasEmployeeDetails = true;
            }
        }
    }

    if (hasEmployeeDetails) {
        finalPayload.employee_details = cleanDetails;
    } else {
        // If no actual details were provided, explicitly set to null.
        // The backend's `omitempty` on AdminCreateUserDTO.EmployeeDetails handles `nil`.
        finalPayload.employee_details = null; 
    }

    onSubmit(finalPayload);
  };

  const styles = {
    modalOverlay: {
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
      backgroundColor: 'rgba(0, 0, 0, 0.6)',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      zIndex: 1050, 
      fontFamily: '"Segoe UI", Tahoma, Geneva, Verdana, sans-serif'
    },
    modalContent: {
      backgroundColor: '#ffffff',
      padding: '30px',
      borderRadius: '8px',
      boxShadow: '0 5px 15px rgba(0, 0, 0, 0.3)',
      width: '90%',
      maxWidth: '550px',
      maxHeight: '90vh',
      overflowY: 'auto',
      position: 'relative', 
    },
    modalHeader: {
      borderBottom: '1px solid #e9ecef',
      paddingBottom: '15px',
      marginBottom: '20px',
      fontSize: '1.5rem',
      fontWeight: 500,
      color: '#343a40'
    },
    inputGroup: { marginBottom: '20px' },
    label: { display: 'block', marginBottom: '8px', fontWeight: '500', color: '#495057' },
    input: { 
      width: '100%', 
      padding: '10px 12px', 
      borderRadius: '4px', 
      border: '1px solid #ced4da', 
      fontSize: '1rem',
      boxSizing: 'border-box', 
      transition: 'border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out',
    },
    select: { 
        width: '100%', 
        padding: '10px 12px', 
        borderRadius: '4px', 
        border: '1px solid #ced4da', 
        fontSize: '1rem',
        boxSizing: 'border-box',
        backgroundColor: '#fff' 
    },
    buttonContainer: {
        marginTop: '30px',
        paddingTop: '20px',
        borderTop: '1px solid #e9ecef',
        display: 'flex',
        justifyContent: 'flex-end' 
    },
    button: { 
      padding: '10px 20px', 
      marginRight: '10px', 
      cursor: 'pointer', 
      borderRadius: '4px',
      fontSize: '1rem',
      fontWeight: '500',
      border: 'none',
      transition: 'background-color 0.2s ease-in-out'
    },
    buttonPrimary: {
        backgroundColor: '#007bff',
        color: 'white',
    },
    buttonSecondary: {
        backgroundColor: '#6c757d',
        color: 'white',
        marginRight: 0 
    },
    error: { color: '#dc3545', marginBottom: '15px', fontSize: '0.9rem' },
    sectionTitle: {
        fontSize: '1.2rem',
        fontWeight: 500,
        color: '#495057',
        marginTop: '30px',
        marginBottom: '15px',
        borderBottom: '1px solid #dee2e6',
        paddingBottom: '10px'
    }
  };

  return (
    <div style={styles.modalOverlay}>
      <div style={styles.modalContent}>
        <h2 style={styles.modalHeader}>{initialData ? 'Editar Usuario' : 'Añadir Nuevo Usuario'}</h2>
        <form onSubmit={handleSubmit}>
          {formError && <p style={styles.error}>{formError}</p>}
          {apiError && <p style={styles.error}>Error del servidor: {apiError}</p>} 
          
          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="username">Username:</label>
            <input style={styles.input} type="text" id="username" name="username" value={formData.username || ''} onChange={handleChange} required />
          </div>

          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="password">Contraseña:</label>
            <input 
              type="password"
              id="password"
              name="password"
              value={formData.password || ''}
              onChange={handleChange}
              style={styles.input}
              placeholder={initialData ? "Dejar en blanco para no cambiar" : "Mínimo 8 caracteres"}
            />
          </div>

          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="role">Rol:</label>
            <select style={styles.select} id="role" name="role" value={formData.role || 'Employee'} onChange={handleChange}>
              {availableRoles && availableRoles.map(role => (
                <option key={role} value={role}>{role.charAt(0).toUpperCase() + role.slice(1).toLowerCase()}</option>
              ))}
            </select>
          </div>

          <h4 style={styles.sectionTitle}>Detalles del Empleado (Opcional):</h4>
          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="name">Nombre:</label>
            <input style={styles.input} type="text" id="name" name="name" value={formData.employee_details?.name || ''} onChange={handleEmployeeDetailsChange} />
          </div>
          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="last_name">Apellido:</label>
            <input style={styles.input} type="text" id="last_name" name="last_name" value={formData.employee_details?.last_name || ''} onChange={handleEmployeeDetailsChange} />
          </div>
          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="email">Email:</label>
            <input style={styles.input} type="email" id="email" name="email" value={formData.employee_details?.email || ''} onChange={handleEmployeeDetailsChange} />
          </div>
          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="phone_number">Teléfono:</label>
            <input style={styles.input} type="text" id="phone_number" name="phone_number" value={formData.employee_details?.phone_number || ''} onChange={handleEmployeeDetailsChange} />
          </div>
          <div style={styles.inputGroup}>
            <label style={styles.label} htmlFor="position">Posición:</label>
            <input style={styles.input} type="text" id="position" name="position" value={formData.employee_details?.position || ''} onChange={handleEmployeeDetailsChange} />
          </div>

          <div style={styles.buttonContainer}>
            <button type="submit" style={{ ...styles.button, ...styles.buttonPrimary }}>
              {initialData ? 'Guardar Cambios' : 'Crear Usuario'}
            </button>
            <button type="button" style={{ ...styles.button, ...styles.buttonSecondary }} onClick={onClose}>Cancelar</button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default UserFormModal;
