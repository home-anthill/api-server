import logo from './logo.svg';
import './App.css';
import React, { Component, useEffect } from 'react';
import {
  BrowserRouter as Router,
  Switch,
  Route, Link,
  useLocation, useHistory, useRouteMatch
} from 'react-router-dom';

class App extends Component {
  state = {loginURL: null};

  login = () => {
    window.location.href = this.state.loginURL;
  }

  componentDidMount() {
    fetch('http://localhost:8080/login')
      .then(response => response.json())
      .then(responseData => {
        console.log('responseData', responseData);
        if (responseData) {
          const loginURL = responseData.loginURL;
          this.setState({loginURL: loginURL});
        }
      });
  }

  render() {
    return (
      <div className="App">
        <Router>

          <header className="App-header">
            <img src={logo} className="App-logo" alt="logo"/>
            <button onClick={this.login}>Login</button>
            <div>{this.state.loginURL}</div>
            <Link to="/postlogin">Postlogin</Link>
            <Link to="/secret">Secret</Link>
            <p>
              Edit <code>src/App.js</code> and save to reload.
            </p>
            <a
              className="App-link"
              href="https://reactjs.org"
              target="_blank"
              rel="noopener noreferrer"
            >
              Learn React
            </a>

            <Switch>
              <Route path="/postlogin">
                <PostLogin/>
              </Route>
              <Route path="/Secret">
                <Secret/>
              </Route>
              <Route path="/">
                <No/>
              </Route>
            </Switch>
          </header>
        </Router>
      </div>
    );
  }
}

const No = () => {
  const location = useLocation();
  const history = useHistory();
  const match = useRouteMatch('write-the-url-you-want-to-match-here');

  useEffect(() => {
    console.log('useEffect');
  });

  return (
    <div>{location.pathname}</div>
  )
}


const Secret = () => {

  useEffect(() => {
    let token = localStorage.getItem('token');
    let headers = {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    };

    fetch('http://localhost:8080/api/homes', {headers})
      .then(response => response.json())
      .then(responseData => {
        console.log('responseData', responseData);
      });
  });

  return (
    <div>
      <h1>Secret</h1>
    </div>
  )
}

const PostLogin = () => {
  const location = useLocation();
  const history = useHistory();
  const match = useRouteMatch('write-the-url-you-want-to-match-here');

  useEffect(() => {
    const token = new URLSearchParams(location.search).get('token');
    localStorage.setItem('token', token);
    history.push("/secret");
  });

  return (
    <div>
      <h1>PostLogin</h1>
      <div>{location.pathname}</div>
    </div>
  )
}

export default App;
