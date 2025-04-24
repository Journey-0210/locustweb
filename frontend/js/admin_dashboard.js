const API = "http://localhost:8080/admin";
const token = localStorage.getItem("token");

// 获取任务
async function loadTasks() {
    const res = await fetch(`${API}/tasks`, {
        headers: { "Authorization": token }
    });
    const { tasks } = await res.json();
    const tbody = document.querySelector("#taskTable tbody");
    tbody.innerHTML = "";
    tasks.forEach(t => {
        const row = `<tr>
        <td>${t.id}</td>
        <td>${t.user_id}</td>
        <td>${t.num_users}</td>
        <td>${t.target_url}</td>
        <td><button onclick="approve(${t.id})">通过</button></td>
      </tr>`;
        tbody.insertAdjacentHTML("beforeend", row);
    });
}
async function approve(id) {
    const fd = new FormData(); fd.append("id", id);
    const res = await fetch(`${API}/approve`, {
        method: "POST", body: fd,
        headers: { "Authorization": token }
    });
    const msg = await res.json();
    alert(msg.message);
    loadTasks();
}
loadTasks();
