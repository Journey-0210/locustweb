// frontend/js/dashboard.js

const API_BASE = "http://localhost:8080/api";
const token = localStorage.getItem("token");

// 如果没有 token，则跳转回登录页面
if (!token) {
    alert("请先登录！");
    window.location.href = "index.html";
}

/**
 * 将 datetime-local 控件的值转换成 "YYYY-MM-DDTHH:MM" 格式，
 * 如果原始字符串长度超过16（比如包含秒数），则截取前16个字符
 */
function convertToLayout(datetimeStr) {
    // datetime-local 的标准返回格式一般为 "YYYY-MM-DDTHH:MM" 或 "YYYY-MM-DDTHH:MM:SS"
    if (datetimeStr.length >= 16) {
        return datetimeStr.slice(0, 16);  // 只取年月日和小时分钟
    }
    return datetimeStr;
}

async function submitTask() {
    let userId = document.getElementById("userId").value;

    // 将 userId 转换为整数
    userId = parseInt(userId);

    // 确保 userId 转换成功
    if (isNaN(userId)) {
        alert("User ID 必须是整数！");
        return;
    }

    const numUsers = parseInt(document.getElementById("numUsers").value);
    const rampUp = parseInt(document.getElementById("rampUp").value);
    const targetUrl = document.getElementById("targetUrl").value;

    // 获取前端 datetime-local 输入控件的值
    let startTimeRaw = document.getElementById("startTime").value;
    let endTimeRaw = document.getElementById("endTime").value;

    // 转换为 "YYYY-MM-DDTHH:MM" 格式
    const startTime = convertToLayout(startTimeRaw);
    const endTime = convertToLayout(endTimeRaw);

    // Debug：打印提交的数据
    const payload = {
        user_id: userId,  // 确保是数字
        num_users: numUsers,
        ramp_up: rampUp,
        target_url: targetUrl,
        start_time: startTime,
        end_time: endTime
    };
    console.log("提交的数据:", JSON.stringify(payload));

    // 使用 fetch 请求发送数据
    const res = await fetch(`${API_BASE}/submit`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify(payload)
    });
    const data = await res.json();
    if (res.ok) {
        alert("任务提交成功，等待审批");
    } else {
        alert("提交失败: " + data.error);
    }
}

function downloadReport(format) {
    const taskId = document.getElementById("reportTaskId").value;
    if (!taskId) {
        alert("请输入任务ID");
        return;
    }
    // 直接通过新窗口打开下载链接
    const url = `${API_BASE}/download_report?test_id=${taskId}&format=${format}`;
    window.open(url, "_blank");
}

function logout() {
    localStorage.removeItem("token");
    window.location.href = "index.html";
}
