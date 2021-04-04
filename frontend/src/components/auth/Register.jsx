import React, { useState } from 'react';
import { Form, Button } from 'react-bootstrap';
import FormGroup from './FormGroup';
import logo from '../../assets/registerHeader.svg';

const Register = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirmation, setPasswordConfirmation] = useState('');
  const handleSubmit = () => 1; // TODO: implement
  return (
    <div id="Register">
      <Form className="Form" onSubmit={handleSubmit}>
        <div className="HeaderWrapper">
          <img className="Header" alt="logo" src={logo} />
        </div>

        <FormGroup label="username" value={username} setValue={setUsername} />
        <FormGroup label="password" value={password} setValue={setPassword} />
        <FormGroup
          label="password confirmation"
          value={passwordConfirmation}
          setValue={setPasswordConfirmation}
        />

        <div className="ButtonWrapper">
          <Button className="Button" value="GO!" type="submit">
            GO!
          </Button>
        </div>

        <div className="Login">
          <p>Already have an account?</p>
          <p>
            <a href="/login">Login here.</a>
          </p>
        </div>
      </Form>
    </div>
  );
};

export default Register;
