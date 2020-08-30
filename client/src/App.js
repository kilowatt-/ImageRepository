import React from 'react';
import './App.css';
import Route from "react-router-dom/es/Route";
import Home from "./components/Home/Home";
import Login from "./components/Login/Login.jsx";
import { Switch } from "react-router-dom";

function App() {
  return (
      <div>
          <Switch>
              <Route path="/" component={Home} exact />
              <Route path="/login" component={Login} />
              <Route component={() => (<div><h1>404 Not Found</h1></div>)} />
          </Switch>
      </div>

  );
}

export default App;
