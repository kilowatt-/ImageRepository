import React from 'react';
import './App.css';
import Home from "./components/Home/Home";
import Login from "./components/Login/Login.jsx";
import { Switch, Route } from "react-router-dom";
import { UserContextProvider } from "./context/UserContext";
import Signup from "./components/Signup/Signup";

function App() {
  return (
      <div>
          <UserContextProvider>
              <Switch>
                  <Route path="/" component={Home} exact />
                  <Route path="/login" component={Login} />
                  <Route path="/signup" component={Signup} />
                  <Route component={() => (<div><h1>404 Not Found</h1></div>)} />
              </Switch>
          </UserContextProvider>
      </div>
  );
}

export default App;
