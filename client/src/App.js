import React from 'react';
import './App.css';
import Route from "react-router-dom/es/Route";
import Home from "./components/Home/Home";
import Login from "./components/Login/Login.jsx";

function App() {
  return (
      <div>
          <Route path="/" component={Home} exact />
          <Route path="/login" component={Login} />
      </div>

  );
}

export default App;
