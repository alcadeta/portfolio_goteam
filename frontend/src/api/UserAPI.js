import axios from 'axios';

const UserAPI = {
  login: (username, password) => (
    axios.post(`${process.env.REACT_APP_BACKEND_URL}/login/`, {
      username,
      password,
    })
  ),
  register: (username, password, passwordConfirmation) => (
    axios.post(`${process.env.REACT_APP_BACKEND_URL}/register/`, {
      username,
      password,
      password_confirmation: passwordConfirmation,
    })
  ),
  verifyToken: () => (
    axios.post(`${process.env.REACT_APP_BACKEND_URL}/verify-token/`, {
      username: sessionStorage.getItem('username'),
      token: sessionStorage.getItem('auth-token'),
    })
  ),
};

export default UserAPI;
