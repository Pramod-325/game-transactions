import React, { useState, useEffect } from "react";
import axios from "axios";
import "./App.css"; 

const API_URL = import.meta.env.VITE_API_URL;

function App() {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(localStorage.getItem("token"));
  const [view, setView] = useState("login");

  useEffect(() => {
    if (token) {
      fetchUserData();
      setView("dashboard");
    }
  }, [token]);

  const fetchUserData = async () => {
    try {
      const res = await axios.get(`${API_URL}/balance`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setUser(res.data);
    } catch (err) {
      logout();
    }
  };

  const logout = () => {
    localStorage.removeItem("token");
    setToken(null);
    setUser(null);
    setView("login");
  };

  return (
    <div className="app-container">
      <header className="main-header">
        <h1>💎 Game Wallet</h1>
        {user && <button className="logout-btn" onClick={logout}>Logout</button>}
      </header>

      <main className="content-area">
        {view === "login" && <LoginForm setToken={setToken} setView={setView} />}
        {view === "signup" && <SignupForm setToken={setToken} setView={setView} />}
        {view === "dashboard" && user && (
          <Dashboard user={user} token={token} refreshData={fetchUserData} />
        )}
      </main>
    </div>
  );
}

const LoginForm = ({ setToken, setView }) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const handleLogin = async (e) => {
    e.preventDefault();
    try {
      const res = await axios.post(`${API_URL}/login`, { username, password });
      const newToken = res.data.token;
      localStorage.setItem("token", newToken);
      setToken(newToken);
    } catch (err) {
      alert("Login failed!");
    }
  };

  return (
    <div className="card">
      <h2>Welcome Back</h2>
      <form onSubmit={handleLogin}>
        <input placeholder="Username" onChange={(e) => setUsername(e.target.value)} required />
        <input type="password" placeholder="Password" onChange={(e) => setPassword(e.target.value)} required />
        <button type="submit" style={{width: '100%'}}>Enter Game</button>
      </form>
      <p>Need an account?</p>
      <button onClick={() => setView("signup")} className="link">Signup</button>
    </div>
  );
};

const SignupForm = ({ setToken, setView }) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [referral, setReferral] = useState("");

  const handleSignup = async (e) => {
    e.preventDefault();
    try {
      await axios.post(`${API_URL}/signup`, { username, password, referralCode: referral });
      const res = await axios.post(`${API_URL}/login`, { username, password });
      localStorage.setItem("token", res.data.token);
      setToken(res.data.token);
    } catch (err) {
      alert("Signup failed!");
    }
  };

  return (
    <div className="card">
      <h2>Create Character</h2>
      <form onSubmit={handleSignup}>
        <input placeholder="Username" onChange={(e) => setUsername(e.target.value)} required />
        <input type="password" placeholder="Password" onChange={(e) => setPassword(e.target.value)} required />
        <input placeholder="Referral Code (Optional)" onChange={(e) => setReferral(e.target.value)} />
        <button type="submit" style={{width: '100%'}}>Start Journey</button>
      </form>
      <p onClick={() => setView("login")} className="link">Have an account? Login</p>
    </div>
  );
};

const Dashboard = ({ user, token, refreshData }) => {
  const [loading, setLoading] = useState(false);

  const handleAction = async (endpoint, payload = {}) => {
    setLoading(true);
    try {
      await axios.post(`${API_URL}/${endpoint}`, payload, {
        headers: { Authorization: `Bearer ${token}` }
      });
      refreshData();
    } catch (err) {
      alert(err.response?.data?.error || "Action failed.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="dashboard-grid">
      <div className="stats-bar">
        <div className="stat-box">
          <h3>Player Info</h3>
          <p><strong>{user.username}</strong></p>
          <small>Code: {user.referralCode}</small>
        </div>
        <div className="stat-box wallet">
          <h3>Wallet Balance</h3>
          <h2 style={{color: '#646cff'}}>💎 {user.balance}</h2>
        </div>
      </div>

      <div className="card">
        <h3>💰 Shop</h3>
        <div className="shop-item">
          <span>Gold Coin (10 💎)</span>
          <button disabled={loading} onClick={() => handleAction('purchase', {item: 'gold_coin'})}>Buy</button>
        </div>
        <div className="shop-item">
          <span>Treasure Box (50 💎)</span>
          <button disabled={loading} onClick={() => handleAction('purchase', {item: 'treasure_box'})}>Buy</button>
        </div>
      </div>

      <div className="card">
        <h3>🎒 Inventory</h3>
        <div className="shop-item">
          <span>Gold Coins</span>
          <strong>{user.inventory?.goldCoins || 0}</strong>
        </div>
        <div className="shop-item">
          <span>Treasure Boxes</span>
          <strong>{user.inventory?.treasureBoxes || 0}</strong>
        </div>
      </div>
      
      <div className="card">
         <h3>Bank</h3>
         <button className="topup-btn" disabled={loading} onClick={() => handleAction('top-up', {amount: 100})}>
           + Add 100 Diamonds
         </button>
      </div>
    </div>
  );
};

export default App;