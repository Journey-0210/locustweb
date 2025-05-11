// frontend/js/dashboard.js

const API_BASE = "http://localhost:8080/api";
const token = localStorage.getItem("token");

// 如果没有 token，则跳回登录页
if (!token) {
    alert("请先登录！");
    window.location.href = "index.html";
}

/**
 * 将 datetime-local 控件的值转换为 RFC3339 格式
 * 如 "2025-04-24T15:30" => "2025-04-24T15:30:00Z"
 */
function toRFC3339(datetimeLocal) {
    return datetimeLocal + ":00Z";
}

// 提交压测任务
async function submitTask() {
    const numUsers  = parseInt(document.getElementById("numUsers").value);
    const rampUp    = parseInt(document.getElementById("rampUp").value);
    const targetUrl = document.getElementById("targetUrl").value;
    const startTime = toRFC3339(document.getElementById("startTime").value);
    const endTime   = toRFC3339(document.getElementById("endTime").value);

    const payload = {
        num_users:  numUsers,
        ramp_up:    rampUp,
        target_url: targetUrl,
        start_time: startTime,
        end_time:   endTime
    };
    console.log("提交的数据:", payload);

    const res = await fetch(`${API_BASE}/submit`, {
        method: "POST",
        headers: {
            "Content-Type":  "application/json",
            "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify(payload)
    });
    const data = await res.json();
    if (res.ok) {
        alert("任务提交成功，等待审批");
        loadMyTasks();
    } else {
        alert("提交失败: " + (data.error || JSON.stringify(data)));
    }
}

// 下载报告
function downloadReport(format) {
    const taskId = document.getElementById("reportTaskId").value;
    if (!taskId) {
        alert("请输入任务ID");
        return;
    }
    const url = `${API_BASE}/download_report?test_id=${taskId}&format=${format}`;
    window.open(url, "_blank");
}

// 退出登录
function logout() {
    localStorage.removeItem("token");
    window.location.href = "index.html";
}

// 加载“我的任务列表”
async function loadMyTasks() {
    const res = await fetch(`${API_BASE}/tasks`, {
        headers: {
            "Content-Type":  "application/json",
            "Authorization": `Bearer ${token}`
        }
    });
    if (!res.ok) {
        console.error("加载我的任务失败:", await res.text());
        return;
    }
    const { tasks } = await res.json();
    const tbody = document.getElementById("myTasksBody");
    tbody.innerHTML = "";

    tasks.forEach(t => {
        const duration = Math.floor((new Date(t.end_time) - new Date(t.start_time)) / 1000);
        const fmtStart = new Date(t.start_time).toLocaleString();
        const fmtEnd   = new Date(t.end_time).toLocaleString();

        let reportLinks = "";
        if (t.status === "completed") {
            reportLinks = `
                <a href="${API_BASE}/download_report?test_id=${t.id}&format=csv" target="_blank">CSV</a>
                <a href="${API_BASE}/download_report?test_id=${t.id}&format=pdf" target="_blank">PDF</a>
            `;
        }

        const tr = document.createElement("tr");
        tr.innerHTML = `
            <td>${t.id}</td>
            <td>${t.num_users}</td>
            <td>${t.ramp_up}</td>
            <td>${t.target_url}</td>
            <td>${fmtStart}</td>
            <td>${fmtEnd}</td>
            <td>${duration}</td>
            <td>${t.status}</td>
            <td>${reportLinks}</td>
        `;
        tbody.appendChild(tr);
    });
}

// 页面加载完成后绑定事件
document.addEventListener("DOMContentLoaded", () => {
    document.getElementById("submitBtn").onclick      = submitTask;
    document.getElementById("downloadCsvBtn").onclick = () => downloadReport("csv");
    document.getElementById("downloadPdfBtn").onclick = () => downloadReport("pdf");
    document.getElementById("logoutBtn").onclick      = logout;

    loadMyTasks();
});
