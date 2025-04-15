// frontend/js/main.js

const API_BASE = "http://localhost:8080/api";

async function login() {
    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    const res = await fetch(`${API_BASE}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password })
    });

    const data = await res.json();
    if (res.ok) {
        alert("登录成功");
        // 将 token 存入 localStorage
        localStorage.setItem("token", data.token);
        // 登录成功后跳转到 dashboard 页面
        window.location.href = "dashboard.html";
    } else {
        alert(data.error);
    }
}

async function register() {
    const username = document.getElementById("reg_username").value;
    const password = document.getElementById("reg_password").value;

    const res = await fetch(`${API_BASE}/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password })
    });

    const data = await res.json();
    if (res.ok) {
        alert("注册成功，请登录！");
    } else {
        alert(data.error);
    }
}
