// frontend/js/admin_dashboard.js

const API = "http://localhost:8080/admin";
const token = localStorage.getItem("token");

// 如果未登录或非管理员，直接跳回登录页
if (!token) {
    alert("请先登录");
    location.href = "/";
}

document.addEventListener("DOMContentLoaded", () => {
    const logoutBtn    = document.getElementById("logout");
    const statusSelect = document.getElementById("statusFilter");

    // 退出登录
    logoutBtn.onclick = () => {
        localStorage.removeItem("token");
        location.href = "/";
    };

    // 选择框变化时重新加载
    statusSelect.onchange = loadTasks;

    // 首次加载
    loadTasks();
});

async function loadTasks() {
    const status = document.getElementById("statusFilter").value;
    const url = `${API}/tasks?status=${status}`;
    const res = await fetch(url, {
        headers: { "Authorization": "Bearer " + token }
    });

    if (res.status === 403) {
        alert("你不是管理员或登录已过期");
        return;
    }
    if (!res.ok) {
        console.error("加载任务失败:", await res.text());
        return;
    }

    const { tasks } = await res.json();
    const tbody = document.querySelector("#taskTable tbody");
    tbody.innerHTML = "";

    tasks.forEach(task => {
        // 计算压测时长（秒）
        const duration = Math.floor(
            (new Date(task.end_time) - new Date(task.start_time)) / 1000
        );
        // 格式化时间
        const fmtStart = new Date(task.start_time).toLocaleString();
        const fmtEnd   = new Date(task.end_time).toLocaleString();

        tbody.insertAdjacentHTML("beforeend", `
      <tr>
        <td>${task.username}</td>
        <td>${task.num_users}</td>
        <td>${task.ramp_up}</td>
        <td>${duration}</td>
        <td>${task.target_url}</td>
        <td>${fmtStart}</td>
        <td>${fmtEnd}</td>
        <td>${task.status}</td>
        <td>
          ${task.status === "pending"
            ? `<button onclick="approveTask(${task.id})">通过</button>
               <button onclick="rejectTask(${task.id})">拒绝</button>`
            : ""
        }
        </td>
      </tr>
    `);
    });
}

// 审批通过
window.approveTask = async id => {
    const form = new URLSearchParams();
    form.append("id", id);
    const res = await fetch(`${API}/approve`, {
        method: "POST",
        headers: { "Authorization": "Bearer " + token },
        body: form
    });
    if (!res.ok) {
        alert("审批失败: " + await res.text());
        return;
    }
    alert("已通过任务ID: " + id);
    loadTasks();
};

// 审批拒绝
window.rejectTask = async id => {
    const form = new URLSearchParams();
    form.append("id", id);
    const res = await fetch(`${API}/reject`, {
        method: "POST",
        headers: { "Authorization": "Bearer " + token },
        body: form
    });
    if (!res.ok) {
        alert("拒绝失败: " + await res.text());
        return;
    }
    alert("已拒绝任务ID: " + id);
    loadTasks();
};
