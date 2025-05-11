const API_BASE = "http://localhost:8080/api";

async function login() {
    const username = document.getElementById("username").value.trim();
    const password = document.getElementById("password").value;

    if (!username || !password) {
        alert("用户名或密码不能为空");
        return;
    }

    const res = await fetch(`${API_BASE}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password })
    });

    const data = await res.json();
    if (res.ok) {
        alert("登录成功");

        // 存储 token 和 role
        localStorage.setItem("token", data.token);
        localStorage.setItem("role", data.role);

        // 根据角色跳转
        if (data.role === "admin") {
            window.location.href = "/static/admin.html";
        } else {
            window.location.href = "/static/dashboard.html";
        }
    } else {
        alert(data.error || "登录失败");
    }
}

async function register() {
    const username = document.getElementById("reg_username").value.trim();
    const password = document.getElementById("reg_password").value;

    if (!username || !password) {
        alert("用户名或密码不能为空");
        return;
    }

    const res = await fetch(`${API_BASE}/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password })
    });

    const data = await res.json();
    if (res.ok) {
        alert("注册成功，请登录！");
    } else {
        alert(data.error || "注册失败");
    }
}
